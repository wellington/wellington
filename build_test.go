package wellington

import (
	"bytes"
	"io"
	"os"
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
