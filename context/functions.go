package context

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	sw "github.com/drewwells/spritewell"
)

func init() {

	RegisterHandler("image-width($a)", ImageURL)
	RegisterHandler("sprite-map($a,$position:0px,$spacing:5px)", SpriteMap)
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
	res, err := Marshal(filepath.Join(ctx.RelativeImage(), path[0]))
	if err != nil {
		fmt.Println(err)
	}
	return res
}

func ImageHeight(ctx *Context, usv UnionSassValue) UnionSassValue {
	return usv
}

func ImageWidth(ctx *Context, usv UnionSassValue) UnionSassValue {
	return usv
}

func InlineImage(ctx *Context, usv UnionSassValue) UnionSassValue {
	return usv
}

func SpriteFile(ctx *Context, usv UnionSassValue) UnionSassValue {
	return usv
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
	ctx.Sprites[glob] = imgs
	gpath, err := imgs.Export()
	if err != nil {
		log.Fatal(err)
	}
	res, err := Marshal(gpath)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func Sprite(ctx *Context, usv UnionSassValue) UnionSassValue {
	return usv
}
