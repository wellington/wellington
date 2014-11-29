package context

import "testing"

func TestUnmarshalNumber(t *testing.T) {

	c := float64(1)
	sv := makevalue(c)
	var i float64
	Unmarshal(sv, &i)
	if c != i {
		t.Errorf("got: %d wanted: %d", i, c)
	}

	d := 1.5
	dv := makevalue(d)
	var ed float64
	Unmarshal(dv, &ed)
	if d != ed {
		t.Errorf("got: %d wanted: %d", ed, d)
	}
}

func TestUnmarshalValue(t *testing.T) {
	e := "example"
	input := makevalue(e)
	var s string
	Unmarshal(input, &s)
	if e != s {
		t.Errorf("got: % #v\nwanted: %s", s, e)
	}

	var gsv SassValue
	Unmarshal(input, &gsv)
	if e != gsv {
		t.Errorf("got: % #v\nwanted: %s", gsv, e)
	}

}

func TestUnmarshalComplex(t *testing.T) {
	// Only SassValue supported for lists, is this ok?
	e := []SassValue{"ex1", "ex2"}
	list := makevalue(e)
	var s []SassValue
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
