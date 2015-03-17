package wellington

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImport_path(t *testing.T) {
	p := NewParser()
	dir, file := "test/sass", "var"
	contents, _ := ioutil.ReadFile("test/sass/_var.scss")
	path, res, err := p.ImportPath(dir, file)

	if err != nil {
		t.Errorf("Error accessing file: %s", file)
	}

	if !bytes.Equal(res, contents) {
		t.Errorf("Contents did not match expected:\n%s\nwas:\n%s",
			string(contents), string(res))
	}

	rel := strings.Replace(path, os.Getenv("PWD"), "", 1)
	if e := "/test/sass"; e != rel {
		t.Errorf("Invalid path expected:\n%s\nwas:\n%s", e, rel)
	}

	p.Includes = []string{"test"}
	// Is this how it should work?
	// Or should dir be appended to Includes
	dir, file = "test/sass", "var"
	path, res, err = p.ImportPath(dir, file)

	if err != nil {
		t.Errorf("Error accessing file: %s", file)
		t.Error(err)
	}

	if !bytes.Equal(res, contents) {
		t.Errorf("Contents did not match expected:%s\nwas:%s",
			string(contents), string(res))
	}
}

func TestImport_includes(t *testing.T) {
	p := NewParser()
	p.Includes = []string{filepath.Join(os.Getenv("PWD"), "test", "includes")}
	p.SassDir = filepath.Join(os.Getenv("PWD"), "test")
	_, _, err := p.ImportPath("includes", "includea")
	if err != nil {
		t.Fatal(err)
	}
}

func TestImport_missing(t *testing.T) {
	p := NewParser()
	p.SassDir = filepath.Join(os.Getenv("PWD"), "test")
	dir, file := "test", "notafile"
	_, res, err := p.ImportPath(dir, file)
	if res != nil {
		t.Errorf("Result from import on missing file: %s", file)
	}

	rel := strings.Replace(err.Error(), os.Getenv("PWD"), "", 1)
	if e := "Could not import: notafile\nTried:\n" +
		"    .\n    "; rel != e {
		t.Errorf("Error message invalid\nexpected: %s\nwas: %s",
			e, err.Error())
	}
}

func TestImport_sass(t *testing.T) {
	p := NewParser()
	dir, file := "test/whitespace", "two"

	_, res, err := p.ImportPath(dir, file)
	if err != nil {
		t.Fatal(err)
	}

	e := []byte(`nav {
  ul {
    margin: 0;
    padding: 0;
    list-style: none; }

  li {
    display: inline-block; }

  a {
    display: block;
    padding: 6px 12px;
    text-decoration: none; } }
`)

	if !bytes.Equal(e, res) {
		t.Errorf("got:\n%s\nwanted:\n%s", string(res), string(e))
	}

	// Importer
	//dir, file := "test/whitespace", "import"
}

func TestImport_bigfile(t *testing.T) {
	bs, err := ioutil.ReadFile("test/bigfile/_flex-box.scss")
	if err != nil {
		t.Fatal(err)
	}
	p := NewParser()
	_, contents, err := p.ImportPath("test/bigfile", "flex-box")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(bs, contents) {
		t.Errorf("got:\n%s\nwanted:\n%s", string(contents), string(bs))
	}
}
