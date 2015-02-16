package wellington

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
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
	// fmt.Println("Importing: " + file)
	baseerr := ""
	//Load and retrieve all tokens from imported file
	/*path, _ := filepath.Abs(fmt.Sprintf("%s/%s.scss", dir, file))
	pwd := filepath.Dir(path)
	// Sass put _ in front of imported files
	fpath := filepath.Join(pwd, "/_"+filepath.Base(path))
	contents, err := readSassBytes(fpath)
	if err == nil {
		p.PartialMap.AddRelation(p.MainFile, fpath)
		return pwd, string(contents), nil
	}
	// Try again without _, invalidish
	fpath = filepath.Join(pwd, filepath.Base(path))
	contents, err = readSassBytes(fpath)
	if err == nil {
		p.PartialMap.AddRelation(p.MainFile, fpath)
		return pwd, string(contents), nil
	}*/
	r, fpath, err := importPath(dir, file)
	if err == nil {
		p.PartialMap.AddRelation(p.MainFile, fpath)
		contents, _ := ioutil.ReadAll(r)
		return filepath.Dir(fpath), string(contents), nil
	}
	baseerr += fpath + "\n"
	if strings.HasSuffix(err.Error(), "no such file or directory") {
		// Look through the import path for the file
		for _, lib := range p.Includes {
			path, _ := filepath.Abs(lib + "/" + file)
			pwd := filepath.Dir(path)
			fpath = filepath.Join(pwd, "/_"+filepath.Base(path)+".scss")
			contents, err := readSassBytes(fpath)
			baseerr += fpath + "\n"
			if err == nil {
				p.PartialMap.AddRelation(p.MainFile, fpath)
				return pwd, string(contents), nil
			}
			// Attempt invalid name lookup (no _)
			fpath = filepath.Join(pwd, "/"+filepath.Base(path)+".scss")
			contents, err = readSassBytes(fpath)
			baseerr += fpath + "\n"
			if err == nil {
				p.PartialMap.AddRelation(p.MainFile, fpath)
				return pwd, string(contents), nil
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
	return filepath.Dir(fpath), "", errors.New("Could not import: " +
		file + "\nTried:\n" + baseerr)
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

	fpath = filepath.Join(pwd, "_"+base+".sass")
	if r, err := readSass(fpath); err == nil {
		return r, fpath, err
	}

	fpath = filepath.Join(pwd, base+".scss")
	if r, err := readSass(fpath); err == nil {
		return r, fpath, err
	}

	fpath = filepath.Join(pwd, base+".sass")
	if r, err := readSass(fpath); err == nil {
		return r, fpath, err
	}

	return nil, "", errors.New("Unable to import path:" + dir + " " + file)
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

	fmt.Println("Sass?", IsSass(bufio.NewReader(tr)))
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
