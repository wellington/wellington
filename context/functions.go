package context

import (
	"fmt"
	"log"
	"path/filepath"

	sw "github.com/drewwells/spritewell"
)

func init() {

	RegisterHandler("image-width($a)", ImageURL)
	RegisterHandler("sprite-map($a,$position:0px,$spacing:5px)", SpriteMap)
	//RegisterHandler("image-width($a)", ImageURL)
}

// ImageURL handles calls to resolve a local image from the
// built css file path.
func ImageURL(ctx *Context, csv UnionSassValue) UnionSassValue {
	var path []string
	err := Unmarshal(csv, &path)
	fmt.Println("Imageurl")
	// This should create and throw a sass error
	if err != nil {
		panic(err)
		fmt.Println(err)
	}
	res, err := Marshal(filepath.Join(ctx.RelativeImage(), path))
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
	fmt.Println("Sprite Map")
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
	imgs.Decode(glob)
	imgs.Combine()
	ctx.Sprites[glob] = imgs
	gpath, err := imgs.Export()
	if err != nil {
		log.Fatal(err)
	}
	return Marshal(gpath)
}

func Sprite(ctx *Context, usv UnionSassValue) UnionSassValue {
	return usv
}
