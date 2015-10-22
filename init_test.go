package wellington

import "testing"

func TestPayload(t *testing.T) {

	load := newPayload()
	if load.s == nil {
		t.Fatal("image payload missing")
	}

	if load.i == nil {
		t.Fatal("image payload missing")
	}

	if load.Sprite() == nil {
		t.Fatal("can not retrieve sprite map")
	}

	if load.Image() == nil {
		t.Fatal("can not retrieve image map")
	}
}
