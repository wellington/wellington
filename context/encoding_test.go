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

type TestError interface {
	Error(...interface{})
}

func testMarshal(t TestError, v interface{}) UnionSassValue {
	res, err := Marshal(v)
	if err != nil {
		t.Error(err)
	}
	return res
}

func TestUnmarshalNumber(t *testing.T) {

	c := SassNumber{1.0, "mm"}
	sv, _ := makevalue(c)
	var i SassNumber
	Unmarshal(sv, &i)
	if c != i {
		t.Errorf("got: %v wanted: %v", i, c)
	}

	d := SassNumber{1.5, "pt"}
	dv, _ := makevalue(d)
	var ed SassNumber
	Unmarshal(dv, &ed)
	if d != ed {
		t.Errorf("got: %v wanted: %v", ed, d)
	}

	d = SassNumber{2.0, "TaylorSwifts"}
	dv, _ = makevalue(d)
	var ei SassNumber
	err := Unmarshal(dv, &ei)
	if err == nil {
		t.Error("No error thrown for invalid type")
	}
	if e := fmt.Sprintf("SassNumber units %s are unsupported", d.Unit); e != err.Error() {
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

func TestUnmarshalError(t *testing.T) {
	e := "error message"
	obj := Error(fmt.Errorf("%s", e))
	var s string
	err := Unmarshal(obj, &s)
	if err != nil {
		t.Error(err)
	}
	if e != s {
		t.Errorf("got: %s wanted: %s", s, e)
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

func TestMarshalInvalidUnitSassNumber(t *testing.T) {
	num := SassNumber{45, "em"}
	var num2 SassNumber
	x := testMarshal(t, num)
	error := Unmarshal(x, &num2)

	if error.Error() != "SassNumber units em are unsupported" {
		t.Errorf("got: %s wanted: %s", error.Error(), "SassNumber units em are unsupported")
	}
}

func TestMarshalList(t *testing.T) {
	lst1 := []SassNumber{SassNumber{1, "px"}, SassNumber{2, "rad"}, SassNumber{3, "grad"}, SassNumber{4, "deg"}}
	var lst2 []SassNumber

	x := testMarshal(t, lst1)
	_ = Unmarshal(x, &lst2)

	if len(lst1) != len(lst2) {
		t.Error("List length mismatch")
	}

	for i := range lst1 {
		if lst1[i] != lst2[i] {
			t.Errorf("wanted: %v got: %v", lst1[i], lst2[i])
		}
	}
}

func TestMarshalNumberInterface(t *testing.T) {
	var fl = SassNumber{3, "turn"}
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
	var lst = []interface{}{SassNumber{5, "pt"}, "a", true}
	var i SassNumber
	var s string
	var b bool
	var ir = SassNumber{5, "pt"}
	var sr = string("a")
	var br = bool(true)

	lstm := testMarshal(t, lst)
	_ = Unmarshal(lstm, &i, &s, &b)

	if !reflect.DeepEqual(i, ir) {
		t.Errorf("got: %v wanted: %v", i, ir)
	}
	if s != sr {
		t.Errorf("got: %s wanted: %s", s, sr)
	}
	if b != br {
		t.Errorf("got: %t wanted: %t", b, br)
	}
}

func TestMarshalInterfaceListToMultiVariablewList(t *testing.T) {
	var lst = []interface{}{SassNumber{5, "pt"}, "a", true, []string{"a", "b", "c", "d"}}
	var i SassNumber
	var s string
	var b bool
	var sl []string
	var ir = SassNumber{5, "pt"}
	var sr = string("a")
	var br = bool(true)
	var slr = []string{"a", "b", "c", "d"}

	lstm := testMarshal(t, lst)
	_ = Unmarshal(lstm, &i, &s, &b, &sl)

	if !reflect.DeepEqual(i, ir) {
		t.Errorf("got: %v wanted: %v", i, ir)
	}
	if s != sr {
		t.Errorf("got: %s wanted: %s", s, sr)
	}
	if b != br {
		t.Errorf("got: %t wanted: %t", b, br)
	}
	if !reflect.DeepEqual(sl, slr) {
		t.Errorf("got: %s wanted: %s", sl, slr)
	}
}

func TestMarshalInterfaceListSingleVariable(t *testing.T) {
	var lst = []interface{}{SassNumber{5, "mm"}}
	var i SassNumber
	var ir = SassNumber{5, "mm"}

	lstm := testMarshal(t, lst)
	_ = Unmarshal(lstm, &i)

	if !reflect.DeepEqual(i, ir) {
		t.Errorf("got: %v wanted: %v", i, ir)
	}
}

func TestMarshalSassNumber(t *testing.T) {
	sn := SassNumber{
		Value: float64(3.5),
		Unit:  "px",
	}
	var sne = SassNumber{}

	snm := testMarshal(t, sn)
	_ = Unmarshal(snm, &sne)

	if !reflect.DeepEqual(sne, sn) {
		t.Errorf("wanted:\n%#v\ngot:\n% #v", sn, sne)
	}
}

func TestMarshalError(t *testing.T) {
	e := "error has been thrown"
	err := fmt.Errorf(e)
	eusv := Error(err)
	var s string
	Unmarshal(eusv, &s)

	if s != e {
		t.Errorf("got:\n%s\nwanted:\n%s", s, e)
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

func TestListListtoInterfaceList(t *testing.T) {
	var lst = []interface{}{"a", "b"}
	var lstlst = []interface{}{lst}

	var lst2 []interface{}

	var elst = []interface{}{"a", "b"}

	x := testMarshal(t, lstlst)
	_ = Unmarshal(x, &lst2)

	if len(lst2) != len(elst) {
		t.Error("List length mismatch")
	}

	if !reflect.DeepEqual(lst2, elst) {
		t.Errorf("What the damn hell. Wanted:\n%#v\ngot:\n% #v", elst, lst2)
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
	var lst = []interface{}{SassNumber{5, "px"}, "a", true}
	var i SassNumber
	var s string
	var b bool
	var s2 string
	var i2 float64
	var ir = SassNumber{5, "px"}
	var sr = string("a")
	var br = bool(true)

	lstm := testMarshal(t, lst)
	_ = Unmarshal(lstm, &i, &s, &b, &s2, &i2)

	if !reflect.DeepEqual(i, ir) {
		t.Errorf("got: %v wanted: %v", i, ir)
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
	e := "I can't work with this. arg UnionSassValue must not be nil. - Unmarshaller"
	if err.Error() != e {
		t.Errorf("got: %s wanted: %s", err, e)
	}
}

func TestWrongUnmarshalToFloatType(t *testing.T) {
	s := "Taylor Swift"
	var ie SassNumber

	sm := testMarshal(t, s)
	err := Unmarshal(sm, &ie)

	e := "Sassvalue is type string and has value Taylor Swift but expected color.RGBA or SassNumber"
	if err.Error() != e {
		t.Errorf("got: %s wanted: %s", err, e)
	}
}
