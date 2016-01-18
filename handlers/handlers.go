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
	"strings"

	"golang.org/x/net/context"

	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/go-libsass/libs"
	sw "github.com/wellington/spritewell"
	"github.com/wellington/wellington/payload"
)

func init() {

	libsass.RegisterSassFunc("image-url($name)", ImageURL)
	libsass.RegisterSassFunc("image-height($path)", ImageHeight)
	libsass.RegisterSassFunc("image-width($path)", ImageWidth)
	libsass.RegisterHandler("inline-image($path, $encode: false)", InlineImage)
}

var ErrPayloadNil = errors.New("payload is nil")

// ImageURL handles calls to resolve the path to a local image from the
// built css file path.
func ImageURL(ctx context.Context, csv libsass.SassValue) (*libsass.SassValue, error) {
	comp, err := libsass.CompFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	pather := comp.(libsass.Pather)
	var path []string
	err = libsass.Unmarshal(csv, &path)
	// This should create and throw a sass error
	if err != nil {
		return nil, err
	}

	if len(path) != 1 {
		return nil, errors.New("path not found")
	}

	imgdir := pather.ImgDir()

	abspath := filepath.Join(imgdir, path[0])
	method := comp.CacheBust()

	qry, err := qs(method, abspath)
	if err != nil {
		return nil, err
	}
	rel, err := relativeImage(comp.BuildDir(), comp.ImgDir())
	if err != nil {
		return nil, err
	}
	url := strings.Join([]string{
		rel,
		path[0],
	}, "/")
	res, err := libsass.Marshal(fmt.Sprintf("url('%s%s')", url, qry))
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func relativeImage(buildDir, imageDir string) (string, error) {
	rel, err := filepath.Rel(buildDir, imageDir)
	return filepath.ToSlash(filepath.Clean(rel)), err
}

// ImageHeight takes a file path (or sprite glob) and returns the
// height in pixels of the image being referenced.
func ImageHeight(mainctx context.Context, usv libsass.SassValue) (*libsass.SassValue, error) {
	var (
		glob string
		name string
	)

	comp, err := libsass.CompFromCtx(mainctx)
	if err != nil {
		return nil, err
	}

	err = libsass.Unmarshal(usv, &name)
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
			return nil, err
		}
		glob = infs[0].(string)
		name = infs[1].(string)
	}
	imgs := sw.New(&sw.Options{
		ImageDir:  comp.ImgDir(),
		BuildDir:  comp.BuildDir(),
		GenImgDir: comp.GenImgDir(),
	})

	loadctx := comp.Payload()
	if loadctx == nil {
		return nil, ErrPayloadNil
	}

	images := payload.Image(loadctx)
	if images == nil {
		return nil, errors.New("inline payload not available")
	}

	if len(glob) == 0 {
		exst := images.Get(name)
		if exst != nil {
			imgs = exst
		} else {
			imgs.Decode(name)
			// Store images in global cache
			images.Set(name, imgs)
		}
	} else {
		sprites := payload.Sprite(loadctx)
		imgs = sprites.Get(glob)
		if imgs == nil {
			return nil, errors.New("Sprite not found")
		}
	}
	height := imgs.SImageHeight(name)
	Hheight := libs.SassNumber{
		Value: float64(height),
		Unit:  "px",
	}
	res, err := libsass.Marshal(Hheight)
	return &res, err
}

// ImageWidth takes a file path (or sprite glob) and returns the
// width in pixels of the image being referenced.
func ImageWidth(mainctx context.Context, usv libsass.SassValue) (rsv *libsass.SassValue, err error) {
	var (
		glob, name string
	)
	comp, err := libsass.CompFromCtx(mainctx)
	if err != nil {
		return
	}

	err = libsass.Unmarshal(usv, &name)
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
			return
		}
		glob = infs[0].(string)
		name = infs[1].(string)
	}
	imgs := sw.New(&sw.Options{
		ImageDir:  comp.ImgDir(),
		BuildDir:  comp.BuildDir(),
		GenImgDir: comp.GenImgDir(),
	})

	loadctx := comp.Payload()
	var images payload.Payloader

	if len(glob) == 0 {
		images = payload.Image(loadctx)
		hit := images.Get(name)
		if hit != nil {
			imgs = hit
		} else {
			imgs.Decode(name)
			images.Set(name, imgs)
		}
	} else {
		// Glob present, look up in sprites
		sprites := payload.Sprite(loadctx)
		imgs = sprites.Get(glob)
	}
	w := imgs.SImageWidth(name)
	ww := libs.SassNumber{
		Value: float64(w),
		Unit:  "px",
	}
	res, err := libsass.Marshal(ww)
	return &res, err
}

// Resolver interface returns a ReadCloser from a path
type Resolver interface {
	Do(string) (io.ReadCloser, error)
}

// img implements Resolver for http urls
type img struct {
}

func (g img) Do(name string) (io.ReadCloser, error) {
	// Check for valid URL
	u, err := url.Parse(name)
	if err != nil || len(u.Scheme) == 0 {
		return nil, fmt.Errorf("invalid image URL: %s", name)
	}

	// No point in check this error, we already are more aggressive
	// about error checking by validating scheme exists
	req, _ := http.NewRequest("GET", name, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		resp.Body.Close()
		return nil, fmt.Errorf("could not resolve image: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

var imgResolver Resolver = img{}

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

	// check for valid URL. If true, attempt to resolve image.
	// This is really going to slow down compilation think about
	// writing data to disk instead of inlining.
	u, err := url.Parse(name)
	if err == nil && len(u.Scheme) > 0 {
		f, err = imgResolver.Do(u.String())
	} else {
		f, err = os.Open(filepath.Join(ctx.ImageDir, name))
	}
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	defer f.Close()

	var buf bytes.Buffer
	err = sw.Inline(f, &buf, encode)
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}
	res, err := libsass.Marshal(buf.String())
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
