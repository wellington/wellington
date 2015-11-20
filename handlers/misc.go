package handlers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/net/context"

	libsass "github.com/wellington/go-libsass"
)

func init() {
	libsass.RegisterSassFunc("font-url($path, $raw: false)", FontURL)
}

// FontURL builds a relative path to the requested font file from the built CSS.
func FontURL(ctx context.Context, usv libsass.SassValue) (*libsass.SassValue, error) {

	var (
		path, format string
		csv          libsass.SassValue
		raw          bool
	)
	comp, err := libsass.CompFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	libctx := comp.Context()
	err = libsass.Unmarshal(usv, &path, &raw)

	if err != nil {
		return nil, err
	}

	// Enter warning
	if libctx.FontDir == "." || libctx.FontDir == "" {
		s := "font-url: font path not set"
		return nil, errors.New(s)
	}

	rel, err := filepath.Rel(libctx.BuildDir, libctx.FontDir)
	if err != nil {
		return nil, err
	}

	if raw {
		format = "%s%s"
	} else {
		format = `url("%s%s")`
	}
	// TODO: FontDir() on compiler
	fontdir := libctx.FontDir
	var qs string
	if comp.CacheBust() {
		fileinfo, err := os.Stat(filepath.Join(fontdir, path))
		if err != nil {
			return nil, err
		}
		qs, err = modHash(fileinfo)
		if err != nil {
			return nil, err
		}
	}
	csv, err = libsass.Marshal(fmt.Sprintf(format,
		filepath.ToSlash(filepath.Join(rel, path)),
		qs,
	))

	return &csv, err
}
