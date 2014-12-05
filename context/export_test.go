package context

import "testing"

func TestSampleCB(t *testing.T) {
	ctx := NewContext()
	ctx.Cookies = make([]Cookie, 1)
	usv, err := Marshal(float64(1))
	if err != nil {
		t.Error(err)
	}
	usv = SampleCB(ctx, usv)

	var b bool
	err = Unmarshal(usv, &b)
	if err != nil {
		t.Error(err)
	}
	if e := false; b != e {
		t.Errorf("wanted: %t got: %t", e, b)
	}
}
