package sprite_sass_test

import (
	"bytes"

	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestMixin(t *testing.T) {

	ib := []byte(`$s: sprite-map("*.png");
div {
    @include sprite-dimensions($s, "139");
    background-image: inline-image("139.png");
}`)
	ob := []byte(`div {
    width: 96px;
height: 139px;
background-image:url('data:image/png;base64,iVBORw0KGg');
}`)

	in := bytes.NewBuffer(ib)
	par := Parser{}
	bytes := par.Start(in, "test")

	if string(bytes) != string(ob) {
		t.Errorf("inline-image failed expeced:%s\nwas:", string(ib), string(ob))
	}
}
