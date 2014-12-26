package handlers

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	sw "github.com/wellington/spritewell"
	cx "github.com/wellington/wellington/context"
)

func init() {

	cx.RegisterHandler("sprite-map($glob, $spacing: 0px)", SpriteMap)
	cx.RegisterHandler("sprite-file($map, $name)", SpriteFile)
	cx.RegisterHandler("image-url($name)", ImageURL)
	cx.RegisterHandler("image-height($path)", ImageHeight)
	cx.RegisterHandler("image-width($path)", ImageWidth)
	cx.RegisterHandler("inline-image($path, $encode: false)", InlineImage)
	cx.RegisterHandler("font-url($path, $raw: false)", FontURL)
	cx.RegisterHandler("sprite($map, $name, $offsetX: 0px, $offsetY: 0px)", Sprite)
}

// ImageURL handles calls to resolve a local image from the
// built css file path.
func ImageURL(ctx *cx.Context, csv cx.UnionSassValue) cx.UnionSassValue {
	var path []string
	err := cx.Unmarshal(csv, &path)
	// This should create and throw a sass error
	if err != nil {
		return cx.Error(err)
	}
	url := filepath.Join(ctx.RelativeImage(), path[0])
	res, err := cx.Marshal(fmt.Sprintf("url('%s')", url))
	if err != nil {
		return cx.Error(err)
	}
	return res
}

