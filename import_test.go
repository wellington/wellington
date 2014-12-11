package sprite_sass

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestImportPath(t *testing.T) {
	p := NewParser()
	dir, file := "test/sass", "var"
	contents, _ := ioutil.ReadFile("test/sass/_var.scss")
	partialMap := NewPartialMap()
	path, res, err := p.ImportPath(dir, file, "var", partialMap)

	if err != nil {
		t.Errorf("Error accessing file: %s", file)
	}

	if res != string(contents) {
		t.Errorf("Contents did not match expected:%s\nwas:%s",
			string(contents), res)
	}

	rel := strings.Replace(path, os.Getenv("PWD"), "", 1)
	if e := "/test/sass"; e != rel {
		t.Errorf("Invalid path expected:%s\nwas:%s", e, rel)
	}

	p.Includes = []string{"test"}
	// Is this how it should work?
	// Or should dir be appended to Includes
	dir, file = "test/sass", "var"
	path, res, err = p.ImportPath(dir, file, "var", partialMap)

	if err != nil {
		t.Errorf("Error accessing file: %s", file)
		t.Error(err)
	}

	if res != string(contents) {
		t.Errorf("Contents did not match expected:%s\nwas:%s",
			string(contents), res)
	}
}

func TestMissingImport(t *testing.T) {
	p := NewParser()
	dir, file := "test", "notafile"
	var partialMap SafePartialMap
	_, res, err := p.ImportPath(dir, file, "var", &partialMap)
	if res != "" {
		t.Errorf("Result from import on missing file: %s", file)
	}

	rel := strings.Replace(err.Error(), os.Getenv("PWD"), "", 1)
	if e := "Could not import: notafile\nTried:\n" +
		"/test/_notafile.scss\n"; rel != e {
		t.Errorf("Error message invalid expected:%s\nwas:%s", e, err.Error())
	}
}
