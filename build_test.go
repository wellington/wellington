package wellington

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/wellington/go-libsass"
)

func init() {
	testch = make(chan struct{})
	close(testch)
	color.NoColor = true
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

	matches, err := filepath.Glob(filepath.Join(tdir, "test", "*.css"))
	if err != nil {
		t.Fatal(err)
	}
	if e := 0; len(matches) != e {
		t.Errorf("got: %d wanted: %d", len(matches), e)
	}

	matches, err = filepath.Glob(filepath.Join(tdir, "test", "sass", "*.css"))
	if err != nil {
		t.Fatal(err)
	}
	if e := 1; len(matches) != e {
		t.Errorf("got: %d wanted: %d", len(matches), e)
	}

}

func TestNewBuild_dir(t *testing.T) {
	tdir, _ := ioutil.TempDir("", "testnewbuild_two")
	bb := NewBuild([]string{"test/sass"},
		&BuildArgs{BuildDir: tdir}, NewPartialMap(), false)
	os.RemoveAll(filepath.Join(tdir, "*"))

	err := bb.Build()
	if err == nil {
		t.Fatal("expected error")
	}

	matches, err := filepath.Glob(filepath.Join(tdir, "test", "*.css"))
	if err != nil {
		t.Fatal(err)
	}
	if e := 0; len(matches) != e {
		t.Errorf("got: %d wanted: %d", len(matches), e)
	}

	bb = NewBuild([]string{"test/subdir"},
		&BuildArgs{BuildDir: tdir}, NewPartialMap(), false)
	os.RemoveAll(filepath.Join(tdir, "test"))

	err = bb.Build()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(tdir, "test", "subdir", "*.css")
	matches, err = filepath.Glob(path)
	if err != nil {
		t.Fatal(err)
	}
	if e := 1; len(matches) != e {
		t.Errorf("matches: % #v\n", matches)
		t.Errorf("got: %d wanted: %d", len(matches), e)
	}

	path = filepath.Join(tdir, "test", "subdir", "sub", "*.css")
	matches, err = filepath.Glob(path)
	if err != nil {
		t.Fatal(err)
	}
	if e := 1; len(matches) != e {
		t.Errorf("matches: % #v\n", matches)
		t.Errorf("got: %d wanted: %d", len(matches), e)
	}
}

func ExampleNewBuild() {
	b := NewBuild([]string{"test/sass/file.scss"},
		&BuildArgs{}, NewPartialMap(), false)

	err := b.Build()
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

	e := `Invalid CSS after "div {": expected "}", was ""`
	if !strings.HasSuffix(err.Error(), e) {
		t.Fatalf("Error contains invalid text:\n%s", err)
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
