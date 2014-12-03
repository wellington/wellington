package context

import (
	"log"
	"path/filepath"

	sw "github.com/drewwells/spritewell"
)

func init() {

	RegisterHandler("image-url($a)", ImageUrl)
}

// ImageUrl handles calls to Rel()
func ImageUrl(ctx *Context, csv UnionSassValue) UnionSassValue {
	var path string
	err := Unmarshal(csv, &path)
	// This should create and throw a sass error
	if err != nil {
		log.Fatal(err)
	}
	res, _ := Marshal(filepath.Join(ctx.RelativeImage(), path))
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

func SpriteMap(ctx *Context, usv UnionSassValue) UnionSassValue {
	var glob string
	Unmarshal(usv, &glob)
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
