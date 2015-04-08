package wellington

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/wellington/wellington/context"
)

type nopCloser struct {
	io.Reader
}

func (n nopCloser) Close() error { return nil }

// Attempt _{}.scss, _{}.sass, {}.scss, {}.sass paths and return
// reader if found
// Returns the file contents, pwd, and error if any
func findFile(dir, file string) (io.ReadCloser, string, error) {
	var errs string
	spath, _ := filepath.Abs(filepath.Join(dir, file))
	pwd := filepath.Dir(spath)
	base := filepath.Base(spath)
	fpath := filepath.Join(pwd, "_"+base+".scss")
	r, err := readSass(fpath)
	if err == nil {
		return r, filepath.Dir(fpath), err
	}
	errs += "importPath:\n    " + err.Error()

	fpath = filepath.Join(pwd, base+".scss")
	r, err = readSass(fpath)
	if err == nil {
		return r, filepath.Dir(fpath), err
	}
	errs += "\n    " + err.Error()

	fpath = filepath.Join(pwd, "_"+base+".sass")
	r, err = readSass(fpath)
	if err == nil {
		return r, filepath.Dir(fpath), err
	}
	errs += "\n    " + err.Error()

	fpath = filepath.Join(pwd, base+".sass")
	r, err = readSass(fpath)
	if err == nil {
		return r, filepath.Dir(fpath), err
	}
	errs += "\n    " + err.Error()

	return nil, pwd, os.ErrNotExist
}

// readSassBytes converts readSass to []byte
func readSassBytes(path string) ([]byte, error) {
	reader, err := readSass(path)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

// readSass retrives a file from path. If found, it converts Sass
// to Scss or returns found Scss;
func readSass(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return ToScssReader(file)
}

// ToScssReader ...
func ToScssReader(r io.Reader) (io.ReadCloser, error) {
	var (
		buf bytes.Buffer
	)

	tr := io.TeeReader(r, &buf)
	if IsSass(tr) {

		var ibuf bytes.Buffer
		context.ToScss(io.MultiReader(&buf, r), &ibuf)
		return nopCloser{&ibuf}, nil
	}
	mr := io.MultiReader(&buf, r)

	return nopCloser{mr}, nil
}

// IsSass determines if the given reader is Sass (not Scss).
// This is predicted by the presence of semicolons
func IsSass(r io.Reader) bool {
	scanner := bufio.NewScanner(r)
	var cmt bool
	var last string
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		last = text
		// FIXME: This is not a suitable way to detect comments
		if strings.HasPrefix(text, "/*") {
			cmt = true
		}
		if strings.HasSuffix(text, "*/") {
			cmt = false
			continue
		}
		if cmt {
			continue
		}
		if strings.HasSuffix(text, "{") ||
			strings.HasSuffix(text, "}") {
			return false
		}
		if strings.HasSuffix(text, ";") {
			return false
		}
	}
	// If type is still undecided and file ends with comment, assume
	// this is a scss file
	if strings.HasSuffix(last, "*/") {
		return false
	}
	return true
}
