package context

import "testing"

func TestUnmarshal(t *testing.T) {
	e := "example"
	input := makevalue(e)

	var s string
	Unmarshal(input, &s)
	if e != s {
		t.Errorf("got: % #v\nwanted: %s", s, e)
	}

	var sv SassValue
	Unmarshal(input, &sv)
	if e != sv {
		t.Errorf("got: % #v\nwanted: %s", sv, e)
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
