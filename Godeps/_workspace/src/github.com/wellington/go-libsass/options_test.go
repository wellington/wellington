package context

import (
	"bytes"
	"testing"
)

func TestOption_precision(t *testing.T) {
	t.Skip("precision does not work")
	in := bytes.NewBufferString(`a { height: 1.1111px; }`)

	var out bytes.Buffer
	ctx := Context{
		Precision: 3,
	}
	err := ctx.Compile(in, &out)
	if err != nil {
		t.Fatal(err)
	}

	e := `a {
  height: 1.111px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", out.String(), e)
	}

}

func TestOption_style(t *testing.T) {
	in := bytes.NewBufferString(`div { p { color: red; } }`)

	var out bytes.Buffer
	ctx := Context{
		OutputStyle: 0,
	}
	err := ctx.Compile(in, &out)
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
	err = ctx.Compile(in, &out)
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
	err = ctx.Compile(in, &out)
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
	err = ctx.Compile(in, &out)
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
	ctx := Context{
		Comments: false,
	}
	err := ctx.Compile(in, &out)
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
	err = ctx.Compile(in, &out)
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
	ctx := Context{
		IncludePaths: []string{"test/scss"},
	}
	err := ctx.Compile(in, &out)
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
