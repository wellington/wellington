package context

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"

	sw "github.com/wellington/spritewell"
)

func init() {

	RegisterHandler("sprite-map($glob, $spacing: 0px)", SpriteMap)
	RegisterHandler("sprite-file($map, $name)", SpriteFile)
	RegisterHandler("image-url($name)", ImageURL)
	RegisterHandler("image-height($path)", ImageHeight)
	RegisterHandler("image-width($path)", ImageWidth)
	RegisterHandler("inline-image($path)", InlineImage)
	RegisterHandler("font-url($path, $raw: false)", FontURL)
	RegisterHandler("sprite($map, $name, $offsetX: 0px, $offsetY: 0px)", Sprite)
}

// ImageURL handles calls to resolve a local image from the
// built css file path.
func ImageURL(ctx *Context, csv UnionSassValue) UnionSassValue {
	var path []string
	err := Unmarshal(csv, &path)
	// This should create and throw a sass error
	if err != nil {
		return Error(err)
	}
	url := filepath.Join(ctx.RelativeImage(), path[0])
	res, err := Marshal(fmt.Sprintf("url('%s')", url))
	if err != nil {
		return Error(err)
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
		ctx.Sprites.RLock()
		imgs = ctx.Sprites.M[glob]
		ctx.Sprites.RUnlock()
	}
	height := imgs.SImageHeight(name)
	Hheight := SassNumber{
		Value: float64(height),
		Unit:  "px",
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
		ctx.Sprites.RLock()
		imgs = ctx.Sprites.M[glob]
		ctx.Sprites.RUnlock()
	}
	v := imgs.SImageWidth(name)
	vv := SassNumber{
		Value: float64(v),
		Unit:  "px",
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

	if !sw.CanDecode(filepath.Ext(name)) {
		s := fmt.Sprintf("inline-image: %s filetype %s is not supported",
			name, filepath.Ext(name))
		fmt.Println(s)
		// TODO: Replace with warning
		res, _ := Marshal(s)
		return res
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
		return Error(err)
	}
	infs := []interface{}{glob, name}
	res, err := Marshal(infs)
	return res
}

// Sprite returns the source and background position for an image in the
// spritesheet.
func Sprite(ctx *Context, usv UnionSassValue) UnionSassValue {
	var glob, name string
	var offsetX, offsetY SassNumber
	_, _ = offsetX, offsetY // TODO: ignore these for now
	err := Unmarshal(usv, &glob, &name, &offsetX, &offsetY)
	if err != nil {
		return Error(err)
	}
	ctx.Sprites.RLock()
	defer ctx.Sprites.RUnlock()
	imgs, ok := ctx.Sprites.M[glob]
	if !ok {
		keys := make([]string, 0, len(ctx.Sprites.M))
		for i := range ctx.Sprites.M {
			keys = append(keys, i)
		}

		return Error(fmt.Errorf(
			"Variable not found matching glob: %s sprite:%s", glob, name))
	}

	path, err := imgs.OutputPath()
	if err != nil {
		return Error(err)
	}

	if imgs.Lookup(name) == -1 {
		return Error(fmt.Errorf("image %s not found\n"+
			"   try one of these: %v", name, imgs.Paths))
	}
	// This is an odd name for what it does
	pos := imgs.GetPack(imgs.Lookup(name))
	relPath, err := filepath.Rel(ctx.BuildDir,
		filepath.Join(ctx.GenImgDir, path))
	if err != nil {
		return Error(err)
	}
	str, err := Marshal(fmt.Sprintf(`url("%s") -%dpx -%dpx`,
		relPath, pos.X, pos.Y))
	if err != nil {
		return Error(err)
	}
	return str
}

// SpriteMap returns a sprite from the passed glob and sprite
// parameters.
func SpriteMap(ctx *Context, usv UnionSassValue) UnionSassValue {
	var glob string
	var spacing SassNumber
	err := Unmarshal(usv, &glob, &spacing)
	if err != nil {
		return Error(err)
	}
	imgs := sw.ImageList{
		ImageDir:  ctx.ImageDir,
		BuildDir:  ctx.BuildDir,
		GenImgDir: ctx.GenImgDir,
	}
	imgs.Padding = int(spacing.Value)
	if cglob, err := strconv.Unquote(glob); err == nil {
		glob = cglob
	}

	key := glob + strconv.FormatInt(int64(spacing.Value), 10)
	// TODO: benchmark a single write lock against this
	// read lock then write lock
	ctx.Sprites.RLock()
	if _, ok := ctx.Sprites.M[key]; ok {
		ctx.Sprites.RUnlock()
		res, err := Marshal(key)
		if err != nil {
			return Error(err)
		}
		return res
	}
	ctx.Sprites.RUnlock()

	err = imgs.Decode(glob)
	if err != nil {
		return Error(err)
	}
	gpath, err := imgs.Combine()
	_ = gpath
	if err != nil {
		return Error(err)
	}

	_, err = imgs.Export()
	if err != nil {
		return Error(err)
	}

	res, err := Marshal(key)
	ctx.Sprites.Lock()
	ctx.Sprites.M[key] = imgs
	ctx.Sprites.Unlock()
	if err != nil {
		return Error(err)
	}

	return res
}

// SpriteFile proxies the sprite glob and image name through.
func FontURL(ctx *Context, usv UnionSassValue) UnionSassValue {

	var (
		path, format string
		csv          UnionSassValue
		raw          bool
	)
	err := Unmarshal(usv, &path, &raw)

	if err != nil {
		return Error(err)
	}

	// Enter warning
	if ctx.FontDir == "." {
		s := "font-url: font path not set"
		fmt.Println(s)
		res, _ := Marshal(s)
		return res
	}

	rel, err := filepath.Rel(ctx.BuildDir, ctx.FontDir)

	if err != nil {
		return Error(err)
	}
	if raw {
		format = "%s"
	} else {
		format = `url("%s")`
	}

	csv, err = Marshal(fmt.Sprintf(format, filepath.Join(rel, path)))
	if err != nil {
		return Error(err)
	}
	return csv
}
