package sprite_sass

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

func (p *Parser) ImportPath(dir, file string) (string, string, error) {
	// fmt.Println("Importing: " + file)
	baseerr := ""
	//Load and retrieve all tokens from imported file
	path, err := filepath.Abs(fmt.Sprintf("%s/%s.scss", dir, file))
	if err != nil {
		return "", "", err
	}
	pwd := filepath.Dir(path)
	// Sass put _ in front of imported files
	fpath := pwd + "/_" + filepath.Base(path)
	contents, err := ioutil.ReadFile(fpath)
	if err == nil {
		return pwd, string(contents), nil
	}
	baseerr += fpath + "\n"
	if strings.HasSuffix(err.Error(), "no such file or directory") {
		// Look through the import path for the file
		for _, lib := range p.Includes {
			path, err := filepath.Abs(lib + "/" + file)
			if err != nil {
				return "", "", err
			}
			pwd := filepath.Dir(path)
			fpath = pwd + "/_" + filepath.Base(path) + ".scss"
			contents, err := ioutil.ReadFile(fpath)
			baseerr += fpath + "\n"
			if err == nil {
				return pwd, string(contents), nil
			}
		}
	}
	// Ignore failures on compass
	re := regexp.MustCompile("compass\\/")
	if re.Match([]byte(pwd)) {
		return pwd, string(contents), nil
	}
	return pwd, string(contents), errors.New("Could not import: " +
		file + "\nTried:\n" + baseerr)
}
