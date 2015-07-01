package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/go-libsass/libs"
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
func ImageURL(v interface{}, csv libsass.SassValue, rsv *libsass.SassValue) error {
	ctx := v.(*libsass.Context)
	var path []string
	err := libsass.Unmarshal(csv, &path)
	// This should create and throw a sass error
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	url := filepath.Join(ctx.RelativeImage(), path[0])
	res, err := libsass.Marshal(fmt.Sprintf("url('%s')", url))
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	if rsv != nil {
		*rsv = res
	}
	return nil
}

// ImageHeight takes a file path (or sprite glob) and returns the
// height in pixels of the image being referenced.
func ImageHeight(v interface{}, usv libsass.SassValue, rsv *libsass.SassValue) error {
	var (
		glob string
		name string
	)
	ctx := v.(*libsass.Context)
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
			return setErrorAndReturn(err, rsv)
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
	Hheight := libs.SassNumber{
		Value: float64(height),
		Unit:  "px",
	}
	res, err := libsass.Marshal(Hheight)
	if err != nil {
		fmt.Println(err)
	}
	if rsv != nil {
		*rsv = res
	}
	return nil
}

// ImageWidth takes a file path (or sprite glob) and returns the
// width in pixels of the image being referenced.
func ImageWidth(v interface{}, usv libsass.SassValue, rsv *libsass.SassValue) error {
	var (
		glob, name string
	)
	ctx := v.(*libsass.Context)
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
			return setErrorAndReturn(err, rsv)
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
	w := imgs.SImageWidth(name)
	ww := libs.SassNumber{
		Value: float64(w),
		Unit:  "px",
	}
	res, err := libsass.Marshal(ww)
	if err != nil {
		fmt.Println(err)
	}
	if rsv != nil {
		*rsv = res
	}
	return err
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

func httpInlineImage(url string) (io.ReadCloser, error) {
	req, err := inlineHandler(url)
	if err != nil || req == nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println(url)
	fmt.Printf("% #v\n", err)
	if err != nil {
		fmt.Println("errors")
		return nil, err
	}

	return resp.Body, nil
}

// InlineImage returns a base64 encoded png from the input image
func InlineImage(v interface{}, usv libsass.SassValue, rsv *libsass.SassValue) error {
	var (
		name   string
		encode bool
		f      io.ReadCloser
	)
	ctx := v.(*libsass.Context)
	err := libsass.Unmarshal(usv, &name, &encode)
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}

	f, err = os.Open(filepath.Join(ctx.ImageDir, name))
	defer f.Close()
	if err != nil {
		r, err := httpInlineImage(name)
		if err != nil {
			return setErrorAndReturn(err, rsv)
		}
		f = r
		if r != nil {
			defer r.Close()
		}
	}

	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	var buf bytes.Buffer

	sw.Inline(f, &buf, encode)
	res, err := libsass.Marshal(buf.String())
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	if rsv != nil {
		*rsv = res
	}
	return nil
}

// SpriteFile proxies the sprite glob and image name through.

func SpriteFile(v interface{}, usv libsass.SassValue, rsv *libsass.SassValue) error {
	var glob, name string
	err := libsass.Unmarshal(usv, &glob, &name)
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	infs := []interface{}{glob, name}
	res, err := libsass.Marshal(infs)
	if rsv != nil {
		*rsv = res
	}
	return nil
}

