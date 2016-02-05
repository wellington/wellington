package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wellington/wellington"
)

func resetFlags() {
	proj = ""
	includes = nil
	font = ""
	dir = ""
	gen = ""
	style = ""
	comments = false
	cpuprofile = ""
	buildDir = ""
	httpPath = ""
	timeB = false
	config = ""
	debug = false
	cachebust = ""
	relativeAssets = false
	paths = nil
}

func TestHTTP(t *testing.T) {
	resetFlags()

	os.Args = []string{
		os.Args[0],
		"serve",
	}

	req, err := http.NewRequest("POST", "http://localhost:12345",
		bytes.NewBufferString(`div { p { color: red; } }`))

	if err != nil {
		t.Fatal(err)
	}

	go main()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Body == nil {
		t.Fatal("no response")
	}
	bs, _ := ioutil.ReadAll(resp.Body)

	var r wellington.Response
	err = json.Unmarshal(bs, &r)
	if err != nil {
		t.Fatal(err)
	}

	e := "div p {\n  color: red; }\n"
	if e != r.Contents {
		t.Errorf("got:\n%s\nwanted:\n%s", r.Contents, e)
	}
	// Shutdown HTTP server
	lis.Close()
}

func TestStdin_import(t *testing.T) {
	resetFlags()

	oldOut := os.Stdout
	defer func() {
		os.Stdout = oldOut
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	includeDir := filepath.Join(pwd, "..", "test", "sass")
	os.Args = []string{
		os.Args[0],
		"-p", includeDir,
		"compile", "../test/sass/import.scss",
	}

	main()

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()

	out := <-outC
	out = strings.Replace(out, includeDir, "", 1)
	e := `div {
  background: #00FF00;
  font-size: 10pt; }
`

	if !bytes.Contains([]byte(out), []byte(e)) {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestStdin_sprite(t *testing.T) {
	resetFlags()
	os.RemoveAll("../test/img/build")

	oldStd := os.Stdin
	oldOut := os.Stdout
	defer func() {
		os.Stdin = oldStd
		os.Stdout = oldOut
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Args = []string{
		os.Args[0],
		"--dir", "../test/img",
		"--gen", "../test/img/build",
		"compile",
	}

	var err error
	os.Stdin, err = os.Open("../test/sass/sprite.scss")
	if err != nil {
		t.Fatal(err)
	}
	main()

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()

	out := <-outC

	e := `div {
  height: 139px;
  width: 96px;
  background: url("../test/img/build/f0a220.png") 0px 0px; }
`

	if !bytes.Contains([]byte(out), []byte(e)) {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestFile(t *testing.T) {
	// TODO: Tests for file importing here
}

func TestFile_comprehensive(t *testing.T) {
	resetFlags()
	os.RemoveAll("../test/img/build")

	oldOut := os.Stdout
	defer func() {
		os.Stdout = oldOut
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Args = []string{
		os.Args[0],
		"--dir", "../test/img",
		"--gen", "../test/img/build",
		"compile", "../test/comprehensive/compreh.scss",
	}
	main()

	outC := make(chan bytes.Buffer)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf
	}()

	w.Close()

	out := <-outC

	e, err := ioutil.ReadFile("../test/comprehensive/expected.css")
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(out.Bytes(), e) != 0 {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), string(e))
	}

}

func TestWatch_comprehensive(t *testing.T) {
	resetFlags()

	os.RemoveAll("../test/build/testwatch")

	os.Args = []string{
		os.Args[0],
		"--dir", "../test/img",
		"-b", "../test/build/testwatch",
		"--gen", "../test/build/testwatch/img",
		"watch", "../test/comprehensive/compreh.scss",
	}
	main()
	_, err := os.Stat("../test/build/testwatch/compreh.css")
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat("../test/build/testwatch/img/5905b8.png")
	if err != nil {
		t.Fatal(err)
	}
}
