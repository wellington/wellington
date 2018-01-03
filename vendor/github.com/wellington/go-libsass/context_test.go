package libsass

import (
	"bytes"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/wellington/go-libsass/libs"
)

var rerandom *regexp.Regexp

func init() {
	// Setup build directory
	os.MkdirAll("test/build/img", 0755)
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
	ctx := newContext()
	err := ctx.compile(&out, in)
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
	ctx := newContext()
	err := ctx.compile(&out, &in)
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
	ctx := newContext()
	err := ctx.compile(&out, in)
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

func TestLibsassError(t *testing.T) {
	in := bytes.NewBufferString(`div {
  color: red(blue, purple);
}`)

	var out bytes.Buffer
	ctx := newContext()

	ctx.Funcs.Add(Func{
		Sign: "foo()",
		Fn:   TestCallback,
		Ctx:  &ctx,
	})
	err := ctx.compile(&out, in)

	if err == nil {
		t.Error("No error thrown for incorrect arity")
	}

	if e := "wrong number of arguments (2 for 1) for `red'"; e != ctx.err.Message {
		t.Errorf("wanted:%s\ngot:%s\n", e, ctx.err.Message)
	}
	e := `Error > stdin:2
wrong number of arguments (2 for 1) for ` + "`" + `red'
div {
  color: red(blue, purple);
}
`
	if e != err.Error() {
		t.Errorf("wanted:\n%s\ngot:\n%s\n", e, err)
	}
}

func ExampleContext_Compile() {
	in := bytes.NewBufferString(`div {
			  color: red(blue);
			  background: foo();
			}`)

	ctx := newContext()

	ctx.Funcs.Add(Func{
		Sign: "foo()",
		Fn: func(v interface{}, usv libs.UnionSassValue, rsv *libs.UnionSassValue) error {
			res, _ := Marshal("no-repeat")
			*rsv = res.Val()
			return nil
		},
		Ctx: &ctx,
	})
	err := ctx.compile(os.Stdout, in)
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// div {
	//   color: 0;
	//   background: no-repeat; }
}

func BenchmarkContextCompile(b *testing.B) {
	bits := []byte(`div { color: #005500; }`)
	big := []byte(`div { color: #005500; }          `)
	ctx := newContext()
	out := bytes.NewBuffer(big)

	for i := 0; i < b.N; i++ {
		in := bytes.NewBuffer(bits)
		out.Reset()
		err := ctx.compile(out, in)
		if err != nil {
			b.Error(err)
		}
	}
}
