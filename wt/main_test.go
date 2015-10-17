package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	s := new(string)
	wtCmd.PersistentFlags().StringVarP(s, "test", "t", "", "dummy for testing")

}

func TestWatch(t *testing.T) {
	t.Skip()
	tdir, err := ioutil.TempDir("", "TestWatch")
	if err != nil {
		t.Fatal(err)
	}
	wtCmd.ResetFlags()

	watch = true
	wtCmd.SetArgs([]string{
		"--dir", tdir,
		"watch",
	})
	main()

}

func TestStdin_import(t *testing.T) {
	wtCmd.ResetFlags()

	oldOut := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	includeDir := filepath.Join(pwd, "..", "test", "sass")
	wtCmd.SetArgs([]string{
		"-p", includeDir,
		"compile", "../test/sass/import.scss"})
	main()

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
	wtCmd.ResetFlags()

	oldStd := os.Stdin
	oldOut := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	wtCmd.SetArgs([]string{
		"--dir", "../test/img",
		"--gen", "../test/img/build",
		"compile", "../test/sass/sprite.scss"})
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
  background: url("../test/img/build/f0a220.png") 0px 0px; }
`

	if !bytes.Contains([]byte(out), []byte(e)) {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestFile(t *testing.T) {
	// TODO: Tests for file importing here
}
