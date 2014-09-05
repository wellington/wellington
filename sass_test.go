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
	ctx := SassContext{
		Context: &Context{
			Options: Options{
				OutputStyle:  NESTED_STYLE,
				IncludePaths: make([]string, 0),
			},
			SourceString: fileString("test/file1.scss"),
			OutputString: "",
		},
	}
	ctx.Compile()

	bytes, _ := ioutil.ReadFile("test/file1.css")
	file1 := string(bytes)

	if ctx.Context.OutputString != file1 {
		t.Errorf("file1.scss string mismatch found: \n%s, expected \n%s", ctx.Context.OutputString, file1)
	}

	ctx.Context.SourceString = fileString("test/file1a.scss")
	ctx.Compile()

	if ctx.Context.OutputString != file1 {
		t.Errorf("file2.scss string mismatch found: \n%s, expected \n%s", ctx.Context.OutputString, file1)
	}
}

func TestExport(t *testing.T) {
	ctx := SassContext{
		Output: "test/sheet1.css",
		Context: &Context{
			Options: Options{
				OutputStyle:  NESTED_STYLE,
				IncludePaths: make([]string, 0),
			},
			SourceString: fileString("test/sheet1.scss"),
			OutputString: "",
		},
	}
	ctx.Compile()
}
