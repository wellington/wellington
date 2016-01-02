package handlers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	libsass "github.com/wellington/go-libsass"
)

func ExampleSprite_position() {
	in := bytes.NewBufferString(`
$map: sprite-map("*.png", 10px); // One argument
div {
  background-position: sprite-position($map, "140");
}

div.retina {
  background-position: 10px ceil(nth(sprite-position($map, "140"), 2) /2 );
}`)

	ctx := oldContext()
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
	//   background-position: 0px, -149px; }
	//
	// div.retina {
	//   background-position: 10px -74px; }
}

func TestFuncSpriteFile(t *testing.T) {
	ctx := oldContext()
	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"

	// Add real arguments when sass lists can be [un]marshalled
	lst := []interface{}{"*.png", "139"}
	usv, _ := libsass.Marshal(lst)
	var rsv libsass.SassValue
	SpriteFile(ctx, usv, &rsv)
	var glob, path string
	err := libsass.Unmarshal(rsv, &glob, &path)
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

// func TestFuncSpriteMap(t *testing.T) {
// 	ctx := oldContext()
// 	ctx.BuildDir = "../test/build"
// 	ctx.GenImgDir = "../test/build/img"
// 	ctx.ImageDir = "../test/img"

// 	// Add real arguments when sass lists can be [un]marshalled
// 	lst := []interface{}{
// 		"*.png",
// 		libs.SassNumber{Value: 5, Unit: "px"},
// 	}
// 	usv, _ := libsass.Marshal(lst)
// 	var rsv libsass.SassValue
// 	SpriteMap(ctx, usv, &rsv)
// 	var path string
// 	err := libsass.Unmarshal(rsv, &path)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if e := "*.png5"; e != path {
// 		t.Errorf("got: %s wanted: %s", path, e)
// 	}
// }

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

	ctx := oldContext()

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
	in := bytes.NewBufferString(`$map: sprite-map("*.png",10px);
div {
  content: $map;
}`)

	ctx := oldContext()

	ctx.ImageDir = "../test/img"
	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"

	var out bytes.Buffer
	err := ctx.Compile(in, &out)
	if err != nil {
		t.Error(err)
	}
	exp := `div {
  content: *.png10; }
`
	if exp != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), exp)
	}
}
