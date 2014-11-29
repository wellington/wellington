package context

import "testing"

func TestUnmarshal(t *testing.T) {
	e := "example"
	input := makevalue("string", e)

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
