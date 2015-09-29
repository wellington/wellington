package wellington

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/wellington/go-libsass"
)

func TestCompileStdin_imports(t *testing.T) {

	in := bytes.NewBufferString(`@import "compass";
@import "compass/utilities/sprite/base";

`)
	ctx := libsass.NewContext()
	InitializeContext(ctx)
	ctx.Imports.Init()

	var buf bytes.Buffer
	err := ctx.Compile(in, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()

	if e := ``; e != out {
		t.Fatalf("mismatch expected:\n%s\nwas:\n%s\n", e, out)
	}

}

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
	defer func() {
		w.Close()
		os.Stdout = oo
	}()
	os.Stdout = w
	err := LoadAndBuild("test/sass/error.scss", &BuildArgs{}, NewPartialMap())
	qs := fmt.Sprintf("%q", err.Error())

	e := `Invalid CSS after \"div {\": expected \"}\", was \"\"`
	if !strings.Contains(qs, e) {
		t.Fatalf("Error contains invalid text:\n%s", qs)
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
	r, w, _ := os.Pipe()

	bArgs := &BuildArgs{
		Includes: "test",
	}

	err := loadAndBuild("test/sass/file.scss", bArgs,
		NewPartialMap(), w, "")
	if err != nil {
		t.Fatal(err)
	}

	bs, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	e := `div {
  color: black; }
`
	if e != string(bs) {
		t.Errorf("got:\n%s\nwanted:\n%s", string(bs), e)
	}
}

func TestLoadAndBuild_comply(t *testing.T) {
	r, w, _ := os.Pipe()

	err := loadAndBuild("test/compass/top.scss",
		&BuildArgs{
			Includes: "test",
		},
		NewPartialMap(), w, "")

	bs, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	e := `one {
  color: red; }

two {
  color: blue; }

three {
  color: purple; }
`

	if e != string(bs) {
		t.Errorf("got:\n%s\nwanted:\n%s", string(bs), e)
	}
}
