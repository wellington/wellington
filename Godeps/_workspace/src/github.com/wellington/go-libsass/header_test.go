package context

import (
	"bytes"
	"testing"
)

func TestSassHeader(t *testing.T) {
	t.Skip("disabled after extraction")
	var out bytes.Buffer
	ctx := Context{}

	ctx.Init(NewSassOptions())
	err := ctx.FileCompile("../test/sass/sprite-dimensions.scss", &out)
	if err != nil {
		t.Fatal(err)
	}
	e := `div {
  height: image-height(sprite-file(sprite-map("../img/*.png"), "139"));
  width: image-width(sprite-file(sprite-map("../img/*.png"), "139")); }
`
	if e != out.String() {
		t.Fatalf("got:\n%s\nwanted:\n%s", out.String(), e)
	}

}
