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
)

// ImportPath accepts a directory and file path to find partials for importing.
// Returning a new pwd and string of the file contents.
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
	baseerr := ""
	r, fpath, err := importPath(dir, file)
	if err == nil {
		p.PartialMap.AddRelation(p.MainFile, fpath)
		contents, _ := ioutil.ReadAll(r)
		return filepath.Dir(fpath), string(contents), nil
	}
	rel, _ := filepath.Rel(p.SassDir, fpath)
	if rel == "" {
		rel = "./"
	}
	baseerr += rel + "\n"
	if strings.HasSuffix(err.Error(), "no such file or directory") {
		// Look through the import path for the file
		for _, lib := range p.Includes {
			r, pwd, err := importPath(lib, file)
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

	baseerr += strings.Join(p.Includes, "\n")
	return filepath.Dir(fpath), "",
		errors.New("Could not import: " + file + "\nTried:\n" + baseerr)
}

// Attempt _{}.scss, _{}.sass, {}.scss, {}.sass paths and return
// reader if found
func importPath(dir, file string) (io.Reader, string, error) {
	spath, _ := filepath.Abs(dir + "/" + file)
	pwd := filepath.Dir(spath)
	base := filepath.Base(spath)

	fpath := filepath.Join(pwd, "_"+base+".scss")
	if r, err := readSass(fpath); err == nil {
		return r, fpath, err
	}
	fpath = filepath.Join(pwd, base+".scss")
	if r, err := readSass(fpath); err == nil {
		return r, fpath, err
	}

	fpath = filepath.Join(pwd, "_"+base+".sass")
	if r, err := readSass(fpath); err == nil {
		return r, fpath, err
	}

	fpath = filepath.Join(pwd, base+".sass")
	if r, err := readSass(fpath); err == nil {
		return r, fpath, err
	}

	return nil, pwd, errors.New("Unable to import path:" + dir + " " + file)
}

func readSassBytes(path string) ([]byte, error) {
	reader, err := readSass(path)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

// readSass retrives a file from path. If found, it converts Sass
// to Scss or returns found Scss;
func readSass(path string) (io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	tr := io.TeeReader(file, &buf)
	_ = tr
	//fmt.Println("Sass?", IsSass(bufio.NewReader(tr)))
	mr := io.MultiReader(&buf, file)
	return mr, nil
}

// IsSass determines if the given reader is Sass (not Scss).
// This is predicted by the presence of semicolons
func IsSass(r *bufio.Reader) bool {
	for {
		line, err := r.ReadString('\n')
		clean := strings.TrimSpace(line)
		// Errors, empty file probably
		if err != nil {
			return false
		}
		if strings.HasSuffix(clean, "{") ||
			strings.HasSuffix(clean, "}") {
			continue
		}
		if strings.HasSuffix(clean, ";") {
			return false
		}
		// Probably Sass, say so
		return true
	}
}
