package context

import "testing"

func TestSampleCB(t *testing.T) {
	ctx := NewContext()
	ctx.Cookies = make([]Cookie, 1)
	usv := Marshal(float64(1))
	usv = SampleCB(ctx, usv)

	var b bool
	err := Unmarshal(usv, &b)
	if err != nil {
		panic(err)
	}
	if e := false; b != e {
		t.Errorf("wanted: %t got: %t", e, b)
	}
}
