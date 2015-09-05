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

	abs, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}

	relPath := strings.TrimPrefix(ctx.ResolvedImports[0], abs)

	if e := string(filepath.Separator) + path; relPath != e {
		t.Errorf("got: %s wanted: %s", relPath, e)
	}

}
