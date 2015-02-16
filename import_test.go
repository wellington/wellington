package wellington

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
	path, res, err := p.ImportPath(dir, file)

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
	path, res, err = p.ImportPath(dir, file)

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
	_, res, err := p.ImportPath(dir, file)
	if res != "" {
		t.Errorf("Result from import on missing file: %s", file)
	}

	rel := strings.Replace(err.Error(), os.Getenv("PWD"), "", 1)
	if e := "Could not import: notafile\nTried:\n" +
		"./\n"; rel != e {
		t.Errorf("Error message invalid\nexpected: %s\nwas: %s",
			e, err.Error())
	}
}

func TestImportSass(t *testing.T) {
	p := NewParser()
	dir, file := "test/whitespace", "one"

	_, res, err := p.ImportPath(dir, file)
	if err != nil {
		t.Fatal(err)
	}
	if res != "" {

	}

	t.Error(res)

	// Importer
	//dir, file := "test/whitespace", "import"
}
