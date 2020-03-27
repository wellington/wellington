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
	"sync"
	"testing"

	"github.com/wellington/wellington"
)

// Sometimes circleci detects races in these tests. This may prevent it
var wtCmdMu sync.RWMutex

func init() {
	s := new(string)
	wtCmdMu.Lock()
	wtCmd.PersistentFlags().StringVarP(s, "test", "t", "", "dummy for testing")
	wtCmdMu.Unlock()

}

func resetFlags() {
	wtCmdMu.Lock()
	defer wtCmdMu.Unlock()
	wtCmd.ResetFlags()
}

func testMain() {
	wtCmdMu.Lock()
	AddCommands()
	root()
	wtCmdMu.Unlock()

	wtCmdMu.RLock()
	wtCmd.Execute()
	wtCmdMu.RUnlock()
}

func TestHTTP(t *testing.T) {
	wtCmd.SetArgs([]string{
		"--comment",
		"serve",
	})

	// No way to shut this down
	go func() {
		testMain()
	}()

	req, err := http.NewRequest("POST", "http://localhost:12345",
		bytes.NewBufferString(`div { p { color: red; } }`))

	if err != nil {
		t.Fatal(err)
	}

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

	e := "/* line 1, stdin */\ndiv p {\n  color: red; }\n"
	if e != r.Contents {
		t.Errorf("got:\n%s\nwanted:\n%s", r.Contents, e)
	}
	// Shutdown HTTP server
	lis.Close()
}

func TestStdin_import(t *testing.T) {
	resetFlags()

	oldOut := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	includeDir := filepath.Join(pwd, "..", "test", "sass")
	wtCmd.SetArgs([]string{
		// "-p", includeDir,
		"compile", "../test/sass/import.scss"})
	testMain()

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = oldOut

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

	oldStd := os.Stdin
	var oldOut *os.File
	oldOut = os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	wtCmd.SetArgs([]string{
		"--dir", "../test/img",
		"--build", "../test/build",
		"--gen", "../test/build/img",
		"compile"})

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
	os.Stdin = oldStd
	os.Stdout = oldOut

	out := <-outC

	e := `div {
  height: 139px;
  width: 96px;
  background: url("img/13dd5c.png") 0px 0px; }
`

	if !bytes.Contains([]byte(out), []byte(e)) {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestFile(t *testing.T) {
	// TODO: Tests for file importing here
}

func TestFile_comprehensive(t *testing.T) {
	os.RemoveAll("../test/build/testfile")
	resetFlags()

	wtCmd.SetArgs([]string{
		"--dir", "../test/img",
		"--build", "../test/build/testfile",
		"--gen", "../test/build/testfile/img",
		"--comment=false",
		"compile", "../test/comprehensive/compreh.scss"})
	main()

	e := `div {
  background: url("img/46339f.png") 0px -139px; }

div {
  background-file: "../img/*.png0, 140";
  background-position: 0px, -139px; }

div {
  background: url('../../img/IgotImage.png');
  height: 140px;
  width: 96px; }

div.inline {
  background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAAAA1BMVEX/TQBcNTh/AAAAAXRSTlMz/za5cAAAAA5JREFUeJxiYAAEAAD//wACAAFLuymfAAAAAElFTkSuQmCC"); }
`

	o, err := ioutil.ReadFile("../test/build/testfile/compreh.css")
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(o, []byte(e)) != 0 {
		t.Errorf("got:\n%s\nwanted:\n%s", string(o), string(e))
	}

}

func TestWatch_comprehensive(t *testing.T) {
	os.RemoveAll("../test/build/testwatch")
	resetFlags()

	wtCmd.SetArgs([]string{
		"--dir", "../test/img",
		"-b", "../test/build/testwatch",
		"--gen", "../test/build/testwatch/img",
		"--comment=false",
		"watch", "../test/comprehensive/compreh.scss",
	})
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
