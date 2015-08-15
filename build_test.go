package wellington

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLoadAndBuild(t *testing.T) {
	oo := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	err := LoadAndBuild("test/sass/file.scss", &BuildArgs{}, NewPartialMap())
	if err != nil {
		t.Error(err)
	}
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = oo
	out := <-outC

	e := `div {
  color: black; }
Rebuilt: test/sass/file.scss
`
	if e != out {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestLandB_error(t *testing.T) {
	oo := os.Stdout
	var w *os.File
	defer w.Close()
	os.Stdout = w
	err := LoadAndBuild("test/sass/error.scss", &BuildArgs{}, NewPartialMap())
	qs := fmt.Sprintf("%q", err.Error())

	if !strings.HasPrefix(qs, `"\x1b[31mError >`) {
		t.Fatalf("Error is not colorized")
	}

	if !strings.HasSuffix(qs, `error.scss:1\nInvalid CSS after \"div {\\a\": expected \"}\", was \"\"\x1b[0m"`) {
		t.Fatalf("Error contains invalid text:\n%s", qs)
	}
	os.Stdout = oo
}

func TestLandB_updateFile(t *testing.T) {
	s := "file.scss"
	ren := updateFileOutputType(s)
	if e := "file.css"; e != ren {
		t.Errorf("got: %s wanted: %s", ren, e)
	}
}

func TestLoadAndBuild_args(t *testing.T) {
	oo := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	err := LoadAndBuild("test/sass/file.scss",
		&BuildArgs{
			BuildDir: "test/build",
			Includes: "test",
		},
		NewPartialMap(),
	)
	if err != nil {
		t.Error(err)
	}
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = oo
	out := <-outC

	e := `Rebuilt: test/sass/file.scss
`
	if e != out {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestLoadAndBuild_comply(t *testing.T) {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := LoadAndBuild("test/compass/top.scss",
		&BuildArgs{
			BuildDir: "test/build",
			Includes: "test",
		},
		NewPartialMap(),
	)
	if err != nil {
		t.Error(err)
	}
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = stdout
	out := <-outC

	e := `Rebuilt: test/compass/top.scss
`
	if e != out {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}
}
