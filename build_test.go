package wellington

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/wellington/go-libsass"
)

func init() {
	testch = make(chan struct{})
	close(testch)
}

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

func TestNewBuild(t *testing.T) {

	b := NewBuild([]string{"test/sass/error.scss"}, &BuildArgs{}, nil, false)
	if b == nil {
		t.Fatal("build is nil")
	}

	err := b.Build()
	if err != ErrPartialMap {
		t.Errorf("got: %s wanted: %s", err, ErrPartialMap)
	}
	b.Close()
}

func TestNewBuild_two(t *testing.T) {
	tdir, _ := ioutil.TempDir("", "testnewbuild_two")
	bb := NewBuild([]string{"test/sass/file.scss"},
		&BuildArgs{BuildDir: tdir}, NewPartialMap(), false)

	err := bb.Build()
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Error(err)
	}

}

func TestNewBuild_dir(t *testing.T) {
	tdir, _ := ioutil.TempDir("", "testnewbuild_two")
	bb := NewBuild([]string{"test/sass"},
		&BuildArgs{BuildDir: tdir}, NewPartialMap(), false)

	err := bb.Build()
	if err == nil {
		t.Fatal("expected error")
	}

}

func ExampleBuild() {
	err := LoadAndBuild("test/sass/file.scss", &BuildArgs{}, NewPartialMap())
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// div {
	//   color: black; }
}

func TestBuild_error(t *testing.T) {

	_, w, _ := os.Pipe()

	err := loadAndBuild("test/sass/error.scss", &BuildArgs{},
		NewPartialMap(), w, "")

	if err == nil {
		t.Fatal("no error thrown")
	}

	qs := fmt.Sprintf("%q", err.Error())

	e := `Invalid CSS after \"div {\": expected \"}\", was \"\"`
	if !strings.Contains(qs, e) {
		t.Fatalf("Error contains invalid text:\n%s", qs)
	}
}

func TestBuild_args(t *testing.T) {
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

func TestBuild_comply(t *testing.T) {
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

func TestUpdateFileOutputType(t *testing.T) {
	s := "file.scss"
	ren := updateFileOutputType(s)
	if e := "file.css"; e != ren {
		t.Errorf("got: %s wanted: %s", ren, e)
	}
}
