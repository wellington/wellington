package libsass

import (
	"bytes"
	"testing"
)

func TestCompiler_stdin(t *testing.T) {
	var dst bytes.Buffer
	src := bytes.NewBufferString(`div { p { color: red; } }`)

	comp, err := New(&dst, src)
	if err != nil {
		t.Fatal(err)
	}
	err = comp.Run()
	if err != nil {
		t.Fatal(err)
	}

	e := `div p {
  color: red; }
`
	if e != dst.String() {
		t.Errorf("got: %s wanted: %s", dst.String(), e)
	}
}

func TestCompiler_path(t *testing.T) {
	var dst bytes.Buffer

	comp, err := New(&dst, nil, Path("test/scss/basic.scss"))
	if err != nil {
		t.Fatal(err)
	}
	err = comp.Run()
	if err != nil {
		t.Fatal(err)
	}

	e := `div p {
  color: red; }
`
	if e != dst.String() {
		t.Errorf("got: %s wanted: %s", dst.String(), e)
	}

	if e := 1; len(comp.Imports()) != e {
		t.Errorf("got: %d wanted: %d", len(comp.Imports()), e)
	}

}
