package sprite_sass

import (
	"bytes"
	"io"
	"testing"
)

func TestLibsassErrorBasic(t *testing.T) {
	in := bytes.NewBufferString(`div {
  @include invalid-function('');
}`)

	testMap := []lError{
		lError{2, "no mixin named invalid-function"},
		lError{Pos: 2, Message: ""},
	}
	ctx, _, _ := setupCtx(in)

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

func TestLibsassErrorUnbound(t *testing.T) {
	in := bytes.NewBufferString(`div {
  background: map-get($sprite,139);
}`)
	setupCtx(in)
}

func TestLibsassErrorFunctionTrace(t *testing.T) {
	in := bytes.NewBufferString(`// Empty line
@function uniqueFnName($file) {
  @return map-get($file,prop);
}
div {
  background: uniqueFnName(randfile);
}`)
	setupCtx(in)
}

func setupCtx(in *bytes.Buffer) (Context, string, error) {
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
	go func(in io.Reader, w io.WriteCloser) {
		err = ctx.Run(in, w, "test")
	}(in, w)

	io.Copy(&out, r)
	return ctx, out.String(), err
}
