package libsass

import (
	"bytes"
	"testing"
)

func TestOption_precision(t *testing.T) {

	in := bytes.NewBufferString(`a { height: (1/3)px; }`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.Precision = 3
	err := ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}

	e := `a {
  height: 0.333 px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", out.String(), e)
	}

	in = bytes.NewBufferString(`a { height: (1/3)px; }`)
	out.Reset()
	ctx = newContext()
	ctx.Precision = 6
	err = ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}

	e = `a {
  height: 0.333333 px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", out.String(), e)
	}

}

func TestOption_style(t *testing.T) {
	in := bytes.NewBufferString(`div { p { color: red; } }`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.OutputStyle = 0

	err := ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}

	e := `div p {
  color: red; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", out.String(), e)
	}

	in = bytes.NewBufferString(`div { p { color: red; } }`)
	out.Reset()
	ctx.OutputStyle = 1
	err = ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}

	e = `div p {
  color: red;
}
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", out.String(), e)
	}

	in = bytes.NewBufferString(`div { p { color: red; } }`)
	out.Reset()
	ctx.OutputStyle = 2
	err = ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}

	e = `div p { color: red; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", out.String(), e)
	}

	in = bytes.NewBufferString(`div { p { color: red; } }`)
	out.Reset()
	ctx.OutputStyle = 3
	err = ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}

	e = `div p{color:red}
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", out.String(), e)
	}

}

func TestOption_comment(t *testing.T) {
	in := bytes.NewBufferString(`div { p { color: red; } }`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.Comments = false

	err := ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}

	e := `div p {
  color: red; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", out.String(), e)
	}

	in = bytes.NewBufferString(`div { p { color: red; } }`)
	out.Reset()
	ctx.Comments = true
	err = ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}

	e = `/* line 1, stdin */
div p {
  color: red; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", out.String(), e)
	}

}

func TestOption_include(t *testing.T) {
	in := bytes.NewBufferString(`@import "include";`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.IncludePaths = []string{"test/scss"}
	err := ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}

	e := `a {
  color: blue; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", out.String(), e)
	}

}