// Sprite returns the source and background position for an image in the
// spritesheet.
func Sprite(v interface{}, usv libsass.SassValue, rsv *libsass.SassValue) error {
	ctx := v.(*libsass.Context)
	var glob, name string
	var offsetX, offsetY libs.SassNumber
	_, _ = offsetX, offsetY // TODO: ignore these for now
	err := libsass.Unmarshal(usv, &glob, &name, &offsetX, &offsetY)
	if err != nil {
		if err == libsass.ErrSassNumberNoUnit {
			err := fmt.Errorf(
				"Please specify unit for offset ie. (2px)")
			return setErrorAndReturn(err, rsv)
		}
		return setErrorAndReturn(err, rsv)
	}
	ctx.Sprites.RLock()
	defer ctx.Sprites.RUnlock()
	imgs, ok := ctx.Sprites.M[glob]
	if !ok {
		keys := make([]string, 0, len(ctx.Sprites.M))
		for i := range ctx.Sprites.M {
			keys = append(keys, i)
		}

		err := fmt.Errorf(
			"Variable not found matching glob: %s sprite:%s", glob, name)
		return setErrorAndReturn(err, rsv)
	}

	path, err := imgs.OutputPath()

	// FIXME: path directory can not be trusted, rebuild this from the context
	if ctx.HTTPPath == "" {
		ctxPath, _ := filepath.Rel(ctx.BuildDir, ctx.GenImgDir)
		path = filepath.Join(ctxPath, filepath.Base(path))
	} else {
		u, err := url.Parse(ctx.HTTPPath)
		if err != nil {
			return setErrorAndReturn(err, rsv)
		}
		u.Path = filepath.Join(u.Path, "build", filepath.Base(path))
		path = u.String()
	}
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}

	if imgs.Lookup(name) == -1 {
		return setErrorAndReturn(fmt.Errorf("image %s not found\n"+
			"   try one of these: %v", name, imgs.Paths), rsv)
	}
	// This is an odd name for what it does
	pos := imgs.GetPack(imgs.Lookup(name))

	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	str, err := libsass.Marshal(fmt.Sprintf(`url("%s") -%dpx -%dpx`,
		path, pos.X, pos.Y))
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	if rsv != nil {
		*rsv = str
	}
	return nil
}

// SpriteMap returns a sprite from the passed glob and sprite
// parameters.
func SpriteMap(v interface{}, usv libsass.SassValue, rsv *libsass.SassValue) error {
	ctx := v.(*libsass.Context)
	var glob string
	var spacing libs.SassNumber
	err := libsass.Unmarshal(usv, &glob, &spacing)
	if err != nil {
		return setErrorAndReturn(err, rsv)
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
			return setErrorAndReturn(err, rsv)
		}
		if rsv != nil {
			*rsv = res
		}
		return nil
	}
	ctx.Sprites.RUnlock()

	err = imgs.Decode(glob)
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	_, err = imgs.Combine()
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}

	_, err = imgs.Export()
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}

	res, err := libsass.Marshal(key)
	ctx.Sprites.Lock()
	ctx.Sprites.M[key] = imgs
	ctx.Sprites.Unlock()
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}

	if rsv != nil {
		*rsv = res
	}
	return nil
}

func setErrorAndReturn(err error, rsv *libsass.SassValue) error {
	if rsv == nil {
		panic("rsv not initialized")
	}
	*rsv = libsass.Error(err)
	return err
}

// FontURL builds a relative path to the requested font file from the built CSS.
func FontURL(v interface{}, usv libsass.SassValue, rsv *libsass.SassValue) error {

	var (
		path, format string
		csv          libsass.SassValue
		raw          bool
	)
	ctx := v.(*libsass.Context)
	err := libsass.Unmarshal(usv, &path, &raw)

	if err != nil {
		return setErrorAndReturn(err, rsv)
	}

	// Enter warning
	if ctx.FontDir == "." || ctx.FontDir == "" {
		s := "font-url: font path not set"
		return setErrorAndReturn(errors.New(s), rsv)
	}

	rel, err := filepath.Rel(ctx.BuildDir, ctx.FontDir)

	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	if raw {
		format = "%s"
	} else {
		format = `url("%s")`
	}

	csv, err = libsass.Marshal(fmt.Sprintf(format, filepath.Join(rel, path)))
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	if rsv != nil {
		*rsv = csv
	}
	return nil
}
