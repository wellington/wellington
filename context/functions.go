package context

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	sw "github.com/drewwells/spritewell"
)

func init() {

	RegisterHandler("sprite-map($glob,$position:0px,$spacing:5px)", SpriteMap)
	RegisterHandler("sprite-file($map, $name)", SpriteFile)
	RegisterHandler("image-url($name)", ImageURL)
	RegisterHandler("image-height($path)", ImageHeight)
	RegisterHandler("image-width($path)", ImageWidth)
}

// ImageURL handles calls to resolve a local image from the
// built css file path.
func ImageURL(ctx *Context, csv UnionSassValue) UnionSassValue {
	var path []string
	err := Unmarshal(csv, &path)
	// This should create and throw a sass error
	if err != nil {
		fmt.Println(err)
	}
	url := filepath.Join(ctx.RelativeImage(), path[0])
	res, err := Marshal(fmt.Sprintf("url('%s')", url))
	if err != nil {
		fmt.Println(err)
	}
	return res
}

func ImageHeight(ctx *Context, usv UnionSassValue) UnionSassValue {
	var (
		name string
	)
	err := Unmarshal(usv, &name)
	if err != nil {
		fmt.Println(err)
	}
	imgs := sw.ImageList{
		ImageDir:  ctx.ImageDir,
		GenImgDir: ctx.GenImgDir,
	}
	imgs.Decode(name)
	imgs.Combine()
	height := imgs.SImageHeight(name)
	Hheight := SassNumber{
		value: float64(height),
		unit:  "px",
	}
	res, err := Marshal(Hheight)
	if err != nil {
		fmt.Println(err)
	}
	return res
}

func ImageWidth(ctx *Context, usv UnionSassValue) UnionSassValue {
	var (
		name string
	)
	err := Unmarshal(usv, &name)
	if err != nil {
		fmt.Println(err)
	}
	imgs := sw.ImageList{
		ImageDir:  ctx.ImageDir,
		GenImgDir: ctx.GenImgDir,
	}
	imgs.Decode(name)
	imgs.Combine()
	v := imgs.SImageWidth(name)
	vv := SassNumber{
		value: float64(v),
		unit:  "px",
	}
	res, err := Marshal(vv)
	if err != nil {
		fmt.Println(err)
	}
	return res
}

func InlineImage(ctx *Context, usv UnionSassValue) UnionSassValue {
	return usv
}

func SpriteFile(ctx *Context, usv UnionSassValue) UnionSassValue {
	var glob, name string
	err := Unmarshal(usv, &glob, &name)
	if err != nil {
		panic(err)
	}
	sprite := ctx.Sprites[glob].File(name)
	res, err := Marshal(sprite)
	if err != nil {
		panic(err)
	}
	return res
}

// SpriteMap generates a sprite from the passed glob and sprite
// parameters.
func SpriteMap(ctx *Context, usv UnionSassValue) UnionSassValue {
	var glob string
	var spacing float64
	var position float64
	err := Unmarshal(usv, &glob, &spacing, &position)
	if err != nil {
		log.Fatal(err)
	}
	imgs := sw.ImageList{
		ImageDir:  ctx.ImageDir,
		BuildDir:  ctx.BuildDir,
		GenImgDir: ctx.GenImgDir,
		Vertical:  true,
	}
	if cglob, err := strconv.Unquote(glob); err == nil {
		glob = cglob
	}
	err = imgs.Decode(glob)
	if err != nil {
		log.Fatal(err)
	}
	imgs.Combine()
	gpath, err := imgs.Export()
	if err != nil {
		log.Fatal(err)
	}
	res, err := Marshal(gpath)
	ctx.Sprites[gpath] = imgs
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func Sprite(ctx *Context, usv UnionSassValue) UnionSassValue {
	return usv
}
