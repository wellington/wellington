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
)

func init() {

	libsass.RegisterSassFunc("image-url($name)", ImageURL)
	libsass.RegisterHandler("image-height($path)", ImageHeight)
	libsass.RegisterHandler("image-width($path)", ImageWidth)
	libsass.RegisterHandler("inline-image($path, $encode: false)", InlineImage)
}

// ImageURL handles calls to resolve the path to a local image from the
// built css file path.
func ImageURL(ctx context.Context, csv libsass.SassValue) (*libsass.SassValue, error) {
	comp, err := libsass.CompFromCtx(ctx)
	if err != nil {
		return nil, libsass.ErrCompilerNotFound
	}
	libctx := comp.Context()
	var path []string
	err = libsass.Unmarshal(csv, &path)
	// This should create and throw a sass error
	if err != nil {
		return nil, err
	}

	if len(path) != 1 {
		return nil, errors.New("path not found")
	}

	imgdir := comp.ImgDir()

	abspath := filepath.Join(imgdir, path[0])
	method := comp.CacheBust()

	qry, err := qs(method, abspath)
	if err != nil {
		return nil, err
	}

	url := strings.Join([]string{
		libctx.RelativeImage(),
		path[0],
	}, "/")
	res, err := libsass.Marshal(fmt.Sprintf("url('%s%s')", url, qry))
	if err != nil {
		return nil, err
	}
	return &res, nil
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
	imgs := sw.New(&sw.Options{
		ImageDir:  ctx.ImageDir,
		BuildDir:  ctx.BuildDir,
		GenImgDir: ctx.GenImgDir,
	})

	payload, ok := ctx.Payload.(sw.Imager)
	if !ok {
		return setErrorAndReturn(errors.New("inline payload not available"), rsv)
	}
	images := payload.Image()
	if glob == "" {
		images.RLock()
		hit, ok := images.M[name]
		images.RUnlock()
		if ok {
			imgs = hit
		} else {
			imgs.Decode(name)
			images.Lock()
			images.M[name] = imgs
			images.Unlock()
		}
	} else {
		payload, ok := ctx.Payload.(sw.Spriter)
		if !ok {
			return setErrorAndReturn(errors.New("Context payload not found"), rsv)
		}
		sprites := payload.Sprite()

		sprites.RLock()
		imgs = sprites.M[glob]
		sprites.RUnlock()
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
	imgs := sw.New(&sw.Options{
		ImageDir:  ctx.ImageDir,
		BuildDir:  ctx.BuildDir,
		GenImgDir: ctx.GenImgDir,
	})

	payload, ok := ctx.Payload.(sw.Imager)
	if !ok {
		return setErrorAndReturn(errors.New("inline payload not available"), rsv)
	}
	images := payload.Image()

	if glob == "" {
		images.RLock()
		hit, ok := images.M[name]
		images.RUnlock()
		if ok {
			imgs = hit
		} else {
			imgs.Decode(name)
			images.Lock()
			images.M[name] = imgs
			images.Unlock()
		}
	} else {
		payload, ok := ctx.Payload.(sw.Spriter)
		if !ok {
			return setErrorAndReturn(errors.New("Context payload not found"), rsv)
		}
		sprites := payload.Sprite()
		sprites.RLock()
		imgs = sprites.M[glob]
		sprites.RUnlock()
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
