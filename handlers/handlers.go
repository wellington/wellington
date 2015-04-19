package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	libsass "github.com/wellington/libsass"
	sw "github.com/wellington/spritewell"
)

func init() {

	libsass.RegisterHandler("sprite-map($glob, $spacing: 0px)", SpriteMap)
	libsass.RegisterHandler("sprite-file($map, $name)", SpriteFile)
	libsass.RegisterHandler("image-url($name)", ImageURL)
	libsass.RegisterHandler("image-height($path)", ImageHeight)
	libsass.RegisterHandler("image-width($path)", ImageWidth)
	libsass.RegisterHandler("inline-image($path, $encode: false)", InlineImage)
	libsass.RegisterHandler("font-url($path, $raw: false)", FontURL)
	libsass.RegisterHandler("sprite($map, $name, $offsetX: 0px, $offsetY: 0px)", Sprite)
}

// ImageURL handles calls to resolve a local image from the
// built css file path.
func ImageURL(ctx *libsass.Context, csv libsass.UnionSassValue) libsass.UnionSassValue {
	var path []string
	err := libsass.Unmarshal(csv, &path)
	// This should create and throw a sass error
	if err != nil {
		return libsass.Error(err)
	}
	url := filepath.Join(ctx.RelativeImage(), path[0])
	res, err := libsass.Marshal(fmt.Sprintf("url('%s')", url))
	if err != nil {
		return libsass.Error(err)
	}
	return res
}

