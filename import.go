package wellington

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wellington/wellington/context"
)

type nopCloser struct {
	io.Reader
}

func (n nopCloser) Close() error { return nil }

// ImportPath accepts a directory and file path to find partials for importing.
// Returns the new pwd, string of the file contents, and error.
// File can contain a directory and should be evaluated if
// successfully found.
// Dir is used to provide relative context to the importee.  If no file is found
// pwd is echoed back.
//
// Paths are looked up in the following order:
// {includepath}/_file.scss
// {includePath}/_file.sass
// {includepath}/file.scss
// {includePath}/file.sass
// {Dir{dir+file}}/_{Base{file}}.scss
// {Dir{dir+file}}/_{Base{file}}.sass
// {Dir{dir+file}}/{Base{file}}.scss
// {Dir{dir+file}}/{Base{file}}.sass
func (p *Parser) ImportPath(dir, file string) (string, string, error) {
	var baseerr string
	// Attempt pwd
	r, fpath, err := importPath(dir, file)
	if err == nil {
		p.PartialMap.AddRelation(p.MainFile, fpath)
		contents, _ := ioutil.ReadAll(r)
		defer r.Close()
		return fpath, string(contents), nil
	}
	rel, _ := filepath.Rel(p.SassDir, fpath)
	if rel == "" {
		rel = "./"
	}
	baseerr += rel + "\n    "
	if os.IsNotExist(err) {
		// Look through the import path for the file
		for _, lib := range p.Includes {
			r, pwd, err := importPath(lib, file)
			defer r.Close()
			if err == nil {
				p.PartialMap.AddRelation(p.MainFile, fpath)
				bs, _ := ioutil.ReadAll(r)
				return pwd, string(bs), nil
			}
		}
	}
	// Ignore failures on compass
	re := regexp.MustCompile("compass\\/?")
	if re.Match([]byte(file)) {
		return filepath.Dir(fpath), "", nil //errors.New("compass")
	}
	if file == "images" {
		return filepath.Dir(fpath), "", nil
	}

	baseerr += strings.Join(p.Includes, "\n    ")
	return filepath.Dir(fpath), "",
		errors.New("Could not import: " + file + "\nTried:\n    " + baseerr)
}

// Attempt _{}.scss, _{}.sass, {}.scss, {}.sass paths and return
// reader if found
// Returns the file contents, pwd, and error if any
func importPath(dir, file string) (io.ReadCloser, string, error) {
	var errs string
	spath, _ := filepath.Abs(filepath.Join(dir, file))
	pwd := filepath.Dir(spath)
	base := filepath.Base(spath)
	fpath := filepath.Join(pwd, "_"+base+".scss")
	if r, err := readSass(fpath); err == nil {
		return r, filepath.Dir(fpath), err
	} else {
		errs += "importPath:\n    " + err.Error()
	}

	fpath = filepath.Join(pwd, base+".scss")
	if r, err := readSass(fpath); err == nil {
		return r, filepath.Dir(fpath), err
	} else {
		errs += "\n    " + err.Error()
	}

	fpath = filepath.Join(pwd, "_"+base+".sass")
	if r, err := readSass(fpath); err == nil {
		return r, filepath.Dir(fpath), err
	} else {
		errs += "\n    " + err.Error()
	}

	fpath = filepath.Join(pwd, base+".sass")
	if r, err := readSass(fpath); err == nil {
		return r, filepath.Dir(fpath), err
	} else {
		errs += "\n    " + err.Error()
	}

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
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if strings.HasSuffix(text, "{") ||
			strings.HasSuffix(text, "}") {
			return false
		}
		if strings.HasSuffix(text, ";") {
			return false
		}
	}
	return true
}
