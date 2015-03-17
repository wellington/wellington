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
// Returns the new pwd, file contents, and error.
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
func (p *Parser) ImportPath(dir, file string) (string, []byte, error) {
	s, bs, err := p.importPath(dir, file)

	return s, bs, err
}

func (p *Parser) importPath(dir, file string) (string, []byte, error) {
	var baseerr string
	// Attempt pwd
	r, fpath, err := findFile(dir, file)
	if err == nil {
		p.PartialMap.AddRelation(p.MainFile, fpath)
		contents, err := ioutil.ReadAll(r)
		if err != nil {
			return "", nil, err
		}
		r.Close()
		return fpath, contents, nil
	}
	rel, _ := filepath.Rel(p.SassDir, fpath)
	if rel == "" {
		rel = "."
	}
	baseerr += rel + "\n    "

	if os.IsNotExist(err) {
		// Look through the import path for the file
		for _, lib := range p.Includes {
			r, pwd, err := findFile(lib, file)
			if err != nil {
				return "", nil, err
			}
			bs, err := ioutil.ReadAll(r)
			if err != nil {
				return "", nil, err
			}
			p.PartialMap.AddRelation(p.MainFile, fpath)
			r.Close()
			return pwd, bs, nil
		}
	}
	// Ignore failures on compass
	re := regexp.MustCompile("compass\\/?")
	if re.Match([]byte(file)) {
		return filepath.Dir(fpath), nil, nil //errors.New("compass")
	}
	if file == "images" {
		return filepath.Dir(fpath), nil, nil
	}

	baseerr += strings.Join(p.Includes, "\n    ")
	return filepath.Dir(fpath), nil,
		errors.New("Could not import: " + file + "\nTried:\n    " + baseerr)
}

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
