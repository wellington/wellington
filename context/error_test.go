package context

import (
	"bytes"

	"testing"
)

type ErrorMap struct {
	line    int
	message string
}

func TestErrorBasic(t *testing.T) {
	in := bytes.NewBufferString(`div {
  @include invalid-function('');
}`)
	out := bytes.NewBuffer([]byte(""))
	ctx := Context{}
	err := ctx.Compile(in, out)
	if err == nil {
		t.Error("No error returned")
	}

	e := ErrorMap{2, "no mixin named invalid-function\nBacktrace:\n\tstdin:2"}

	if e.line != ctx.Errors.Line {
		t.Errorf("wanted: %d\ngot: %d", e.line, ctx.Errors.Line)
	}

	if e.message != ctx.Errors.Message {
		t.Errorf("wanted:\n%s\ngot:\n%s", e.message, ctx.Errors.Message)
	}

	if ctx.errorString != ctx.Error() {
		t.Errorf("wanted: %s got: %s", ctx.errorString, ctx.Error())
	}
}

func TestErrorUnbound(t *testing.T) {
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

func TestErrorFunction(t *testing.T) {
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

func TestErrorImport(t *testing.T) {
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

func TestProcessSassError(t *testing.T) {
	in := []byte(`{
  "status": 1,
  "file": "stdin",
  "line": 3100,
  "column": 20,
  "message": "error in C function inline-image: format: .svg not supported\nBacktrace:\n\tstdin:3100, in function inline-image\n\tstdin:3100, in mixin printCSSImg\n\tstdin:3117"
}`)
	ctx := Context{}
	bs, err := ctx.ProcessSassError(in)
	if err != nil {
		t.Error(err)
	}

	e := `Error > stdin:3100
error in C function inline-image: format: .svg not supported
Backtrace:
	stdin:3100, in function inline-image
	stdin:3100, in mixin printCSSImg
	stdin:3117`
	if e != bs {
		t.Errorf("got:\n%s\nwanted:\n%s", bs, e)
	}
}

func TestErrorWarn(t *testing.T) {
	// Disabled while new warn integration is built
	// in := bytes.NewBufferString(`
	// @warn "WARNING";`)
	// 	out := bytes.NewBuffer([]byte(""))
	// 	ctx := Context{}
	// 	err := ctx.Compile(in, out)
	// 	_ = err
}

func TestErrorInvalid(t *testing.T) {
	ctx := Context{}
	_, err := ctx.ProcessSassError([]byte("/a"))

	if len(err.Error()) == 0 {
		t.Error("No error thrown on invalid sass json package")
	}
}
