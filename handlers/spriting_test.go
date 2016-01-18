package handlers

import (
	"bytes"
	"log"
	"os"
	"testing"

	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/wellington/payload"
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

	comp, err := libsass.New(os.Stdout, in,
		libsass.Payload(payload.New()),
		libsass.ImgDir("../test/img"),
		libsass.BuildDir("../test/build"),
		libsass.ImgBuildDir("../test/build/img"),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = comp.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// div {
	//   background-position: 0px, -149px; }
	//
	// div.retina {
	//   background-position: 10px -74px; }
}

func TestFuncSpriteFile(t *testing.T) {

	comp, err := libsass.New(nil, nil,
		libsass.Payload(payload.New()),
		libsass.ImgDir("../test/img"),
		libsass.BuildDir("../test/build"),
		libsass.ImgBuildDir("../test/build/img"),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Add real arguments when sass lists can be [un]marshalled
	lst := []interface{}{"*.png", "139"}
	usv, err := libsass.Marshal(lst)
	if err != nil {
		t.Fatal(err)
	}

	rsv, err := SpriteFile(libsass.NewCompilerContext(comp), usv)
	if err != nil {
		t.Fatal(err)
	}
	var glob, path string
	err = libsass.Unmarshal(*rsv, &glob, &path)
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

	var out bytes.Buffer
	_, err := setupComp(t, in, &out)
	if err != nil {
		t.Fatal(err)
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

	var out bytes.Buffer
	_, err := setupComp(t, in, &out)
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
