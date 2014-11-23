package context

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/drewwells/spritewell"
)

func cleanUpSprites(sprites map[string]spritewell.ImageList) {
	if sprites == nil {
		return
	}
	for _, iml := range sprites {
		err := os.Remove(filepath.Join(iml.GenImgDir, iml.OutFile))
		if err != nil {
			log.Fatal(err)
		}
	}
}

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

var rerandom *regexp.Regexp

func init() {
	// Setup build directory
	os.MkdirAll("test/build", 0755)
	rerandom = regexp.MustCompile(`-\w{6}(?:\.(png|jpg))`)

}

func TestContextFile(t *testing.T) {

	in := bytes.NewBufferString(`div {
  span {
    color: black;
  }
  width: 100px;
  height: 100px;
}

p {
	background: red;
}`)

	var out bytes.Buffer
	ctx := Context{}
	err := ctx.Compile(in, &out, "")
	if err != nil {
		panic(err)
	}

	e := `div {
  width: 100px;
  height: 100px; }
  div span {
    color: black; }

p {
  background: red; }
`

	if e != out.String() {
		t.Errorf("wanted:\n%s\n"+
			"got:\n%s\n", e, out.String())
	}
}

func TestContextRun(t *testing.T) {

	in := bytes.NewBufferString(`$red-var: red;
$hex: #00FF00;
div {
  background: $hex;
  $hex: #00DD00;
  font-size: 10pt;
}
`)

	var out bytes.Buffer
	ctx := Context{}
	err := ctx.Compile(in, &out, "")
	if err != nil {
		panic(err)
	}

	e := `div {
  background: #00FF00;
  font-size: 10pt; }
`

	if e != out.String() {
		t.Errorf("wanted:\n%s\n"+
			"got:\n%s\n", e, out.String())
	}

}

func TestContextImport(t *testing.T) {

	ipath := "test/sass/import.scss"
	exp, err := ioutil.ReadFile("test/expected/import.css")
	if err != nil {
		t.Error(err)
	}

	ctx, scanned, _ := setupCtx(ipath)
	_ = ctx
	// defer cleanUpSprites(ctx.Parser.Sprites)

	res := string(scanned)
	if e := string(exp); res != e {
		t.Errorf("Processor file did not match \nexp: "+
			"\n%s\n was:\n%s", e, res)
	}

}

func TestContextFail(t *testing.T) {
	return
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Out:          "",
	}
	_ = ctx
	var scanned []byte
	ipath := "test/_failimport.scss"

	r, w := io.Pipe()
	go func(ipath string, w io.WriteCloser) {

		err := ctx.Compile(fileReader(ipath), w, "test")
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
	// defer cleanUpSprites(ctx.Parser.Sprites)

	scanned = rerandom.ReplaceAll(scanned, []byte(""))
	_ = scanned
}

func TestContextNilRun(t *testing.T) {
	// ctx := Context{}
	// var w io.WriteCloser
	// err := ctx.Run(nil, w, "test")
	// if err == nil {
	// 	t.Errorf("No error returned: %s", err)
	// }

}

func BenchmarkContextCompile(b *testing.B) {
	in := fileReader("test/file1.scss")
	var out io.Writer
	var err error
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		// Src:          ,
		// Out:      "",
		ImageDir: "/tmp",
	}
	for n := 0; n < b.N; n++ {
		err = ctx.Compile(in, out, "")
	}
	_ = err
}
