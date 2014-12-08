package context

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strconv"

	sw "github.com/drewwells/spritewell"
)

func init() {

	RegisterHandler("sprite-map($glob,$position:0px,$spacing:5px)", SpriteMap)
	RegisterHandler("sprite-file($map, $name)", SpriteFile)
	RegisterHandler("image-url($name)", ImageURL)
	RegisterHandler("image-height($path)", ImageHeight)
	RegisterHandler("image-width($path)", ImageWidth)
	RegisterHandler("inline-image($path)", InlineImage)
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
		glob string
		name string
	)
	err := Unmarshal(usv, &name)
	// Check for sprite-file override first
	if err != nil {
		var inf interface{}
		var infs []interface{}
		// Can't unmarshal to []interface{}, so unmarshal to
		// interface{} then reflect it into a []interface{}
		err = Unmarshal(usv, &inf)
		k := reflect.ValueOf(&infs).Elem()
		k.Set(reflect.ValueOf(inf))

		if err != nil {
			log.Fatal(err)
			return Error(err)
		} else {
			glob = infs[0].(string)
			name = infs[1].(string)
		}
	}
	imgs := sw.ImageList{
		ImageDir:  ctx.ImageDir,
		GenImgDir: ctx.GenImgDir,
	}
	if glob == "" {
		if hit, ok := ctx.Imgs.M[name]; ok {
			imgs = hit
		} else {
			imgs.Decode(name)
			imgs.Combine()
			ctx.Imgs.Lock()
			ctx.Imgs.M[name] = imgs
			ctx.Imgs.Unlock()
		}
	} else {
		ctx.Sprites.Lock()
		imgs = ctx.Sprites.M[glob]
		ctx.Sprites.Unlock()
	}
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

// ImageWidth takes a file path (or sprite glob) and returns the
// height in pixels of the image being referenced.
func ImageWidth(ctx *Context, usv UnionSassValue) UnionSassValue {
	var (
		glob, name string
	)
	err := Unmarshal(usv, &name)
	// Check for sprite-file override first
	if err != nil {
		var inf interface{}
		var infs []interface{}
		// Can't unmarshal to []interface{}, so unmarshal to
		// interface{} then reflect it into a []interface{}
		err = Unmarshal(usv, &inf)
		k := reflect.ValueOf(&infs).Elem()
		k.Set(reflect.ValueOf(inf))

		if err != nil {
			log.Fatal(err)
			return Error(err)
		} else {
			glob = infs[0].(string)
			name = infs[1].(string)
		}
	}
	imgs := sw.ImageList{
		ImageDir:  ctx.ImageDir,
		GenImgDir: ctx.GenImgDir,
	}
	if glob == "" {
		if hit, ok := ctx.Imgs.M[name]; ok {
			imgs = hit
		} else {
			imgs.Decode(name)
			imgs.Combine()
			ctx.Imgs.Lock()
			ctx.Imgs.M[name] = imgs
			ctx.Imgs.Unlock()
		}
	} else {
		ctx.Sprites.Lock()
		imgs = ctx.Sprites.M[glob]
		ctx.Sprites.Unlock()
	}
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
	err = imgs.Decode(name)
	if err != nil {
		return Error(err)
	}
	_, err = imgs.Combine()
	if err != nil {
		fmt.Println(err)
	}
	str := imgs.Inline()
	res, err := Marshal(str)
	if err != nil {
		fmt.Println(err)
	}
	return res
}

// SpriteFile proxies the sprite glob and image name through.
func SpriteFile(ctx *Context, usv UnionSassValue) UnionSassValue {
	var glob, name string
	err := Unmarshal(usv, &glob, &name)
	if err != nil {
		panic(err)
	}
	infs := []interface{}{glob, name}
	res, err := Marshal(infs)
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
	gpath, err := imgs.Combine()
	if err != nil {
		log.Fatal(err)
	}

	_, err = imgs.Export()
	if err != nil {
		log.Fatal(err)
	}

	res, err := Marshal(gpath)
	ctx.Sprites.Lock()
	ctx.Sprites.M[gpath] = imgs
	ctx.Sprites.Unlock()
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func Sprite(ctx *Context, usv UnionSassValue) UnionSassValue {
	return usv
}
