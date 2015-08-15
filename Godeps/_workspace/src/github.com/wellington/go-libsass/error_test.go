package context

import (
	"bytes"

	"testing"
)

type ErrorMap struct {
	line    int
	message string
}

func TestError_basic(t *testing.T) {
	in := bytes.NewBufferString(`div {
  @include invalid-function('');
}`)
	out := bytes.NewBuffer([]byte(""))
	ctx := Context{}
	err := ctx.Compile(in, out)
	if err == nil {
		t.Error("No error returned")
	}

	e := ErrorMap{2, "no mixin named invalid-function\n\nBacktrace:\n\tstdin:2"}

	if e.line != ctx.Errors.Line {
		t.Errorf("wanted: %d\ngot: %d", e.line, ctx.Errors.Line)
	}

	if e.message != ctx.Errors.Message {
		t.Errorf("wanted:\n%s\ngot:\n%s", e.message, ctx.Errors.Message)
	}

	if ctx.errorString != ctx.error() {
		t.Errorf("wanted: %s got: %s", ctx.errorString, ctx.error())
	}
}

func TestError_JSON(t *testing.T) {
	in := bytes.NewBufferString(`div {
  height: 10px;`)
	out := &bytes.Buffer{}
	ctx := Context{}
	ctx.Compile(in, out)

	e := `{
  "status": 1,
  "file": "stdin",
  "line": 2,
  "column": 16,
  "message": "Invalid CSS after \"...  height: 10px;\": expected \"}\", was \"\""
}`
	if ctx.libsassError != e {
		t.Fatalf("got: %s\nwanted: %s\n", ctx.libsassError, e)
	}
}

func TestError_unbound(t *testing.T) {
	in := bytes.NewBufferString(`div {
  background: map-get($sprite,139);
}`)
	out := bytes.NewBuffer([]byte(""))
	ctx := Context{}
	err := ctx.Compile(in, out)
	if err == nil {
		t.Error("No error returned")
	}

	e := ErrorMap{2, "unbound variable $sprite"}
	if e.line != ctx.Errors.Line {
		t.Errorf("wanted:\n%d\ngot:\n%d", e.line, ctx.Errors.Line)
	}

}

func TestError_function(t *testing.T) {
	in := bytes.NewBufferString(`// Empty line
@function uniqueFnName($file) {
  @return map-get($file,prop);
}
div {
  background: uniqueFnName(randfile);
}`)
	out := bytes.NewBuffer([]byte(""))
	ctx := Context{}
	err := ctx.Compile(in, out)
	if err == nil {
		t.Error("No error returned")
	}

	e := ErrorMap{3, "argument `$map` of `map-get($map, $key)`" + ` must be a map

Backtrace:
	stdin:3, in function ` + "`map-get`" + `
	stdin:3, in function ` + "`uniqueFnName`" + `
	stdin:6`}
	if e.line != ctx.Errors.Line {
		t.Errorf("wanted:\n%d\ngot:\n%d", e.line, ctx.Errors.Line)
	}

	if e.message != ctx.Errors.Message {
		t.Errorf("wanted:\n%s\ngot:\n%s", e.message, ctx.Errors.Message)
	}
}

func TestError_import(t *testing.T) {
	in := bytes.NewBufferString(`span {}
@import "fail";
`)

	out := bytes.NewBuffer([]byte(""))
	ctx := Context{}
	err := ctx.Compile(in, out)
	if err == nil {
		t.Error("No error returned")
	}
	e := ErrorMap{2, "file to import not found or unreadable: fail\nCurrent dir: "}
	if e.line != ctx.Errors.Line {
		t.Errorf("wanted:\n%d\ngot:\n%d", e.line, ctx.Errors.Line)
	}

	if e.message != ctx.Errors.Message {
		t.Errorf("wanted:\n%s\ngot:\n%s", e.message, ctx.Errors.Message)
	}

}

func TestError_processsass(t *testing.T) {
	in := []byte(`{
  "status": 1,
  "file": "stdin",
  "line": 3100,
  "column": 20,
  "message": "error in C function inline-image: format: .svg not supported\nBacktrace:\n\tstdin:3100, in function inline-image\n\tstdin:3100, in mixin printCSSImg\n\tstdin:3117"
}`)
	ctx := Context{}
	err := ctx.ProcessSassError(in)
	if err != nil {
		t.Error(err)
	}

	e := `Error > stdin:3100
error in C function inline-image: format: .svg not supported
Backtrace:
	stdin:3100, in function inline-image
	stdin:3100, in mixin printCSSImg
	stdin:3117`
	if e != ctx.error() {
		t.Errorf("got:\n%s\nwanted:\n%s", err.Error(), e)
	}
}

func TestError_invalid(t *testing.T) {
	ctx := Context{}
	err := ctx.ProcessSassError([]byte("/a"))

	if len(err.Error()) == 0 {
		t.Error("No error thrown on invalid sass json package")
	}
}

func TestError_line(t *testing.T) {
	ctx := Context{}
	ctx.errorString = "Error > stdin:1000"
	if e := 1000; e != ctx.ErrorLine() {
		t.Errorf("got: %d wanted: %d", ctx.ErrorLine(), e)
	}

}
