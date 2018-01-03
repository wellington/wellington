package libsass

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func ExampleCompiler_stdin() {

	src := bytes.NewBufferString(`div { p { color: red; } }`)

	comp, err := New(os.Stdout, src)
	if err != nil {
		log.Fatal(err)
	}
	err = comp.Run()
	if err != nil {
		log.Fatal(err)
	}

	// 	e := `div p {
	//   color: red; }
	// `
	// 	if e != dst.String() {
	// 		t.Errorf("got: %s wanted: %s", dst.String(), e)
	// 	}

	// Output:
	// div p {
	//   color: red; }
	//

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
