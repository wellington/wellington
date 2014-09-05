package sprite_sass_test

import (
	"io/ioutil"
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestParserVar(t *testing.T) {
	p := Parser{}
	output := p.Start("test/_var.scss")

	file, _ := ioutil.ReadFile("test/var.css")
	if string(file) != output {
		t.Errorf("File output did not match, was:\n%s\nexpected:\n%s", output, string(file))
	}
}

func TestParserImporter(t *testing.T) {

	p := Parser{}
	output := p.Start("test/import.scss")

	file, _ := ioutil.ReadFile("test/import.css")
	if string(file) != output {
		t.Errorf("File output did not match, was:\n%s\nexpected:\n%s", output, string(file))
	}
}
