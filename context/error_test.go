package context

import (
	"bytes"
	"io"
	"log"
	"testing"
)

func setupCtx(f interface{}) (Context, string, error) {
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		BuildDir:     "test/build",
		ImageDir:     "test/img",
		GenImgDir:    "test/build/img",
		Out:          "",
		// Parser: Parser{
		// 	MainFile: "testname",
		// },
	}
	var (
		out bytes.Buffer
		err error
	)

	var reader io.Reader
	switch v := f.(type) {
	case io.Reader:
		reader = f.(io.Reader)
	case string:
		reader = fileReader(f.(string))
	default:
		log.Printf("Unhandled type: %T", v)
		return ctx, "", nil
	}

	err = ctx.Compile(reader, &out, "test/sass")
	if err != nil {
		// This will mask iport errors
		log.Print(err)
	}

	return ctx, out.String(), err
}

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
	err := ctx.Compile(in, out, "")
	if err == nil {
		t.Error("No error returned")
	}

	e := ErrorMap{2, "no mixin named invalid-function\nBacktrace:\n\tstdin:2"}

	if e.line != ctx.Errors.Line {
		t.Error("wanted:\n%s\ngot:\n%s", e.line, ctx.Errors.Line)
	}

	if e.message != ctx.Errors.Message {
		t.Errorf("wanted:\n%s\ngot:\n%s", e.message, ctx.Errors.Message)
	}

}

func TestErrorUnbound(t *testing.T) {
	in := bytes.NewBufferString(`div {
  background: map-get($sprite,139);
}`)
	out := bytes.NewBuffer([]byte(""))
	ctx := Context{}
	err := ctx.Compile(in, out, "")
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
	err := ctx.Compile(in, out, "")
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
	err := ctx.Compile(in, out, "")
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

func TestErrorWarn(t *testing.T) {
	return
	// Disabled while new warn integration is built
	in := bytes.NewBufferString(`
@warn "WARNING";`)
	out := bytes.NewBuffer([]byte(""))
	ctx := Context{}
	err := ctx.Compile(in, out, "")
	_ = err
}
