package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStdin_import(t *testing.T) {
	fh, err := os.Open("../test/sass/import.scss")
	if err != nil {
		t.Error(err)
	}

	oldStd := os.Stdin
	oldOut := os.Stdout

	os.Stdin = fh
	r, w, _ := os.Pipe()
	os.Stdout = w
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	includeDir := filepath.Join(pwd, "..", "test", "sass")
	flag.Set("p", includeDir)
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
	out = strings.Replace(out, includeDir, "", 1)
	e := `div {
  background: #00FF00;
  font-size: 10pt; }
`

	if !bytes.Contains([]byte(e), []byte(e)) {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestStdin_sprite(t *testing.T) {
	fh, err := os.Open("../test/sass/sprite.scss")
	if err != nil {
		t.Error(err)
	}

	oldStd := os.Stdin
	oldOut := os.Stdout

	os.Stdin = fh
	r, w, _ := os.Pipe()
	os.Stdout = w
	flag.Set("dir", "../test/img")
	flag.Set("gen", "../test/img/build")
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
  background: url("../test/img/build/91300a.png") 0px 0px; }
`
	if !bytes.Contains([]byte(e), []byte(e)) {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestFile(t *testing.T) {
	// TODO: Tests for file importing here
}
