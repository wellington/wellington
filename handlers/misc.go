package handlers

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
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

	err = libsass.Unmarshal(usv, &path, &raw)

	if err != nil {
		return nil, err
	}
	paths := comp.(libsass.Pather)
	fdir := paths.FontDir()
	// Enter warning
	if fdir == "." || fdir == "" {
		s := "font-url: font path not set"
		return nil, errors.New(s)
	}

	rel, err := filepath.Rel(paths.BuildDir(), fdir)
	if err != nil {
		return nil, err
	}

	if raw {
		format = "%s%s"
	} else {
		format = `url("%s%s")`
	}

	abspath := filepath.Join(fdir, path)
	qry, err := qs(comp.CacheBust(), abspath)
	if err != nil {
		return nil, err
	}
	csv, err = libsass.Marshal(fmt.Sprintf(format,
		filepath.ToSlash(filepath.Join(rel, path)),
		qry,
	))

	return &csv, err
}

func sumHash(f io.ReadCloser) (string, error) {
	defer f.Close()
	hdr := make([]byte, 10*1024)
	n, err := f.Read(hdr)
	// File is empty, which is valid...
	if n == 0 {
		return "?empty", nil
	}
	if err != nil {
		return "", err
	}
	h := crc32.NewIEEE()
	_, err = h.Write(hdr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("?%x", h.Sum(nil)), nil
}

func modHash(info os.FileInfo) (string, error) {
	mod := info.ModTime()
	bs, err := mod.MarshalText()
	if err != nil {
		return "", err
	}
	ts := sha1.Sum(bs)
	return "?" + fmt.Sprintf("%x", ts[:4]), nil
}

func qs(method string, abs string) (string, error) {
	var qry string
	var err error
	switch method {
	case "timestamp":
		var fileinfo os.FileInfo
		fileinfo, err = os.Stat(abs)
		if err != nil {
			return "", err
		}
		qry, err = modHash(fileinfo)
	case "sum":
		var r io.ReadCloser
		r, err := os.Open(abs)
		if err != nil {
			return "", err
		}
		qry, err = sumHash(r)
		if err != nil {
			return "", err
		}
	}
	return qry, err
}
