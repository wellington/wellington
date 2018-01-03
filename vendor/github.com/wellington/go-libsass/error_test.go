package libsass

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
	ctx := newContext()
	err := ctx.compile(out, in)
	if err == nil {
		t.Error("No error returned")
	}

	e := ErrorMap{2, "no mixin named invalid-function\n\nBacktrace:\n\tstdin:2"}

	if e.line != ctx.err.Line {
		t.Errorf("wanted: %d\ngot: %d", e.line, ctx.err.Line)
	}

	if e.message != ctx.err.Message {
		t.Errorf("wanted:\n%s\ngot:\n%s", e.message, ctx.err.Message)
	}

	if ctx.errorString != ctx.Error() {
		t.Errorf("wanted: %s got: %s", ctx.errorString, ctx.Error())
	}
}

func TestError_JSON(t *testing.T) {
	in := bytes.NewBufferString(`div {
  height: 10px;`)
	out := &bytes.Buffer{}
	ctx := newContext()
	err := ctx.compile(out, in)

	e := `Error > stdin:2
Invalid CSS after "  height: 10px;": expected "}", was ""
div {
  height: 10px;
`

	if err == nil {
		t.Fatal("no error thrown")
	}

	if e != err.Error() {
		t.Fatalf("got: %s\nwanted: %s", err, e)
	}
}

func TestError_unbound(t *testing.T) {
	in := bytes.NewBufferString(`div {
  background: map-get($sprite,139);
}`)
	out := bytes.NewBuffer([]byte(""))
	ctx := newContext()
	err := ctx.compile(out, in)
	if err == nil {
		t.Error("No error returned")
	}

	e := ErrorMap{2, "unbound variable $sprite"}
	if e.line != ctx.err.Line {
		t.Errorf("wanted:\n%d\ngot:\n%d", e.line, ctx.err.Line)
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
	ctx := newContext()
	err := ctx.compile(out, in)
	if err == nil {
		t.Error("No error returned")
	}

	e := ErrorMap{3, "argument `$map` of `map-get($map, $key)`" + ` must be a map

Backtrace:
	stdin:3, in function ` + "`map-get`" + `
	stdin:3, in function ` + "`uniqueFnName`" + `
	stdin:6`}
	if e.line != ctx.err.Line {
		t.Errorf("wanted:\n%d\ngot:\n%d", e.line, ctx.err.Line)
	}

	if e.message != ctx.err.Message {
		t.Errorf("wanted:\n%s\ngot:\n%s", e.message, ctx.err.Message)
	}
}

func TestError_import(t *testing.T) {
	in := bytes.NewBufferString(`span {}
@import "fail";
`)

	out := bytes.NewBuffer([]byte(""))
	ctx := newContext()
	err := ctx.compile(out, in)
	if err == nil {
		t.Error("No error returned")
	}
	e := ErrorMap{2, `File to import not found or unreadable: fail
Parent style sheet: stdin`}
	if e.line != ctx.err.Line {
		t.Errorf("wanted:\n%d\ngot:\n%d", e.line, ctx.err.Line)
	}

	if e.message != ctx.err.Message {
		t.Errorf("wanted:%s\ngot:%s", e.message, ctx.err.Message)
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
	ctx := newContext()
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
	if e != ctx.Error() {
		t.Errorf("got:\n%s\nwanted:\n%s", err.Error(), e)
	}
}

func TestError_invalid(t *testing.T) {
	ctx := newContext()
	err := ctx.ProcessSassError([]byte("/a"))

	if len(err.Error()) == 0 {
		t.Error("No error thrown on invalid sass json package")
	}
}

func TestError_line(t *testing.T) {
	ctx := newContext()
	ctx.errorString = "Error > stdin:1000"
	if e := 1000; e != ctx.ErrorLine() {
		t.Errorf("got: %d wanted: %d", ctx.ErrorLine(), e)
	}

}
