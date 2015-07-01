package context

import (
	"bytes"
	"testing"
)

func TestFile_resolved(t *testing.T) {
	path := "test/scss/main.scss"
	var out bytes.Buffer
	ctx := Context{}
	err := ctx.FileCompile(path, &out)
	if err != nil {
		panic(err)
	}

	e := ``

	if e != out.String() {
		t.Errorf("wanted:\n%s\n"+
			"got:\n%s\n", e, out.String())
	}

	if e := 1; len(ctx.ResolvedImports) != e {
		t.Errorf("got: %d wanted: %d", len(ctx.ResolvedImports), e)
	}

	if e := path; ctx.ResolvedImports[0] != e {
		t.Errorf("got: %s wanted: %s", ctx.ResolvedImports[0], e)
	}

}
