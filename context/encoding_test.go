package context

import "testing"

func TestUnmarshalNumber(t *testing.T) {

	c := float64(1)
	sv := makevalue(c)
	var i float64
	Unmarshal(sv, &i)
	if c != i {
		t.Errorf("got: %f wanted: %f", i, c)
	}

	d := 1.5
	dv := makevalue(d)
	var ed float64
	Unmarshal(dv, &ed)
	if d != ed {
		t.Errorf("got: %f wanted: %f", ed, d)
	}

	d = 2
	dv = makevalue(d)
	var ei int
	err := Unmarshal(dv, &ei)
	if err == nil {
		t.Error("No error thrown for invalid type")
	}
	if e := "Unsupported SassValue"; e != err.Error() {
		t.Errorf("got:\n%s\nwanted:\n%s\n", err.Error(), e)
	}

}

func TestUnmarshalUnknown(t *testing.T) {
	testUnmarshalUnknown(t)
}

func TestUnmarshalStringValue(t *testing.T) {
	e := "example"
	input := makevalue(e)
	var s string
	Unmarshal(input, &s)
	if e != s {
		t.Errorf("got: % #v\nwanted: %s", s, e)
	}
}

func TestUnmarshalComplex(t *testing.T) {
	// Only interfaces supported for lists, is this ok?
	e := []string{"ex1", "ex2"}
	list := makevalue(e)
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

func TestMarshalNumber(t *testing.T) {
	testMarshalNumber(t)
}

func TestMarshalList(t *testing.T) {
	testMarshalList(t)
}

func TestMarshalListInterface(t *testing.T) {
	testMarshalNumberInterface(t)
}

// Can't import C in the test package, so this is how to test cgo code
func testUnmarshalUnknown(t *testing.T) {
	// Test for nil (no value, pointer, or empty error)
	var unk UnionSassValue
	x := Marshal(unk)
	var v interface{}
	_ = Unmarshal(x, &v)
	if v != "<nil>" {
		t.Error("non-nil returned")
	}

	// Need a test for non-supported type
}

func testMarshalNumber(t *testing.T) {
	num := float64(24)
	var num2 float64
	x := Marshal(num)
	_ = Unmarshal(x, &num2)

	if num2 != num {
		t.Errorf("got: %v wanted: %v", num2, num)
	}
}

func testMarshalList(t *testing.T) {
	lst1 := []float64{1, 2, 3, 4}
	var lst2 []float64

	x := Marshal(lst1)
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

func testMarshalNumberInterface(t *testing.T) {
	var fl = float64(3)
	var intf interface{}

	x := Marshal(fl)
	_ = Unmarshal(x, &intf)

	if fl != intf {
		t.Errorf("got: %v wanted: %v", intf, fl)
	}
}

func testMarshalInterfaceListToMultiVariable(t *testing.T) {
	var lst = []interface{}{5, "a", true}
	var i float64
	var s string
	var b bool
	var ir = float64(5)
	var sr = string("a")
	var br = bool(true)

	lstm := Marshal(lst)
	_ = Unmarshal(lstm, &i, &s, &b)

	if i != ir {
		t.Errorf("got: %f wanted: %f", ir, i)
	}
	if s != sr {
		t.Errorf("got: %s wanted: %s", sr, s)
	}
	if b != br {
		t.Errorf("got: %t wanted: %t", br, b)
	}
}
