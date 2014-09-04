package sprite_sass_test

import (
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestParser(t *testing.T) {
	p := Parser{}
	output := p.Start("test/_var.scss")
	t.Errorf("\n%s", output)
}

func TestImporter(t *testing.T) {
	return
	p := Parser{}
	output := p.Start("test/import.scss")
	t.Error(output)

}
