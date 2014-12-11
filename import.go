package sprite_sass

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

func (p *Parser) ImportPath(dir, file string, mainfile string, partialMap *SafePartialMap) (string, string, error) {
	// fmt.Println("Importing: " + file)
	baseerr := ""
	//Load and retrieve all tokens from imported file
	path, _ := filepath.Abs(fmt.Sprintf("%s/%s.scss", dir, file))
	pwd := filepath.Dir(path)
	// Sass put _ in front of imported files
	fpath := filepath.Join(pwd, "/_"+filepath.Base(path))
	contents, err := ioutil.ReadFile(fpath)
	if err == nil {
		partialMap.AddRelation(mainfile, fpath)
		return pwd, string(contents), nil
	}
	baseerr += fpath + "\n"
	if strings.HasSuffix(err.Error(), "no such file or directory") {
		// Look through the import path for the file
		for _, lib := range p.Includes {
			path, _ := filepath.Abs(lib + "/" + file)
			pwd := filepath.Dir(path)
			fpath = filepath.Join(pwd, "/_"+filepath.Base(path)+".scss")
			contents, err := ioutil.ReadFile(fpath)
			baseerr += fpath + "\n"
			if err == nil {
				partialMap.AddRelation(mainfile, fpath)
				return pwd, string(contents), nil
			} else {
				// Attempt invalid name lookup (no _)
				fpath = filepath.Join(pwd, "/"+filepath.Base(path)+".scss")
				contents, err = ioutil.ReadFile(fpath)
				baseerr += fpath + "\n"
				if err == nil {
					partialMap.AddRelation(mainfile, fpath)
					return pwd, string(contents), nil
				}
			}
		}
	}
	// Ignore failures on compass
	re := regexp.MustCompile("compass\\/?")
	if re.Match([]byte(file)) {
		return pwd, string(contents), nil //errors.New("compass")
	}
	if file == "images" {
		return pwd, string(contents), nil
	}
	return pwd, string(contents), errors.New("Could not import: " +
		file + "\nTried:\n" + baseerr)
}
