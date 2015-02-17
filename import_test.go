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
		t.Errorf("Contents did not match expected:\n%s\nwas:\n%s",
			string(contents), res)
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
	dir, file := "test/whitespace", "two"

	_, res, err := p.ImportPath(dir, file)
	if err != nil {
		t.Fatal(err)
	}
	if res != "" {

	}

	e := `nav {
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
`

	if e != res {
		t.Errorf("got:\n%s\nwanted:\n%s", res, e)
	}

	// Importer
	//dir, file := "test/whitespace", "import"
}
