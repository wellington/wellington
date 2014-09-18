package sprite_sass_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestImportPath(t *testing.T) {
	p := NewParser()
	dir, file := "test", "var"
	contents, _ := ioutil.ReadFile("test/_var.scss")
	path, res, err := p.ImportPath(dir, file)

	if err != nil {
		t.Errorf("Error accessing file: %s", file)
	}

	if res != string(contents) {
		t.Errorf("Contents did not match expected:%s\nwas:%s",
			string(contents), res)
	}

	rel := strings.Replace(path, os.Getenv("PWD"), "", 1)
	if e := "/test"; e != rel {
		t.Errorf("Invalid path expected:%s\nwas:%s", e, rel)
	}

	p.Includes = []string{"test"}
	dir, file = "", "var"
	path, res, err = p.ImportPath(dir, file)

	if err != nil {
		t.Errorf("Error accessing file: %s", file)
	}

	if res != string(contents) {
		t.Errorf("Contents did not match expected:%s\nwas:%s",
			string(contents), res)
	}
}

func TestMissingImport(t *testing.T) {
	p := NewParser()
	dir, file := "test", "notafile"
	_, res, err := p.ImportPath(dir, file)
	if res != "" {
		t.Errorf("Result from import on missing file: %s", file)
	}
	if e := "Could not import: notafile\nTried:\n" +
		"/Users/drew/go/src/github.com/drewwells/" +
		"sprite_sass/test/_notafile.scss\n"; err.Error() != e {
		t.Errorf("Error message invalid expected:%s\nwas:%s", e, err.Error())
	}
}
