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

func TestFromBuildArgs(t *testing.T) {
	_, err := FromBuildArgs(nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = FromBuildArgs(nil, nil, &BuildArgs{})
	if err != nil {
		t.Fatal(err)
	}

}

// InitializeContext sets up some data structures necessary
// to use wellington
func InitializeContext(ctx *libsass.Context) {
	ctx.Payload = newPayload()
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

func BenchmarkNewBuild(b *testing.B) {
	ins := []string{"test/sass/file.scss"}
	pmap := NewPartialMap()
	args := &BuildArgs{}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		bld := NewBuild(ins, args, pmap)
		err := bld.Run()
		if err != nil {
			b.Fatal(err)
		}
		bld.Close()
	}
}

func TestNewBuild(t *testing.T) {

	b := NewBuild([]string{"test/sass/error.scss"}, &BuildArgs{}, nil)
	if b == nil {
		t.Fatal("build is nil")
	}

	err := b.Run()
	if err != ErrPartialMap {
		t.Errorf("got: %s wanted: %s", err, ErrPartialMap)
	}
	b.Close()
}

func TestNewBuild_two(t *testing.T) {
	tdir, _ := ioutil.TempDir("", "testnewbuild_two")
	sdir := filepath.Join(tdir, "sass")
	bdir := filepath.Join(tdir, "build")
	err := os.MkdirAll(sdir, 0744)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Create(filepath.Join(sdir, "one.scss"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Create(filepath.Join(sdir, "two.scss"))
	if err != nil {
		t.Fatal(err)
	}
	ps := []string{sdir}
	bb := NewBuild(ps,
		&BuildArgs{
			paths:    ps,
			BuildDir: bdir,
		}, NewPartialMap())

	err = bb.Run()
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Error(err)
	}

	matches, err := filepath.Glob(filepath.Join(bdir, "*.css"))
	if err != nil {
		t.Fatal(err)
	}
	if e := 2; len(matches) != e {
		t.Errorf("got: %d wanted: %d", len(matches), e)
	}
}

func TestNewBuild_dir(t *testing.T) {
	tdir, _ := ioutil.TempDir("", "testnewbuild_two")
	ps := []string{"test/sass"}
	bb := NewBuild(
		ps,
		&BuildArgs{BuildDir: tdir, paths: ps},
		NewPartialMap())
	os.RemoveAll(filepath.Join(tdir, "*"))

	err := bb.Run()
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

	ps = []string{"test/subdir"}
	bb = NewBuild(ps,
		&BuildArgs{
			paths:    ps,
			BuildDir: tdir,
		},
		NewPartialMap())
	os.RemoveAll(filepath.Join(tdir, "test"))

	err = bb.Run()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(tdir, "main.css")
	matches, _ = filepath.Glob(path)
	if e := 1; len(matches) != e {
		t.Errorf("matches: % #v\n", matches)
		t.Fatalf("got: %d wanted: %d", len(matches), e)
	}

	path = filepath.Join(tdir, "sub", "*.css")
	matches, _ = filepath.Glob(path)
	if e := 1; len(matches) != e {
		t.Errorf("matches: % #v\n", matches)
		t.Errorf("got: %d wanted: %d", len(matches), e)
	}
}

func ExampleNewBuild() {
	ps := []string{"test/sass/file.scss"}
	b := NewBuild(ps,
		&BuildArgs{paths: ps}, NewPartialMap())

	err := b.Run()
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
	w.Close()
	e := `Invalid CSS after "div {": expected "}", was ""`
	if !strings.HasSuffix(err.Error(), e) {
		t.Fatalf("Error contains invalid text:\n%s", err)
	}
}

func TestBuild_args(t *testing.T) {
	r, w, _ := os.Pipe()

	bArgs := &BuildArgs{
		Includes: []string{"test"},
	}

	err := loadAndBuild("test/sass/file.scss", bArgs,
		NewPartialMap(), w, "")
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
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
			Includes: []string{"test"},
		},
		NewPartialMap(), w, "")
	w.Close()
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
