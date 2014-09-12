package sprite_sass_test

import (
	"io"
	"io/ioutil"
	"os"
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

func fileReader(path string) io.Reader {
	reader, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return reader
}

func TestRun(t *testing.T) {

	ipath := "test/_var.scss"
	//opath := "test/var.css.out"

	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Out:          "",
	}

	ctx.Run(fileReader(ipath), os.Stdout, "test")

	//res := fileString(opath)
	e := fileString("test/var.css")
	if e == "" {
		t.Errorf("Error reading in expected file.")
	}

	/*res = rerandom.ReplaceAllString(res, "")
	if strings.TrimSpace(res) !=
		strings.TrimSpace(e) {
		t.Errorf("Processor file did not match was: "+
			"\n~%s~\n exp:\n~%s~", res, e)
	}*/

}

/*func TestCompile(t *testing.T) {
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Src:          fileString("test/file1.scss"),
		Out:          "",
	}
	ctx.Compile()
	fpath := "test/file1.css"
	bytes, _ := ioutil.ReadFile(fpath)
	exp := string(bytes)

	if ctx.Out != exp {
		t.Errorf("%s string mismatch found: \n%s, expected \n%s",
			fpath, ctx.Out, exp)
	}
	fpath = "test/file1a.scss"
	ctx.Src = fileString(fpath)
	ctx.Compile()

	if ctx.Out != exp {
		t.Errorf("%s string mismatch found: \n%s, expected \n%s",
			fpath, ctx.Out, exp)
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
}*/
