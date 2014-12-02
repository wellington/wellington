package context

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
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
	err := ctx.Compile(in, &out)
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

func TestContextNilRun(t *testing.T) {
	var in, out bytes.Buffer
	ctx := Context{}
	err := ctx.Compile(&in, &out)
	if err == nil {
		t.Error("No error returned")
	}
	if e := "No input provided"; e != err.Error() {
		t.Errorf("wanted:\n%s\ngot:\n%s", e, err)
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
	err := ctx.Compile(in, &out)
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

func TestContextCustomSimpleTypes(t *testing.T) {
	in := bytes.NewBufferString(`div {
  background: foo(null, 3, asdf, false, #005500);
}`)

	//var out bytes.Buffer

	ctx := Context{}
	ctx.Cookies = make([]Cookie, 1)
	// Communication channel for the C Sass callback function
	ch := make(chan []SassValue, 1)
	ctx.Cookies[0] = Cookie{
		"foo($null, $num, $str, $bool, $color)", func(usv UnionSassValue) UnionSassValue {
			// Send the SassValue fn arguments to the ch channel
			var sv []SassValue
			Unmarshal(usv, &sv)
			ch <- sv

			return Marshal(false)
		}, &ctx,
	}
	var out bytes.Buffer
	err := ctx.Compile(in, &out)
	if err != nil {
		t.Error(err)
	}

	e := []SassValue{
		"<nil>",
		3.0,
		"asdf",
		false,
		color.RGBA{R: 0x0, G: 0x55, B: 0x0, A: 0x1},
	}
	args := <-ch
	if !reflect.DeepEqual(e, args) {
		t.Errorf("wanted:\n% #v\ngot:\n% #v", e, args)
	}
}

func TestContextCustomComplexTypes(t *testing.T) {
	in := bytes.NewBufferString(`div {
	     background: foo((a,b,1,#003300), (a:(b:#003300,c:(d:4,e:str))));
	   }`)

	var out bytes.Buffer
	ctx := Context{}
	if ctx.Cookies == nil {
		ctx.Cookies = make([]Cookie, 1)
	}
	ch := make(chan []SassValue, 1)
	ctx.Cookies[0] = Cookie{
		"foo($list, $map)", func(usv UnionSassValue) UnionSassValue {
			var sv []SassValue
			Unmarshal(usv, &sv)
			ch <- sv
			return Marshal(false)
		}, &ctx,
	}
	err := ctx.Compile(in, &out)
	if err != nil {
		t.Error(err)
	}

	e := []SassValue{
		[]SassValue{
			"a",
			"b",
			float64(1),
			color.RGBA{R: 0x0, G: 0x33, B: 0x0, A: 0x1},
		},
		//maps not implemented
		map[SassValue]SassValue{
			"a": map[SassValue]SassValue{
				"b": color.RGBA{R: 0x0, G: 0x33, B: 0x0, A: 0x1},
				"c": map[SassValue]SassValue{"d": 4, "e": "str"},
			},
		},
	}

	args := <-ch

	if !reflect.DeepEqual(e[0], args[0]) {
		t.Errorf("wanted:\n%#v\ngot:\n% #v", e[0], args[0])
	}
}

func TestContextCustomArity(t *testing.T) {
	in := bytes.NewBufferString(`div {
  color: red(blue);
  background: foo(1, 2);
}`)

	var out bytes.Buffer
	ctx := Context{}
	if ctx.Cookies == nil {
		ctx.Cookies = make([]Cookie, 1)
	}

	ctx.Cookies[0] = Cookie{
		"foo()", SampleCB, &ctx,
	}
	err := ctx.Compile(in, &out)
	if err == nil {
		t.Error("No error thrown for incorrect arity")
	}

	if e := "function foo only takes 0 arguments; given 2"; e != ctx.Errors.Message {
		t.Errorf("wanted:\n%s\ngot:\n%s\n", e, ctx.Errors.Message)
	}
}

func ExampleContext_Compile() {
	in := bytes.NewBufferString(`div {
			  color: red(blue);
			  background: foo();
			}`)

	var out bytes.Buffer
	ctx := Context{
	//Customs: []string{"foo()"},
	}
	if ctx.Cookies == nil {
		ctx.Cookies = make([]Cookie, 1)
	}
	ctx.Cookies[0] = Cookie{
		"foo()", func(usv UnionSassValue) UnionSassValue {
			return Marshal("no-repeat")
		}, &ctx,
	}
	err := ctx.Compile(in, &out)
	if err != nil {
		panic(err)
	}

	fmt.Print(out.String())
	// Output:
	// div {
	//   color: 0;
	//   background: no-repeat; }
}

func BenchmarkContextCompile(b *testing.B) {
	// TBD
}
