package handlers

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/go-libsass/libs"
	sw "github.com/wellington/spritewell"
	"github.com/wellington/wellington/payload"
)

func init() {
	libsass.RegisterSassFunc("sprite($map, $name, $offsetX: 0px, $offsetY: 0px)", Sprite)
	libsass.RegisterSassFunc("sprite-map($glob, $spacing: 0px)", SpriteMap)
	libsass.RegisterSassFunc("sprite-file($map, $name)", SpriteFile)
	libsass.RegisterSassFunc("sprite-position($map, $file)", SpritePosition)
}

// SpritePosition returns the position of the image in the sprite-map.
// This is useful for passing directly to background-position
func SpritePosition(mainctx context.Context, usv libsass.SassValue) (rsv *libsass.SassValue, err error) {
	comp, err := libsass.CompFromCtx(mainctx)
	if err != nil {
		return
	}

	var glob, name string
	err = libsass.Unmarshal(usv, &glob, &name)
	if err != nil {
		return
	}

	loadctx := comp.Payload()
	if loadctx == nil {
		err = ErrPayloadNil
		return
	}

	sprites := payload.Sprite(loadctx)
	if sprites == nil {
		err = errors.New("Sprites missing")
		return
	}

	imgs := sprites.Get(glob)
	if imgs == nil {
		err = fmt.Errorf(
			"Variable not found matching glob: %s sprite:%s", glob, name)
		return
	}

	if imgs.Lookup(name) == -1 {
		err = fmt.Errorf("image %s not found\n"+
			"   try one of these: %v", name, imgs.Paths())
		return
	}

	// This is an odd name for what it does
	pos := imgs.GetPack(imgs.Lookup(name))
	if err != nil {
		return
	}

	x := libs.SassNumber{Unit: "px", Value: float64(-pos.X)}
	y := libs.SassNumber{Unit: "px", Value: float64(-pos.Y)}

	str, err := libsass.Marshal(
		[]libs.SassNumber{x, y},
	)
	return &str, err
}

// SpriteFile proxies the sprite glob and image name through.
func SpriteFile(ctx context.Context, usv libsass.SassValue) (rsv *libsass.SassValue, err error) {
	var glob, name string
	err = libsass.Unmarshal(usv, &glob, &name)
	if err != nil {
		return nil, err
	}
	infs := []interface{}{glob, name}
	res, err := libsass.Marshal(infs)
	return &res, err
}

// Sprite returns the source and background position for an image in the
// spritesheet.
func Sprite(ctx context.Context, usv libsass.SassValue) (rsv *libsass.SassValue, err error) {

	comp, err := libsass.CompFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	pather := comp.(libsass.Pather)

	var glob, name string
	var offsetX, offsetY libs.SassNumber
	err = libsass.Unmarshal(usv, &glob, &name, &offsetX, &offsetY)
	if err != nil {
		if err == libsass.ErrSassNumberNoUnit {
			err := fmt.Errorf(
				"Please specify unit for offset ie. (2px)")
			return nil, err
		}
		return nil, err
	}

	loadctx := comp.Payload()

	sprites := payload.Sprite(loadctx)

	imgs := sprites.Get(glob)
	if imgs == nil {
		err := fmt.Errorf(
			"Variable not found matching glob: %s sprite:%s", glob, name)
		return nil, err
	}

	path, err := imgs.OutputPath()
	if err != nil {
		return nil, err
	}

	buildDir := pather.BuildDir()
	genImgDir := pather.ImgBuildDir()
	httpPath := pather.HTTPPath()

	// FIXME: path directory can not be trusted, rebuild this from the context
	if len(httpPath) == 0 {
		ctxPath, err := filepath.Rel(buildDir, genImgDir)
		if err != nil {
			return nil, err
		}
		path = strings.Join([]string{ctxPath, filepath.Base(path)}, "/")
	} else {
		u, err := url.Parse(httpPath)
		if err != nil {
			return nil, err
		}
		u.Path = strings.Join([]string{u.Path, "build", filepath.Base(path)}, "/")
		path = u.String()
	}
	if err != nil {
		return nil, err
	}

	if imgs.Lookup(name) == -1 {
		return nil, fmt.Errorf("image %s not found\n"+
			"   try one of these: %v", name, imgs.Paths())
	}
	// This is an odd name for what it does
	pos := imgs.GetPack(imgs.Lookup(name))

	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &str, nil
}

// SpriteMap returns a sprite from the passed glob and sprite
// parameters.
func SpriteMap(mainctx context.Context, usv libsass.SassValue) (*libsass.SassValue, error) {
	var glob string
	var spacing libs.SassNumber
	err := libsass.Unmarshal(usv, &glob, &spacing)
	if err != nil {
		return nil, err
	}
	comp, err := libsass.CompFromCtx(mainctx)
	if err != nil {
		return nil, err
	}
	paths := comp.(libsass.Pather)
	imgs := sw.New(&sw.Options{
		ImageDir:  paths.ImgDir(),
		BuildDir:  paths.BuildDir(),
		GenImgDir: paths.ImgBuildDir(),
		Padding:   int(spacing.Value),
	})
	if cglob, err := strconv.Unquote(glob); err == nil {
		glob = cglob
	}

	key := glob + strconv.FormatInt(int64(spacing.Value), 10)

	loadctx := comp.Payload()
	sprites := payload.Sprite(loadctx)

	// FIXME: wtf is this?
	// sprites.RLock()
	// if _, ok := sprites.M[key]; ok {
	// 	defer sprites.RUnlock()
	// 	res, err := libsass.Marshal(key)
	// 	if err != nil {
	// 		return setErrorAndReturn(err, rsv)
	// 	}
	// 	if rsv != nil {
	// 		*rsv = res
	// 	}
	// 	return nil
	// }
	// sprites.RUnlock()

	err = imgs.Decode(glob)
	if err != nil {
		return nil, err
	}

	_, err = imgs.Export()
	if err != nil {
		return nil, err
	}

	res, err := libsass.Marshal(key)
	if err != nil {
		return nil, err
	}

	sprites.Set(key, imgs)

	return &res, nil
}
