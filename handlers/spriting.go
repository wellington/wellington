package handlers

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"

	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/go-libsass/libs"
	sw "github.com/wellington/spritewell"
)

func init() {
	libsass.RegisterHandler("sprite($map, $name, $offsetX: 0px, $offsetY: 0px)", Sprite)
	libsass.RegisterHandler("sprite-map($glob, $spacing: 0px)", SpriteMap)
	libsass.RegisterHandler("sprite-file($map, $name)", SpriteFile)
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
	err := libsass.Unmarshal(usv, &glob, &name, &offsetX, &offsetY)
	if err != nil {
		if err == libsass.ErrSassNumberNoUnit {
			err := fmt.Errorf(
				"Please specify unit for offset ie. (2px)")
			return setErrorAndReturn(err, rsv)
		}
		return setErrorAndReturn(err, rsv)
	}
	sprites := ctx.Sprites
	sprites.RLock()
	defer sprites.RUnlock()
	imgs, ok := sprites.M[glob]
	if !ok {
		keys := make([]string, 0, len(sprites.M))
		for i := range sprites.M {
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

	x := libs.SassNumber{Unit: "px", Value: float64(-pos.X)}
	x = x.Add(offsetX)

	y := libs.SassNumber{Unit: "px", Value: float64(-pos.Y)}
	y = y.Add(offsetY)

	str, err := libsass.Marshal(
		fmt.Sprintf(`url("%s") %s %s`,
			path, x, y,
		))
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

	payload, ok := ctx.Payload.(sw.Spriter)
	if !ok {
		err := errors.New("context payload not found")
		return setErrorAndReturn(err, rsv)
	}

	sprites := payload.Sprite()

	// TODO: benchmark a single write lock against this
	// read lock then write lock
	sprites.RLock()
	if _, ok := sprites.M[key]; ok {
		res, err := libsass.Marshal(key)
		if err != nil {
			return setErrorAndReturn(err, rsv)
		}
		if rsv != nil {
			*rsv = res
		}
		return nil
	}
	sprites.RUnlock()

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
	sprites.Lock()
	sprites.M[key] = imgs
	sprites.Unlock()
	if err != nil {
		return setErrorAndReturn(err, rsv)
	}

	if rsv != nil {
		*rsv = res
	}
	return nil
}
