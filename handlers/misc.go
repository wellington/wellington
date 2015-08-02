package handlers

import (
	"errors"
	"fmt"
	"path/filepath"

	libsass "github.com/wellington/go-libsass"
)

func init() {
	libsass.RegisterHandler("font-url($path, $raw: false)", FontURL)
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
