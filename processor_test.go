package sprite_sass_test

import (
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestProcessor(t *testing.T) {

	p := Processor{
		Ipath: "test/_var.scss",
		Opath: "test/var.css.out",
	}

	p.Run()
}
