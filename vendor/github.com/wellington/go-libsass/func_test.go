package libsass

import (
	"bytes"
	"image/color"
	"reflect"
	"testing"
	"time"

	"github.com/wellington/go-libsass/libs"
)

func TestFunc_simpletypes(t *testing.T) {
	in := bytes.NewBufferString(`div {
  background: foo(null, 3px, asdf, false);
}`)

	//var out bytes.Buffer
	ctx := newContext()
	// Communication channel for the C Sass callback function
	ch := make(chan []interface{}, 1)

	ctx.Funcs.Add(Func{

		Sign: "foo($null, $num, $str, $bool)",
		Fn: Handler(func(v interface{}, req SassValue, res *SassValue) error {
			var n interface{}
			var num libs.SassNumber
			var s string
			var b bool
			var intf = []interface{}{n, num, s, b}
			err := Unmarshal(req, &intf)
			if err != nil {
				t.Fatal(err)
			}
			// Send the interface fn arguments to the ch channel
			ch <- intf
			return nil
		}),
		Ctx: &ctx,
	})
	var out bytes.Buffer
	err := ctx.compile(&out, in)
	if err != nil {
		t.Error(err)
	}

	e := []interface{}{
		"<nil>",
		libs.SassNumber{Value: 3.0, Unit: "px"},
		"asdf",
		false,
	}
	var args []interface{}
	select {
	case <-time.After(10 * time.Millisecond):
		t.Fatal("timeout")
	case args = <-ch:
	}
	if !reflect.DeepEqual(e, args) {
		t.Errorf("wanted:\n% #v\ngot:\n% #v", e, args)
	}
}

func TestFunc_colortype(t *testing.T) {
	in := bytes.NewBufferString(`div {
  background: foo(#005500);
}`)

	//var out bytes.Buffer
	ctx := newContext()
	// Communication channel for the C Sass callback function
	ch := make(chan []interface{}, 1)

	ctx.Funcs.Add(Func{
		Sign: "foo($color)",
		Fn: func(v interface{}, usv libs.UnionSassValue, rsv *libs.UnionSassValue) error {
			// Send the interface fn arguments to the ch channel
			var infs = []interface{}{color.RGBA{}}
			err := Unmarshal(SassValue{value: usv}, &infs)
			if err != nil {
				t.Fatal(err)
			}
			ch <- infs
			res, _ := Marshal(false)
			*rsv = res.Val()
			return nil
		},
		Ctx: &ctx,
	})
	var out bytes.Buffer
	err := ctx.compile(&out, in)
	if err != nil {
		t.Error(err)
	}

	e := []interface{}{
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
	ctx := newContext()
	ch := make(chan interface{}, 1)
	ctx.Funcs.Add(Func{
		Sign: "foo($list)",
		Fn: func(v interface{}, usv libs.UnionSassValue, rsv *libs.UnionSassValue) error {
			var sv interface{}
			Unmarshal(SassValue{value: usv}, &sv)
			ch <- sv
			res, _ := Marshal(false)
			*rsv = res.Val()
			return nil
		},
		Ctx: &ctx,
	})
	err := ctx.compile(&out, in)
	if err != nil {
		t.Error(err)
	}

	e := []interface{}{
		"a",
		"b",
		libs.SassNumber{Value: 1, Unit: "mm"},
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
	e := "wrong number of arguments (2 for 0) for `foo'"
	if e != ctx.err.Message {
		t.Errorf("wanted:\n%s\ngot:\n%s\n", e, ctx.err.Message)
	}
	e = `Error > stdin:3
wrong number of arguments (2 for 0) for ` + "`" + `foo'
div {
  color: red(blue);
  background: foo(1pt, 2cm);
}
`
	if e != err.Error() {
		t.Errorf("wanted:\n%s\ngot:\n%s\n", e, err)
	}

}
