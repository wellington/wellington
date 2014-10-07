package sprite_sass

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
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

func TestContextRun(t *testing.T) {

	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Out:          "",
	}

	var scanned []byte
	ipath := "test/_var.scss"
	exp, err := ioutil.ReadFile("test/var.css")
	if err != nil {
		panic(err)
	}

	r, w := io.Pipe()
	go func(ipath string, w io.WriteCloser) {

		err := ctx.Run(fileReader(ipath), w, "test")
		if err != nil {
			panic(err)
		}
	}(ipath, w)

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)
	for scanner.Scan() {
		scanned = append(scanned, scanner.Bytes()...)
	}

	defer cleanUpSprites(ctx.Parser.Sprites)

	scanned = rerandom.ReplaceAll(scanned, []byte(""))
	if string(scanned) != string(exp) {
		t.Errorf("Processor file did not match was: "+
			"\n%s\n exp:\n%s", string(scanned), string(exp))
	}

}

func TestContextImport(t *testing.T) {

	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Out:          "",
	}

	var scanned []byte
	ipath := "test/import.scss"
	exp, err := ioutil.ReadFile("test/import.css")
	if err != nil {
		panic(err)
	}

	r, w := io.Pipe()
	go func(ipath string, w io.WriteCloser) {

		err := ctx.Run(fileReader(ipath), w, "test")
		if err != nil {
			panic(err)
		}
	}(ipath, w)

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)

	for scanner.Scan() {
		scanned = append(scanned, scanner.Bytes()...)
	}
	defer cleanUpSprites(ctx.Parser.Sprites)

	scanned = rerandom.ReplaceAll(scanned, []byte(""))
	res := string(scanned)
	if e := string(exp); res != e {
		t.Errorf("Processor file did not match \nexp: "+
			"\n~%s~\n was:\n~%s~", e, res)
	}

}

func TestContextFail(t *testing.T) {

	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Out:          "",
	}

	var scanned []byte
	ipath := "test/_failimport.scss"

	r, w := io.Pipe()
	go func(ipath string, w io.WriteCloser) {

		err := ctx.Run(fileReader(ipath), w, "test")
		if err == nil {
			t.Error("Invalid SCSS was not found")
		}
		errs := strings.Split(err.Error(), "\n")
		libsassErr := errs[0]
		parsedErr := errs[1]

		e := "source string:9: error: invalid top-level expression"
		if e != libsassErr {
			t.Errorf("expected:\n%s\nwas:\n%s\n", e, libsassErr)
		}

		e = "error in fail:4"
		if e != parsedErr {
			t.Errorf("expected:\n%s\nwas:\n%s\n", e, parsedErr)
		}

	}(ipath, w)

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)

	for scanner.Scan() {
		scanned = append(scanned, scanner.Bytes()...)
	}
	defer cleanUpSprites(ctx.Parser.Sprites)

	scanned = rerandom.ReplaceAll(scanned, []byte(""))
	_ = scanned
}

func TestContextNilRun(t *testing.T) {
	ctx := Context{}
	var w io.WriteCloser
	err := ctx.Run(nil, w, "test")
	if err == nil {
		t.Errorf("No error returned: %s", err)
	}

}

func TestContextCompile(t *testing.T) {
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Src:          fileString("test/file1.scss"),
		Out:          "",
	}

	err := ctx.Compile()
	if err != nil {
		t.Errorf("Compilation failed: %s", err)
	}
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

func TestContextExport(t *testing.T) {
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Src:          fileString("test/sheet1.scss"),
		Out:          "",
	}
	ctx.Compile()
	defer cleanUpSprites(ctx.Parser.Sprites)
}

func BenchmarkContextCompile(b *testing.B) {
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Src:          fileString("test/file1.scss"),
		Out:          "",
		ImageDir:     "/tmp",
	}
	for n := 0; n < b.N; n++ {
		ctx.Compile()
	}
}
