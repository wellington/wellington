package wellington

import (
	"os"
	"testing"
)

func TestSassToScss(t *testing.T) {

	p := NewParser()
	p.Includes = []string{"test/whitespace"}
	p.SassDir = os.Getenv("PWD")

	in := fileReader("test/whitespace/import.sass")

	bs, err := p.Start(in, "")
	if err != nil {
		t.Fatal(err)
	}

	e := `
@import "one";

body {
  font: 100% $font-stack;
  background-color: $primary-color; }
`

	if e != string(bs) {
		t.Errorf("got:\n%s\nwanted:\n%s", string(bs), e)
	}

	bs, err = p.Start(fileReader("test/whitespace/base.sass"), "")
	if err != nil {
		t.Fatal(err)
	}

	e = `
html,
body,
ul,
ol {
  margin:  0;
  padding: 0; }
`

	if e != string(bs) {
		t.Errorf("got:\n%s\nwanted:\n%s", string(bs), e)
	}
}