// ImageHeight takes a file path (or sprite glob) and returns the
// height in pixels of the image being referenced.
func ImageHeight(ctx *cx.Context, usv cx.UnionSassValue) cx.UnionSassValue {
	var (
		glob string
		name string
	)
	err := cx.Unmarshal(usv, &name)
	// Check for sprite-file override first
	if err != nil {
		var inf interface{}
		var infs []interface{}
		// Can't unmarshal to []interface{}, so unmarshal to
		// interface{} then reflect it into a []interface{}
		err = cx.Unmarshal(usv, &inf)
		k := reflect.ValueOf(&infs).Elem()
		k.Set(reflect.ValueOf(inf))

		if err != nil {
			return cx.Error(err)
		}
		glob = infs[0].(string)
		name = infs[1].(string)
	}
	imgs := sw.ImageList{
		ImageDir:  ctx.ImageDir,
		BuildDir:  ctx.BuildDir,
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
	Hheight := cx.SassNumber{
		Value: float64(height),
		Unit:  "px",
	}
	res, err := cx.Marshal(Hheight)
	if err != nil {
		fmt.Println(err)
	}
	return res
}

// ImageWidth takes a file path (or sprite glob) and returns the
// width in pixels of the image being referenced.
func ImageWidth(ctx *cx.Context, usv cx.UnionSassValue) cx.UnionSassValue {
	var (
		glob, name string
	)
	err := cx.Unmarshal(usv, &name)
	// Check for sprite-file override first
	if err != nil {
		var inf interface{}
		var infs []interface{}
		// Can't unmarshal to []interface{}, so unmarshal to
		// interface{} then reflect it into a []interface{}
		err = cx.Unmarshal(usv, &inf)
		k := reflect.ValueOf(&infs).Elem()
		k.Set(reflect.ValueOf(inf))

		if err != nil {
			return cx.Error(err)
		}
		glob = infs[0].(string)
		name = infs[1].(string)
	}
	imgs := sw.ImageList{
		ImageDir:  ctx.ImageDir,
		BuildDir:  ctx.BuildDir,
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
	vv := cx.SassNumber{
		Value: float64(v),
		Unit:  "px",
	}
	res, err := cx.Marshal(vv)
	if err != nil {
		fmt.Println(err)
	}
	return res
}

// InlineImage returns a base64 encoded png from the input image
func InlineImage(ctx *cx.Context, usv cx.UnionSassValue) cx.UnionSassValue {
	var (
		name    string
		encoded bool
	)
	err := cx.Unmarshal(usv, &name, &encoded)
	if err != nil {
		return cx.Error(err)
	}

	if !sw.CanDecode(filepath.Ext(name)) {
		// Special fallthrough for svg
		if filepath.Ext(name) == ".svg" {
			fin, err := ioutil.ReadFile(filepath.Join(ctx.ImageDir, name))
			if err != nil {
				return cx.Error(err)
			}
			var out []byte
			if encoded {
				out = sw.InlineSVGBase64(fin)
			} else {
				out = sw.InlineSVG(fin)
			}
			res, err := cx.Marshal(string(out))
			if err != nil {
				return cx.Error(err)
			}
			return res
		}
		s := fmt.Sprintf("inline-image: %s filetype %s is not supported",
			name, filepath.Ext(name))
		fmt.Println(s)
		// TODO: Replace with warning
		res, _ := cx.Marshal(s)
		return res
	}

	imgs := sw.ImageList{
		ImageDir:  ctx.ImageDir,
		BuildDir:  ctx.BuildDir,
		GenImgDir: ctx.GenImgDir,
	}
	err = imgs.Decode(name)
	if err != nil {
		return cx.Error(err)
	}
	_, err = imgs.Combine()
	if err != nil {
		fmt.Println(err)
	}
	str := imgs.Inline()
	res, err := cx.Marshal(str)
	if err != nil {
		return cx.Error(err)
	}
	return res
}

// SpriteFile proxies the sprite glob and image name through.
func SpriteFile(ctx *cx.Context, usv cx.UnionSassValue) cx.UnionSassValue {
	var glob, name string
	err := cx.Unmarshal(usv, &glob, &name)
	if err != nil {
		return cx.Error(err)
	}
	infs := []interface{}{glob, name}
	res, err := cx.Marshal(infs)
	return res
}

// Sprite returns the source and background position for an image in the
// spritesheet.
func Sprite(ctx *cx.Context, usv cx.UnionSassValue) cx.UnionSassValue {
	var glob, name string
	var offsetX, offsetY cx.SassNumber
	_, _ = offsetX, offsetY // TODO: ignore these for now
	err := cx.Unmarshal(usv, &glob, &name, &offsetX, &offsetY)
	if err != nil {
		if strings.Contains(err.Error(), "unsupported") {
			return cx.Error(fmt.Errorf(
				"Please specify unit for offset ie. (2px)"))
		}
		return cx.Error(err)
	}
	ctx.Sprites.RLock()
	defer ctx.Sprites.RUnlock()
	imgs, ok := ctx.Sprites.M[glob]
	if !ok {
		keys := make([]string, 0, len(ctx.Sprites.M))
		for i := range ctx.Sprites.M {
			keys = append(keys, i)
		}

		return cx.Error(fmt.Errorf(
			"Variable not found matching glob: %s sprite:%s", glob, name))
	}

	path, err := imgs.OutputPath()
	if err != nil {
		return cx.Error(err)
	}

	if imgs.Lookup(name) == -1 {
		return cx.Error(fmt.Errorf("image %s not found\n"+
			"   try one of these: %v", name, imgs.Paths))
	}
	// This is an odd name for what it does
	pos := imgs.GetPack(imgs.Lookup(name))

	if err != nil {
		return cx.Error(err)
	}
	str, err := cx.Marshal(fmt.Sprintf(`url("%s") -%dpx -%dpx`,
		path, pos.X, pos.Y))
	if err != nil {
		return cx.Error(err)
	}
	return str
}

// SpriteMap returns a sprite from the passed glob and sprite
// parameters.
func SpriteMap(ctx *cx.Context, usv cx.UnionSassValue) cx.UnionSassValue {
	var glob string
	var spacing cx.SassNumber
	err := cx.Unmarshal(usv, &glob, &spacing)
	if err != nil {
		return cx.Error(err)
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
		res, err := cx.Marshal(key)
		if err != nil {
			return cx.Error(err)
		}
		return res
	}
	ctx.Sprites.RUnlock()

	err = imgs.Decode(glob)
	if err != nil {
		return cx.Error(err)
	}
	gpath, err := imgs.Combine()
	_ = gpath
	if err != nil {
		return cx.Error(err)
	}

	_, err = imgs.Export()
	if err != nil {
		return cx.Error(err)
	}

	res, err := cx.Marshal(key)
	ctx.Sprites.Lock()
	ctx.Sprites.M[key] = imgs
	ctx.Sprites.Unlock()
	if err != nil {
		return cx.Error(err)
	}

	return res
}

// FontURL builds a relative path to the requested font file from the built CSS.
func FontURL(ctx *cx.Context, usv cx.UnionSassValue) cx.UnionSassValue {

	var (
		path, format string
		csv          cx.UnionSassValue
		raw          bool
	)
	err := cx.Unmarshal(usv, &path, &raw)

	if err != nil {
		return cx.Error(err)
	}

	// Enter warning
	if ctx.FontDir == "." || ctx.FontDir == "" {
		s := "font-url: font path not set"
		fmt.Println(s)
		res, _ := cx.Marshal(s)
		return res
	}

	rel, err := filepath.Rel(ctx.BuildDir, ctx.FontDir)

	if err != nil {
		return cx.Error(err)
	}
	if raw {
		format = "%s"
	} else {
		format = `url("%s")`
	}

	csv, err = cx.Marshal(fmt.Sprintf(format, filepath.Join(rel, path)))
	if err != nil {
		return cx.Error(err)
	}
	return csv
}
