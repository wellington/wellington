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

	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/go-libsass/libs"
	sw "github.com/wellington/spritewell"
)

func init() {

	libsass.RegisterHandler("image-url($name)", ImageURL)
	libsass.RegisterHandler("image-height($path)", ImageHeight)
	libsass.RegisterHandler("image-width($path)", ImageWidth)
	libsass.RegisterHandler("inline-image($path, $encode: false)", InlineImage)
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
	url := strings.Join([]string{ctx.RelativeImage(), path[0]}, "/")
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

	payload, ok := ctx.Payload.(sw.Imager)
	if !ok {
		return setErrorAndReturn(errors.New("inline payload not available"), rsv)
	}
	images := payload.Image()
	if glob == "" {
		if hit, ok := images.M[name]; ok {
			imgs = hit
		} else {
			imgs.Decode(name)
			imgs.Combine()
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
	imgs := sw.ImageList{
		ImageDir:  ctx.ImageDir,
		BuildDir:  ctx.BuildDir,
		GenImgDir: ctx.GenImgDir,
	}

	payload, ok := ctx.Payload.(sw.Imager)
	if !ok {
		return setErrorAndReturn(errors.New("inline payload not available"), rsv)
	}
	images := payload.Image()

	if glob == "" {
		if hit, ok := images.M[name]; ok {
			imgs = hit
		} else {
			imgs.Decode(name)
			imgs.Combine()
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

func setErrorAndReturn(err error, rsv *libsass.SassValue) error {
	if rsv == nil {
		panic("rsv not initialized")
	}
	*rsv = libsass.Error(err)
	return err
}
