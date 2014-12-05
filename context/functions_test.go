package context

import (
	"bytes"
	"io"
	"testing"
)

func wrapCallback(sc SassCallback, ch chan UnionSassValue) SassCallback {
	return func(c *Context, usv UnionSassValue) UnionSassValue {
		usv = sc(c, usv)
		ch <- usv
		return usv
	}
}

func setupCtx(r io.Reader, out io.Writer, cookies ...Cookie) (Context, chan UnionSassValue, chan error) {
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		BuildDir:     "test/build",
		ImageDir:     "test/img",
		GenImgDir:    "test/build/img",
		Out:          "",
	}
	var cc chan UnionSassValue
	ec := make(chan error, 1)
	_ = ec
	// If callbacks were made, add them to the context
	// and create channels for communicating with them.
	if len(cookies) > 0 {
		cs := make([]Cookie, len(cookies))
		for i, c := range cookies {
			cs[i] = Cookie{
				c.sign,
				wrapCallback(c.fn, cc),
				&ctx,
			}
		}
	}

	ec <- ctx.Compile(r, out)

	return ctx, cc, ec
}

func TestFuncImageUrl(t *testing.T) {
	ctx := Context{
		BuildDir: "test/build",
		ImageDir: "test/img",
	}

	usv, _ := Marshal("image.png")
	usv = ImageUrl(&ctx, usv)
	var path string
	Unmarshal(usv, &path)

	if e := "../img/image.png"; e != path {
		t.Errorf("got: %s wanted: %s", path, e)
	}
}

func TestFuncSpriteMap(t *testing.T) {
	ctx := NewContext()
	ctx.BuildDir = "test/build"
	ctx.GenImgDir = "test/build/img"
	ctx.ImageDir = "test/img"

	// Add real arguments when sass lists can be [un]marshalled
	lst := []interface{}{"*.png", float64(5), float64(0)}
	usv := Marshal(lst)
	usv = SpriteMap(ctx, usv)
	var path string
	Unmarshal(usv, &path)

	if e := "test/build/img/image-8121ae.png"; e != path {
		t.Errorf("got: %s wanted: %s", path, e)
	}
}

func TestCompileSpriteMap(t *testing.T) {
	in := bytes.NewBufferString(`
$map: sprite-map("*.png",1,2);
div {
width: $map;
}`)

	ctx := NewContext()

	ctx.BuildDir = "test/build"
	ctx.GenImgDir = "test/build/img"
	ctx.ImageDir = "test/img"
	var out bytes.Buffer
	err := ctx.Compile(in, &out)
	if err != nil {
		t.Error(err)
	}
	exp := `div {
  width: test/build/img/image-8121ae.png; }
`

	if exp != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), exp)
	}
}
