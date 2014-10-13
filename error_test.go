package sprite_sass

import (
	"bytes"
	"io"
	"log"
	"testing"
)

func TestErrorBasic(t *testing.T) {
	in := bytes.NewBufferString(`div {
  @include invalid-function('');
}`)

	ctx, _, _ := setupCtx(in)

	testMap := []lError{
		lError{2, "no mixin named invalid-function"},
		lError{2, ""},
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
		lError{3, "in function `map-get`"},
		lError{3, "in function `uniqueFnName`"},
		lError{6, ""},
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

func setupCtx(f interface{}) (Context, string, error) {
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		BuildDir:     "test/build",
		Out:          "",
		Parser: Parser{
			MainFile: "testname",
		},
	}
	var (
		out bytes.Buffer
		err error
	)
	r, w := io.Pipe()

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

	go func(in io.Reader, w io.WriteCloser) {
		err = ctx.Run(in, w, "test/sass")
		if err != nil {
			// log.Fatal(err)
		}
	}(reader, w)

	io.Copy(&out, r)
	return ctx, out.String(), err
}

func TestErrorImport(t *testing.T) {
	ctx, _, _ := setupCtx("test/sass/failimport.scss")

	testMap := []lError{
		lError{42, "invalid top-level expression"},
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

func TestErrorNonmap(t *testing.T) {
	in := bytes.NewBufferString(`
@import "sprite";
div {
  height: image-height('test/img/139.png');
}`)
	ctx, _, _ := setupCtx(in)

	testMap := []lError{
		lError{20, "argument `$map` of `map-get($map $key)` must be a map"},
		lError{20, "in function `map-get`"},
		lError{20, "in function `image-height`"},
		lError{39, ""},
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
