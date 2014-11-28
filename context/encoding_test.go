package context

import "testing"

func TestUnmarshal(t *testing.T) {
	e := "example"
	input := makevalue("string", e)
	var s string
	Decode(input, &s)
	if e != s {
		t.Errorf("got: %s\nwanted: %s", s, e)
	}
}
