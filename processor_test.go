package sprite_sass_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestProcessor(t *testing.T) {

	var (
		result, exp []byte
	)

	p := Processor{
		Ipath: "test/_var.scss",
		Opath: "test/var.css.out",
	}
	p.Run()

	result, err := ioutil.ReadFile(p.Opath)
	if err != nil {
		panic(err)
	}

	exp, err = ioutil.ReadFile("test/var.css")
	if err != nil {
		panic(err)
	}
	if exp == nil {
		t.Errorf("Error reading in expected file.")
	}

	res := rerandom.ReplaceAllString(string(result), "")
	if strings.TrimSpace(res) !=
		strings.TrimSpace(string(exp)) {
		t.Errorf("Processor file did not match was: "+
			"\n~%s~\n exp:\n~%s~", res, exp)
	}

	// Clean up files
	defer func() {
		os.Remove(p.Opath)
	}()

}
