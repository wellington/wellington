package context

import (
	"bytes"
	"image/color"
	"reflect"
	"testing"
	"time"
)

func TestFunc_simpletypes(t *testing.T) {
	in := bytes.NewBufferString(`div {
  background: foo(null, 3px, asdf, false, #005500);
}`)

	//var out bytes.Buffer
	ctx := Context{}
	ctx.Cookies = make([]Cookie, 1)
	// Communication channel for the C Sass callback function
	ch := make(chan []interface{}, 1)

	ctx.Cookies[0] = Cookie{
		"foo($null, $num, $str, $bool, $color)", func(c *Context, usv UnionSassValue) UnionSassValue {
			// Send the interface fn arguments to the ch channel

			var n interface{}
			var num SassNumber
			var s string
			var b bool
			var col = color.RGBA{}
			var intf = []interface{}{n, num, s, b, col}
			Unmarshal(usv, &intf)
			ch <- intf
			res, _ := Marshal(false)
			return res
		}, &ctx,
	}
	var out bytes.Buffer
	err := ctx.Compile(in, &out)
	if err != nil {
		t.Error(err)
	}

	e := []interface{}{
		"<nil>",
		SassNumber{3.0, "px"},
		"asdf",
		false,
		color.RGBA{R: 0x0, G: 0x55, B: 0x0, A: 0x1},
	}
	var args []interface{}
	select {
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	case args = <-ch:
	}
	if !reflect.DeepEqual(e, args) {
		t.Errorf("wanted:\n% #v\ngot:\n% #v", e, args)
	}
}

func TestFunc_complextypes(t *testing.T) {
	in := bytes.NewBufferString(`div {
	     background: foo((a,b,1mm,#003300));
	   }`)

	var out bytes.Buffer
	ctx := Context{}
	if ctx.Cookies == nil {
		ctx.Cookies = make([]Cookie, 1)
	}
	ch := make(chan interface{}, 1)
	ctx.Cookies[0] = Cookie{
		"foo($list)", func(c *Context, usv UnionSassValue) UnionSassValue {
			var sv interface{}
			Unmarshal(usv, &sv)
			ch <- sv
			res, _ := Marshal(false)
			return res
		}, &ctx,
	}
	err := ctx.Compile(in, &out)
	if err != nil {
		t.Error(err)
	}

	e := []interface{}{
		"a",
		"b",
		SassNumber{1, "mm"},
		color.RGBA{R: 0x0, G: 0x33, B: 0x0, A: 0x1},
	}
	var args interface{}
	select {
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	case args = <-ch:
	}
	if !reflect.DeepEqual(e, args) {
		t.Errorf("wanted:\n%#v\ngot:\n% #v", e, args)
	}
}

func TestFunc_customarity(t *testing.T) {
	in := bytes.NewBufferString(`div {
  color: red(blue);
  background: foo(1pt, 2cm);
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
	e := `Error > stdin:3
function foo only takes 0 arguments; given 2
div {
  color: red(blue);
  background: foo(1pt, 2cm);
}
`
	if e != err.Error() {
		t.Errorf("wanted:\n%s\ngot:\n%s\n", e, err)
	}

}