// ImageHeight takes a file path (or sprite glob) and returns the
// height in pixels of the image being referenced.
func ImageHeight(ctx *libsass.Context, usv libsass.UnionSassValue) libsass.UnionSassValue {
	var (
		glob string
		name string
	)
	err := libsass.Unmarshal(usv, &name)
	// Check for sprite-file override first
	if err != nil {
		var inf interface{}
		var infs []interface{}
		// Can't unmarshal to []interface{}, so unmarshal to
		// interface{} then reflect it into a []interface{}
		err = libsass.Unmarshal(usv, &inf)
		k := reflect.ValueOf(&infs).Elem()
		k.Set(reflect.ValueOf(inf))

		if err != nil {
			return libsass.Error(err)
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
	Hheight := libsass.SassNumber{
		Value: float64(height),
		Unit:  "px",
	}
	res, err := libsass.Marshal(Hheight)
	if err != nil {
		fmt.Println(err)
	}
	return res
}

// ImageWidth takes a file path (or sprite glob) and returns the
// width in pixels of the image being referenced.
func ImageWidth(ctx *libsass.Context, usv libsass.UnionSassValue) libsass.UnionSassValue {
	var (
		glob, name string
	)
	err := libsass.Unmarshal(usv, &name)
	// Check for sprite-file override first
	if err != nil {
		var inf interface{}
		var infs []interface{}
		// Can't unmarshal to []interface{}, so unmarshal to
		// interface{} then reflect it into a []interface{}
		err = libsass.Unmarshal(usv, &inf)
		k := reflect.ValueOf(&infs).Elem()
		k.Set(reflect.ValueOf(inf))

		if err != nil {
			return libsass.Error(err)
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
	vv := libsass.SassNumber{
		Value: float64(v),
		Unit:  "px",
	}
	res, err := libsass.Marshal(vv)
	if err != nil {
		fmt.Println(err)
	}
	return res
}

func inlineHandler(name string) (*http.Request, error) {

	u, err := url.Parse(name)
	if err != nil || u.Scheme == "" {
		return nil, err
	}

	req, err := http.NewRequest("GET", name, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// InlineImage returns a base64 encoded png from the input image
func InlineImage(ctx *libsass.Context, usv libsass.UnionSassValue) libsass.UnionSassValue {
	var (
		name   string
		encode bool
		f      io.Reader
	)
	err := libsass.Unmarshal(usv, &name, &encode)
	if err != nil {
		return libsass.Error(err)
	}

	f, err = os.Open(filepath.Join(ctx.ImageDir, name))
	if err != nil {
		req, uerr := inlineHandler(name)
		if uerr != nil || req == nil {
			return libsass.Error(err)
		}
		client := &http.Client{}
		resp, uerr := client.Do(req)
		if uerr != nil {
			return libsass.Error(err)
		}
		defer resp.Body.Close()

		if uerr == nil && f != nil {
			err = uerr
		}
		f = resp.Body
	}

	if err != nil {
		return libsass.Error(err)
	}
	var buf bytes.Buffer

	sw.Inline(f, &buf, encode)
	res, err := libsass.Marshal(buf.String())
	if err != nil {
		return libsass.Error(err)
	}
	return res
}

// SpriteFile proxies the sprite glob and image name through.
func SpriteFile(ctx *libsass.Context, usv libsass.UnionSassValue) libsass.UnionSassValue {
	var glob, name string
	err := libsass.Unmarshal(usv, &glob, &name)
	if err != nil {
		return libsass.Error(err)
	}
	infs := []interface{}{glob, name}
	res, err := libsass.Marshal(infs)
	return res
}

// Sprite returns the source and background position for an image in the
// spritesheet.
func Sprite(ctx *libsass.Context, usv libsass.UnionSassValue) libsass.UnionSassValue {
	var glob, name string
	var offsetX, offsetY libsass.SassNumber
	_, _ = offsetX, offsetY // TODO: ignore these for now
	err := libsass.Unmarshal(usv, &glob, &name, &offsetX, &offsetY)
	if err != nil {
		if strings.Contains(err.Error(), "unsupported") {
			return libsass.Error(fmt.Errorf(
				"Please specify unit for offset ie. (2px)"))
		}
		return libsass.Error(err)
	}
	ctx.Sprites.RLock()
	defer ctx.Sprites.RUnlock()
	imgs, ok := ctx.Sprites.M[glob]
	if !ok {
		keys := make([]string, 0, len(ctx.Sprites.M))
		for i := range ctx.Sprites.M {
			keys = append(keys, i)
		}

		return libsass.Error(fmt.Errorf(
			"Variable not found matching glob: %s sprite:%s", glob, name))
	}

	path, err := imgs.OutputPath()

	// FIXME: path directory can not be trusted, rebuild this from the context
	if ctx.HTTPPath == "" {
		ctxPath, _ := filepath.Rel(ctx.BuildDir, ctx.GenImgDir)
		path = filepath.Join(ctxPath, filepath.Base(path))
	} else {
		u, err := url.Parse(ctx.HTTPPath)
		if err != nil {
			return libsass.Error(err)
		}
		u.Path = filepath.Join(u.Path, "build", filepath.Base(path))
		path = u.String()
	}
	if err != nil {
		return libsass.Error(err)
	}

	if imgs.Lookup(name) == -1 {
		return libsass.Error(fmt.Errorf("image %s not found\n"+
			"   try one of these: %v", name, imgs.Paths))
	}
	// This is an odd name for what it does
	pos := imgs.GetPack(imgs.Lookup(name))

	if err != nil {
		return libsass.Error(err)
	}
	str, err := libsass.Marshal(fmt.Sprintf(`url("%s") -%dpx -%dpx`,
		path, pos.X, pos.Y))
	if err != nil {
		return libsass.Error(err)
	}
	return str
}

// SpriteMap returns a sprite from the passed glob and sprite
// parameters.
func SpriteMap(ctx *libsass.Context, usv libsass.UnionSassValue) libsass.UnionSassValue {
	var glob string
	var spacing libsass.SassNumber
	err := libsass.Unmarshal(usv, &glob, &spacing)
	if err != nil {
		return libsass.Error(err)
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
		res, err := libsass.Marshal(key)
		if err != nil {
			return libsass.Error(err)
		}
		return res
	}
	ctx.Sprites.RUnlock()

	err = imgs.Decode(glob)
	if err != nil {
		return libsass.Error(err)
	}
	_, err = imgs.Combine()
	if err != nil {
		return libsass.Error(err)
	}

	_, err = imgs.Export()
	if err != nil {
		return libsass.Error(err)
	}

	res, err := libsass.Marshal(key)
	ctx.Sprites.Lock()
	ctx.Sprites.M[key] = imgs
	ctx.Sprites.Unlock()
	if err != nil {
		return libsass.Error(err)
	}

	return res
}

// FontURL builds a relative path to the requested font file from the built CSS.
func FontURL(ctx *libsass.Context, usv libsass.UnionSassValue) libsass.UnionSassValue {

	var (
		path, format string
		csv          libsass.UnionSassValue
		raw          bool
	)
	err := libsass.Unmarshal(usv, &path, &raw)

	if err != nil {
		return libsass.Error(err)
	}

	// Enter warning
	if ctx.FontDir == "." || ctx.FontDir == "" {
		s := "font-url: font path not set"
		fmt.Println(s)
		res, _ := libsass.Marshal(s)
		return res
	}

	rel, err := filepath.Rel(ctx.BuildDir, ctx.FontDir)

	if err != nil {
		return libsass.Error(err)
	}
	if raw {
		format = "%s"
	} else {
		format = `url("%s")`
	}

	csv, err = libsass.Marshal(fmt.Sprintf(format, filepath.Join(rel, path)))
	if err != nil {
		return libsass.Error(err)
	}
	return csv
}
