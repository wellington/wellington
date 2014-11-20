package sprite_sass

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
		Parser: Parser{
			MainFile: "testname",
		},
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

	err = ctx.Run(reader, &out, "test/sass")
	if err != nil {
		// This will mask iport errors
		// panic(err)
	}

	return ctx, out.String(), err
}

func TestErrorBasic(t *testing.T) {
	in := bytes.NewBufferString(`div {
  @include invalid-function('');
}`)

	ctx, _, _ := setupCtx(in)

	testMap := []lError{
		lError{2, "no mixin named invalid-function"},
	}

	for i := range testMap {
		e, w := testMap[i], ctx.errors.Errors[i]
		if e.Pos != w.Pos {
			t.Errorf("mismatch expected: %d was: %d",
				e.Pos, w.Pos)
		}

		if e.Message != w.Message {
			t.Errorf("mismatch expected: %s was: %s",
				e.Message, w.Message)
		}
	}

}

func TestErrorUnbound(t *testing.T) {
	in := bytes.NewBufferString(`div {
  background: map-get($sprite,139);
}`)
	ctx, _, _ := setupCtx(in)

	testMap := []lError{
		lError{2, "unbound variable $sprite"},
	}

	for i := range testMap {
		e, w := testMap[i], ctx.errors.Errors[i]
		if e.Pos != w.Pos {
			t.Errorf("mismatch expected: %d was: %d",
				e.Pos, w.Pos)
		}

		if e.Message != w.Message {
			t.Errorf("mismatch expected: %d was: %d",
				e.Message, w.Message)
		}
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
	ctx, _, _ := setupCtx(in)

	testMap := []lError{
		lError{3, "argument `$map` of `map-get($map $key)` must be a map"},
		// lError{3, "in function `map-get`"},
		// lError{3, "in function `uniqueFnName`"},
		// lError{6, ""},
	}

	if len(ctx.errors.Errors) != len(testMap) {
		t.Error("Test map and error map are not the same length")
	}

	for i := range testMap {
		e, w := testMap[i], ctx.errors.Errors[i]
		if e.Pos != w.Pos {
			t.Errorf("mismatch expected: %d was: %d",
				e.Pos, w.Pos)
		}

		if e.Message != w.Message {
			t.Errorf("mismatch expected: %d was: %d",
				e.Message, w.Message)
		}
	}
}

func TestErrorImport(t *testing.T) {
	//return // Tests on files cause race conditions
	ctx, _, _ := setupCtx("test/sass/failimport.scss")

	testMap := []lError{
		lError{66, "invalid top-level expression"},
	}

	for i := range testMap {
		e, w := testMap[i], ctx.errors.Errors[i]
		if e.Pos != w.Pos {
			t.Errorf("mismatch expected: %d:%s was: %d:%s",
				e.Pos, e.Message, w.Pos, w.Message)
		}

		if e.Message != w.Message {
			t.Errorf("mismatch expected: %d was: %d",
				e.Message, w.Message)
		}
	}
}

func TestErrorNonmap(t *testing.T) {
	in := bytes.NewBufferString(`
@import "sprite";
div {
  height: image-height('test/img/139.png');
}`)
	ctx, _, _ := setupCtx(in)

	if len(ctx.errors.Errors) > 0 {
		t.Error("Non-warn thrown for image-height('file')")
	}

	return // libsass throws warnings to stdout, let's wait to test this
	warnLine := "?"

	if e := "WARNING: `test/img/139.png` is not a map."; e != warnLine {
		t.Errorf("Warning did not match expected:\n%s\nwas:\n%s\n",
			e, warnLine)
	}
}
