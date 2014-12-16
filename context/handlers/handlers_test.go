package handlers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/wellington/spritewell"
	cx "github.com/wellington/wellington/context"
)

func init() {
	os.MkdirAll("../test/build/img", 0777)
}

func wrapCallback(sc cx.SassCallback, ch chan cx.UnionSassValue) cx.SassCallback {
	return func(c *cx.Context, usv cx.UnionSassValue) cx.UnionSassValue {
		usv = sc(c, usv)
		ch <- usv
		return usv
	}
}

func testSprite(ctx *cx.Context) {
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

func setupCtx(r io.Reader, out io.Writer, cookies ...cx.Cookie) (*cx.Context, cx.UnionSassValue, error) {
	var usv cx.UnionSassValue
	ctx := cx.NewContext()
	ctx.OutputStyle = cx.NESTED_STYLE
	ctx.IncludePaths = make([]string, 0)
	ctx.BuildDir = "../test/build"
	ctx.ImageDir = "../test/img"
	ctx.FontDir = "../test/font"
	ctx.GenImgDir = "../test/build/img"
	ctx.Out = ""

	testSprite(ctx)
	cc := make(chan cx.UnionSassValue, len(cookies))
	// If callbacks were made, add them to the context
	// and create channels for communicating with them.
	if len(cookies) > 0 {
		cs := make([]cx.Cookie, len(cookies))
		for i, c := range cookies {
			cs[i] = cx.Cookie{
				Sign: c.Sign,
				Fn:   wrapCallback(c.Fn, cc),
				Ctx:  ctx,
			}
		}
		usv = <-cc
	}
	err := ctx.Compile(r, out)
	return ctx, usv, err
}

func TestFuncImageURL(t *testing.T) {
	ctx := cx.Context{
		BuildDir: "test/build",
		ImageDir: "test/img",
	}

	usv, _ := cx.Marshal([]string{"image.png"})
	usv = ImageURL(&ctx, usv)
	var path string
	cx.Unmarshal(usv, &path)
	if e := "url('../img/image.png')"; e != path {
		t.Errorf("got: %s wanted: %s", path, e)
	}

	// Test sending invalid date to imageURL
	usv, _ = cx.Marshal(cx.SassNumber{Value: 1, Unit: "px"})
	_ = usv
	errusv := ImageURL(&ctx, usv)
	var s string
	merr := cx.Unmarshal(errusv, &s)
	if merr != nil {
		t.Error(merr)
	}

	e := "Sassvalue is type context.SassNumber and has value {1 px} but expected slice"

	if e != s {
		t.Errorf("got:\n%s\nwanted:\n%s", s, e)
	}

}

func TestFuncSpriteMap(t *testing.T) {
	ctx := cx.NewContext()
	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"

	// Add real arguments when sass lists can be [un]marshalled
	lst := []interface{}{"*.png", cx.SassNumber{Value: 5, Unit: "px"}}
	usv, _ := cx.Marshal(lst)
	usv = SpriteMap(ctx, usv)
	var path string
	err := cx.Unmarshal(usv, &path)
	if err != nil {
		t.Error(err)
	}

	if e := "*.png5"; e != path {
		t.Errorf("got: %s wanted: %s", path, e)
	}
}

func TestFuncSpriteFile(t *testing.T) {
	ctx := cx.NewContext()
	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"

	// Add real arguments when sass lists can be [un]marshalled
	lst := []interface{}{"*.png", "139"}
	usv, _ := cx.Marshal(lst)
	usv = SpriteFile(ctx, usv)
	var glob, path string
	err := cx.Unmarshal(usv, &glob, &path)
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
$aritymap: sprite-map("*.png", 0px); // Optional arguments
$map: sprite-map("*.png"); // One argument
$paddedmap: sprite-map("*.png", 1px); // One argument
div {
width: $map;
height: $aritymap;
line-height: $paddedmap;
}`)

	ctx := cx.NewContext()

	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"
	var out bytes.Buffer
	err := ctx.Compile(in, &out)
	if err != nil {
		t.Error(err)
	}
	exp := `div {
  width: *.png0;
  height: *.png0;
  line-height: *.png1; }
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

	ctx := cx.NewContext()

	ctx.ImageDir = "../test/img"
	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"

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

func TestRegInlineImageFail(t *testing.T) {
	var f *os.File
	old := os.Stdout
	os.Stdout = f
	defer func() { os.Stdout = old }()
	in := bytes.NewBufferString(`
div {
    background: inline-image("image.svg");
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  background: inline-image: image.svg filetype .svg is not supported; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestFontURLFail(t *testing.T) {
	r, w, _ := os.Pipe()
	old := os.Stdout
	defer func() { os.Stdout = old }()
	os.Stdout = w
	in := bytes.NewBufferString(`@font-face {
  src: font-url("arial.eot");
}`)
	var out bytes.Buffer
	ctx := cx.Context{}
	err := ctx.Compile(in, &out)

	if err != nil {
		t.Error(err)
	}

	outC := make(chan string)
	go func(r *os.File) {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}(r)

	w.Close()
	stdout := <-outC

	if e := "font-url: font path not set\n"; e != stdout {
		t.Errorf("got:\n%s\nwanted:\n%s\n", stdout, e)
	}

}

func ExampleFontURL() {
	in := bytes.NewBufferString(`
$path: font-url("arial.eot", true);
@font-face {
  src: font-url("arial.eot");
  src: url("#{$path}");
}`)

	_, _, err := setupCtx(in, os.Stdout)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// @font-face {
	//   src: url("../font/arial.eot");
	//   src: url("../font/arial.eot"); }
}

func ExampleSprite() {
	in := bytes.NewBufferString(`
$map: sprite-map("dual/*.png", 10px); // One argument
div {
  background: sprite($map, "140");
}`)

	ctx := cx.NewContext()

	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"
	var out bytes.Buffer
	err := ctx.Compile(in, &out)
	if err != nil {
		fmt.Println(err)
	}

	io.Copy(os.Stdout, &out)

	// Output:
	// div {
	//   background: url("img/img-b798ab.png") -0px -149px; }

}

func TestSprite(t *testing.T) {
	in := bytes.NewBufferString(`
$map: sprite-map("dual/*.png", 10px);
div {
  background: sprite($map, "140", 0, 0);
}`)

	ctx := cx.NewContext()

	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"
	var out bytes.Buffer
	err := ctx.Compile(in, &out)

	e := `Error > stdin:4
error in C function sprite: Please specify unit for offset ie. (2px)
Backtrace:
	stdin:4, in function ` + "`sprite`" + `
	stdin:4

$map: sprite-map("dual/*.png", 10px);
div {
  background: sprite($map, "140", 0, 0);
}
`
	if e != err.Error() {
		t.Errorf("got:\n~%s~\nwanted:\n~%s~\n", err.Error(), e)
	}

}

func BenchmarkSprite(b *testing.B) {
	ctx := cx.NewContext()
	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"
	// Add real arguments when sass lists can be [un]marshalled
	lst := []interface{}{"*.png", cx.SassNumber{Value: 5, Unit: "px"}}
	usv, _ := cx.Marshal(lst)

	for i := 0; i < b.N; i++ {
		usv = SpriteMap(ctx, usv)
	}
	// Debug if needed
	// var s string
	// Unmarshal(usv, &s)
	// fmt.Println(s)
}
