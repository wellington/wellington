package context

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/wellington/spritewell"
)

func wrapCallback(sc SassCallback, ch chan UnionSassValue) SassCallback {
	return func(c *Context, usv UnionSassValue) UnionSassValue {
		usv = sc(c, usv)
		ch <- usv
		return usv
	}
}

func testSprite(ctx *Context) {
	// Generate test sprite
	imgs := spritewell.ImageList{
		ImageDir:  ctx.ImageDir,
		BuildDir:  ctx.BuildDir,
		GenImgDir: ctx.GenImgDir,
	}
	glob := "*.png"
	err := imgs.Decode(glob)
	if err != nil {
		panic(err)
	}
	_, err = imgs.Combine()
	if err != nil {
		panic(err)
	}
}

func setupCtx(r io.Reader, out io.Writer, cookies ...Cookie) (*Context, UnionSassValue, error) {
	var usv UnionSassValue
	ctx := NewContext()
	ctx.OutputStyle = NESTED_STYLE
	ctx.IncludePaths = make([]string, 0)
	ctx.BuildDir = "test/build"
	ctx.ImageDir = "test/img"
	ctx.GenImgDir = "test/build/img"
	ctx.Out = ""

	testSprite(ctx)
	cc := make(chan UnionSassValue, len(cookies))
	// If callbacks were made, add them to the context
	// and create channels for communicating with them.
	if len(cookies) > 0 {
		cs := make([]Cookie, len(cookies))
		for i, c := range cookies {
			cs[i] = Cookie{
				c.sign,
				wrapCallback(c.fn, cc),
				ctx,
			}
		}
		usv = <-cc
	}
	err := ctx.Compile(r, out)
	return ctx, usv, err
}

func TestFuncImageURL(t *testing.T) {
	ctx := Context{
		BuildDir: "test/build",
		ImageDir: "test/img",
	}

	usv := testMarshal(t, []string{"image.png"})
	usv = ImageURL(&ctx, usv)
	var path string
	Unmarshal(usv, &path)
	if e := "url('../img/image.png')"; e != path {
		t.Errorf("got: %s wanted: %s", path, e)
	}

	// Test sending invalid date to imageURL
	usv = testMarshal(t, SassNumber{1, "px"})
	_ = usv
	errusv := ImageURL(&ctx, usv)
	var s string
	merr := Unmarshal(errusv, &s)
	if merr != nil {
		t.Error(merr)
	}

	e := "Sassvalue is type context.SassNumber and has value {1 px} but expected slice"

	if e != s {
		t.Errorf("got:\n%s\nwanted:\n%s", s, e)
	}

}

func TestFuncSpriteMap(t *testing.T) {
	ctx := NewContext()
	ctx.BuildDir = "test/build"
	ctx.GenImgDir = "test/build/img"
	ctx.ImageDir = "test/img"

	// Add real arguments when sass lists can be [un]marshalled
	lst := []interface{}{"*.png", SassNumber{5, "px"}}
	usv := testMarshal(t, lst)
	usv = SpriteMap(ctx, usv)
	var path string
	err := Unmarshal(usv, &path)
	if err != nil {
		t.Error(err)
	}

	if e := "*.png5"; e != path {
		t.Errorf("got: %s wanted: %s", path, e)
	}
}

func TestFuncSpriteFile(t *testing.T) {
	ctx := NewContext()
	ctx.BuildDir = "test/build"
	ctx.GenImgDir = "test/build/img"
	ctx.ImageDir = "test/img"

	// Add real arguments when sass lists can be [un]marshalled
	lst := []interface{}{"*.png", "139"}
	usv := testMarshal(t, lst)
	usv = SpriteFile(ctx, usv)
	var glob, path string
	err := Unmarshal(usv, &glob, &path)
	if err != nil {
		t.Error(err)
	}

	if e := "*.png"; e != glob {
		t.Errorf("got: %s wanted: %s", e, glob)
	}

	if e := "139"; e != path {
		t.Errorf("got: %s wanted: %s", e, path)
	}

}

func TestCompileSpriteMap(t *testing.T) {
	in := bytes.NewBufferString(`
$aritymap: sprite-map("*.png",0px); // Optional arguments
$map: sprite-map("*.png"); // One argument
div {
width: $map;
height: $aritymap;
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
  width: testimg-8121ae.png;
  height: *.png0; }
`

	if exp != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), exp)
	}
}

func TestCompileSpritePaddingMap(t *testing.T) {
	in := bytes.NewBufferString(`$map: sprite-map("dual/*.png",10px);
div {
  content: $map;
}`)

	ctx := NewContext()

	ctx.ImageDir = "test/img"
	ctx.BuildDir = "test/build"
	ctx.GenImgDir = "test/build/img"

	var out bytes.Buffer
	err := ctx.Compile(in, &out)
	if err != nil {
		t.Error(err)
	}
	exp := `div {
  content: dual/*.png10; }
`

	if exp != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), exp)
	}

	esz := int64(28160)
	f, err := os.Stat(filepath.Join(ctx.GenImgDir, "testimgdual-ab7eb7.png"))

	if os.IsNotExist(err) {
		t.Error(err)
	} else if f.Size() != esz {
		t.Errorf("got: %d wanted: %d", f.Size(), esz)
	}
}

func TestFuncImageHeight(t *testing.T) {
	in := bytes.NewBufferString(`div {
    height: image-height("139");
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(in, &out)

	if err != nil {
		t.Error(err)
	}

	e := `div {
  height: 139px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestRegImageWidth(t *testing.T) {
	in := bytes.NewBufferString(`div {
    height: image-width("139");
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  height: 96px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestRegSpriteImageHeight(t *testing.T) {
	in := bytes.NewBufferString(`$map: sprite-map("*.png");
div {
  height: image-height(sprite-file($map,"139"));
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  height: 139px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestRegSpriteImageWidth(t *testing.T) {
	in := bytes.NewBufferString(`$map: sprite-map("*.png");
div {
  width: image-width(sprite-file($map,"139"));
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  width: 96px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestRegImageURL(t *testing.T) {
	in := bytes.NewBufferString(`
div {
    background: image-url("139.png");
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  background: url('../img/139.png'); }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestRegInlineImage(t *testing.T) {
	in := bytes.NewBufferString(`
div {
    background: inline-image("pixel/1x1.png");
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  background: url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAMAAAAoyzS7AAAAA1BMVEX/TQBcNTh/AAAAAXRSTlMz/za5cAAAAA5JREFUeJxiYgAEAAD//wAGAAP60FmuAAAAAElFTkSuQmCC'); }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}
