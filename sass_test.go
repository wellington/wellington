package sprite_sass

import (
	"io/ioutil"
	"testing"
)

func fileString(path string) string {
	bytes, _ := ioutil.ReadFile(path)
	return string(bytes)
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
