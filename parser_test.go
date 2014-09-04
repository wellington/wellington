package sprite_sass_test

import (
	"io/ioutil"
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestParser(t *testing.T) {
	p := Parser{}
	output := p.Start("test/_var.scss")

	file, _ := ioutil.ReadFile("test/var.css")
	if string(file) != output {
		t.Errorf("File output did not match, was:\n%s\nexpected:\n%s", output, string(file))
	}
}

func TestImporter(t *testing.T) {
	return
	p := Parser{}
	output := p.Start("test/import.scss")
	t.Error(output)

}
