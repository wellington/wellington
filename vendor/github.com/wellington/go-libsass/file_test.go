package libsass

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFile_resolved(t *testing.T) {
	path := "test/scss/main.scss"
	var out bytes.Buffer
	ctx := newContext()
	err := ctx.fileCompile(path, &out, "", "")
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

	abs, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}

	relPath := strings.TrimPrefix(ctx.ResolvedImports[0], abs)

	if !strings.HasSuffix(relPath, path) {
		t.Errorf("got: %s wanted: %s", relPath, e)
	}

}
