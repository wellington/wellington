package sprite_sass_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func fileString(path string) string {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func TestRun(t *testing.T) {

	ipath := "test/_var.scss"
	opath := "test/var.css.out"

	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Out:          "",
	}

	ctx.Run(ipath, opath)

	res := fileString(opath)
	e := fileString("test/var.css")
	if e == "" {
		t.Errorf("Error reading in expected file.")
	}

	res = rerandom.ReplaceAllString(res, "")
	if strings.TrimSpace(res) !=
		strings.TrimSpace(e) {
		t.Errorf("Processor file did not match was: "+
			"\n~%s~\n exp:\n~%s~", res, e)
	}

	// Clean up files
	defer func() {
		os.Remove(opath)
	}()
}

func TestCompile(t *testing.T) {
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Src:          fileString("test/file1.scss"),
		Out:          "",
	}
	ctx.Compile()

	bytes, _ := ioutil.ReadFile("test/file1.css")
	file1 := string(bytes)

	if ctx.Out != file1 {
		t.Errorf("file1.scss string mismatch found: \n%s, expected \n%s", ctx.Out, file1)
	}

	ctx.Src = fileString("test/file1a.scss")
	ctx.Compile()

	if ctx.Out != file1 {
		t.Errorf("file2.scss string mismatch found: \n%s, expected \n%s", ctx.Out, file1)
	}
}

func TestExport(t *testing.T) {
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Src:          fileString("test/sheet1.scss"),
		Out:          "",
	}
	ctx.Compile()
}
