package context

import (
	"errors"
	"fmt"
	"image/color"
	"reflect"
	"testing"
)

type unsupportedStruct struct {
	value float64
}

func TestUnmarshalNumber(t *testing.T) {

	c := float64(1)
	sv, _ := makevalue(c)
	var i float64
	Unmarshal(sv, &i)
	if c != i {
		t.Errorf("got: %f wanted: %f", i, c)
	}

	d := 1.5
	dv, _ := makevalue(d)
	var ed float64
	Unmarshal(dv, &ed)
	if d != ed {
		t.Errorf("got: %f wanted: %f", ed, d)
	}

	d = 2
	dv, _ = makevalue(d)
	var ei int
	err := Unmarshal(dv, &ei)
	if err == nil {
		t.Error("No error thrown for invalid type")
	}
	if e := "Unsupported SassValue"; e != err.Error() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", err.Error(), e)
	}

}

func TestUnmarshalStringValue(t *testing.T) {
	e := "example"
	input, _ := makevalue(e)
	var s string
	Unmarshal(input, &s)
	if e != s {
		t.Errorf("got: % #v\nwanted: %s", s, e)
	}
}

func TestUnmarshalComplex(t *testing.T) {
	// Only interfaces supported for lists, is this ok?
	e := []string{"ex1", "ex2"}
	list, _ := makevalue(e)
	var s []string
	Unmarshal(list, &s)

	if len(s) != len(e) {
		t.Error("Length mismatch")
		return
	}

	for i := range e {
		if e[i] != s[i] {
			t.Errorf("got: %v wanted %v", s, e)
		}
	}
}

// Can't import C in the test package, so this is how to test cgo code
func TestUnmarshalUnknown(t *testing.T) {
	// Test for nil (no value, pointer, or empty error)
	var unk UnionSassValue
	x, _ := Marshal(unk)
	var v interface{}
	_ = Unmarshal(x, &v)
	if v != "<nil>" {
		t.Error("non-nil returned")
	}

	// Need a test for non-supported type
}

func TestMarshalNumber(t *testing.T) {
	num := float64(24)
	var num2 float64
	x, _ := Marshal(num)
	_ = Unmarshal(x, &num2)

	if num2 != num {
		t.Errorf("got: %v wanted: %v", num2, num)
	}
}

func TestMarshalList(t *testing.T) {
	lst1 := []float64{1, 2, 3, 4}
	var lst2 []float64

	x, _ := Marshal(lst1)
	_ = Unmarshal(x, &lst2)

	if len(lst1) != len(lst2) {
		t.Error("List length mismatch")
	}

	for i := range lst1 {
		if lst1[i] != lst2[i] {
			t.Errorf("wanted: %f got: %f", lst1[i], lst2[i])
		}
	}
}

func TestMarshalNumberInterface(t *testing.T) {
	var fl = float64(3)
	var intf interface{}

	x, _ := Marshal(fl)
	_ = Unmarshal(x, &intf)

	if fl != intf {
		t.Errorf("got: %v wanted: %v", intf, fl)
	}
}

func TestMarshalBool(t *testing.T) {
	var b = bool(true)
	var be bool

	bm, _ := Marshal(b)
	_ = Unmarshal(bm, &be)

	if b != be {
		t.Errorf("got: %t wanted: %t", be, b)
	}
}

func TestMarshalInterfaceListToMultiVariable(t *testing.T) {
	var lst = []interface{}{5, "a", true}
	var i float64
	var s string
	var b bool
	var ir = float64(5)
	var sr = string("a")
	var br = bool(true)

	lstm, _ := Marshal(lst)
	_ = Unmarshal(lstm, &i, &s, &b)

	if i != ir {
		t.Errorf("got: %f wanted: %f", i, ir)
	}
	if s != sr {
		t.Errorf("got: %s wanted: %s", s, sr)
	}
	if b != br {
		t.Errorf("got: %t wanted: %t", b, br)
	}
}

func TestMarshalInterfaceListSingleVariable(t *testing.T) {
	var lst = []interface{}{5}
	var i float64
	var ir = float64(5)

	lstm, _ := Marshal(lst)
	_ = Unmarshal(lstm, &i)

	if i != ir {
		t.Errorf("got: %f wanted: %f", i, ir)
	}
}

func TestMarshalSassNumber(t *testing.T) {
	sn := SassNumber{
		value: float64(3.5),
		unit:  "px",
	}
	var sne = SassNumber{}

	snm, _ := Marshal(sn)
	_ = Unmarshal(snm, &sne)

	if !reflect.DeepEqual(sne, sn) {
		t.Errorf("wanted:\n%#v\ngot:\n% #v", sn, sne)
	}
}

func TestMarshalColor(t *testing.T) {
	c := color.RGBA{
		R: uint8(5),
		G: uint8(6),
		B: uint8(7),
		A: uint8(8),
	}
	var ce = color.RGBA{}

	cm, _ := Marshal(c)
	_ = Unmarshal(cm, &ce)

	if !reflect.DeepEqual(ce, c) {
		t.Errorf("wanted:\n%#v\ngot:\n% #v", c, ce)
	}
}

func TestMarshalUnsupportedStruct(t *testing.T) {
	us := unsupportedStruct{
		value: 5.5,
	}

	_, er := Marshal(us)
	expectedErr := errors.New(fmt.Sprintf("The struct type %s is unsupported for marshalling", reflect.TypeOf(us).String()))

	if !reflect.DeepEqual(expectedErr, er) {
		t.Errorf("Marshalling of unsupported struct did not return an error")
	}
}
