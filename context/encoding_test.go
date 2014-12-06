package context

import (
	"fmt"
	"image/color"
	"reflect"
	"testing"
)

type unsupportedStruct struct {
	value float64
}

func testMarshal(t *testing.T, v interface{}) UnionSassValue {
	res, err := Marshal(v)
	if err != nil {
		t.Error(err)
	}
	return res
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
	x := testMarshal(t, unk)
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
	x := testMarshal(t, num)
	_ = Unmarshal(x, &num2)

	if num2 != num {
		t.Errorf("got: %v wanted: %v", num2, num)
	}
}

func TestMarshalList(t *testing.T) {
	lst1 := []float64{1, 2, 3, 4}
	var lst2 []float64

	x := testMarshal(t, lst1)
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

	x := testMarshal(t, fl)
	_ = Unmarshal(x, &intf)

	if fl != intf {
		t.Errorf("got: %v wanted: %v", intf, fl)
	}
}

func TestMarshalBool(t *testing.T) {
	var b = bool(true)
	var be bool

	bm := testMarshal(t, b)
	Unmarshal(bm, &be)

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

	lstm := testMarshal(t, lst)
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

	lstm := testMarshal(t, lst)
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

	snm := testMarshal(t, sn)
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

	cm := testMarshal(t, c)
	_ = Unmarshal(cm, &ce)

	if !reflect.DeepEqual(ce, c) {
		t.Errorf("What the damn hell. Wanted:\n%#v\ngot:\n% #v", c, ce)
	}
}

func TestMarshalUnsupportedStruct(t *testing.T) {
	us := unsupportedStruct{
		value: 5.5,
	}

	_, err := Marshal(us)

	expectedErr := fmt.Errorf(
		"The struct type %s is unsupported for marshalling",
		reflect.TypeOf(us).String())

	if !reflect.DeepEqual(expectedErr, err) {
		t.Errorf("Marshalling of unsupported struct did not return an error")
	}
}

func TestQuotedStringUnmarshal(t *testing.T) {
	testmap := []string{
		"\"Taylor Swift\"",
		"'Taylor Swift'",
	}
	e := "Taylor Swift"

	var se string

	for _, s := range testmap {
		sm := testMarshal(t, s)
		Unmarshal(sm, &se)
		if e != se {
			t.Errorf("What the damn hell. Got: %s wanted: %s", se, e)
		}

	}
}

func TestOptionalParameters(t *testing.T) {
	var lst = []interface{}{5, "a", true}
	var i float64
	var s string
	var b bool
	var s2 string
	var i2 float64
	var ir = float64(5)
	var sr = string("a")
	var br = bool(true)

	lstm := testMarshal(t, lst)
	_ = Unmarshal(lstm, &i, &s, &b, &s2, &i2)

	if i != ir {
		t.Errorf("got: %f wanted: %f", i, ir)
	}
	if s != sr {
		t.Errorf("got: %s wanted: %s", s, sr)
	}
	if b != br {
		t.Errorf("got: %t wanted: %t", b, br)
	}
	if s2 != "" {
		t.Errorf("got: %s wanted empty string", s)
	}
	if i2 != 0 {
		t.Errorf("got: %f wanted: 0", i2)
	}

}

func TestNullUnionSassValue(t *testing.T) {
	var usv UnionSassValue
	var inf interface{}
	err := Unmarshal(usv, &inf)

	if err.Error() != "I can't work with this. arg UnionSassValue must not be nil. - Unmarshaller" {
		t.Errorf("got: %s wanted 'I can't work with this. arg UnionSassValue must not be nil. - Unmarshaller'", err)
	}
}

func TestWrongUnmarshalToFloatType(t *testing.T) {
	s := "Taylor Swift"
	var ie float64

	sm := testMarshal(t, s)
	err := Unmarshal(sm, &ie)

	if err.Error() != "SassValue type mismatch.  Sassvalue is type \"string\" and has value \"Taylor Swift\" but expected float64" {
		t.Errorf("Unmarshal mismatch error not thrown. Got %s, wanted \"SassValue type mismatch.  Sassvalue is type \"string\" and has value \"Taylor Swift\" but expected float64\"", err)
	}
}
