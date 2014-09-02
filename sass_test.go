package sprite_sass

import (
	"io/ioutil"
	"testing"

	"github.com/moovweb/gosass"
)

func fileString(path string) string {
	bytes, _ := ioutil.ReadFile(path)
	return string(bytes)
}

func TestCompile(t *testing.T) {
	ctx := Context{
		Output: "test/file1.css",
		Sass: &gosass.Context{
			Options: gosass.Options{
				OutputStyle:  gosass.NESTED_STYLE,
				IncludePaths: make([]string, 0),
			},
			SourceString: fileString("test/file1.scss"),
			OutputString: "",
		},
	}
	ctx.Compile()

	bytes, _ := ioutil.ReadFile("test/file1.css")
	file1 := string(bytes)

	if ctx.Sass.OutputString != file1 {
		t.Errorf("file1.scss string mismatch found: \n%s, expected \n%s", ctx.Sass.OutputString, file1)
	}

	ctx.Sass.SourceString = fileString("test/file1a.scss")
	ctx.Compile()

	if ctx.Sass.OutputString != file1 {
		t.Errorf("file2.scss string mismatch found: \n%s, expected \n%s", ctx.Sass.OutputString, file1)
	}
}

func TestExport(t *testing.T) {
	ctx := Context{
		Output: "test/sheet1.css",
		Sass: &gosass.Context{
			Options: gosass.Options{
				OutputStyle:  gosass.NESTED_STYLE,
				IncludePaths: make([]string, 0),
			},
			SourceString: fileString("test/sheet1.scss"),
			OutputString: "",
		},
	}
	ctx.Compile()
}
