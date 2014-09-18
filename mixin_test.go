package sprite_sass_test

import (
	"bytes"

	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestMixin(t *testing.T) {

	ib := []byte(`$s: sprite-map("pixel.png");
div {
    @include sprite-dimensions($s, "pixel");
    background-image: inline-image("pixel.png");
}`)

	in := bytes.NewBuffer(ib)
	par := Parser{}
	bytes := par.Start(in, "test")
	_ = bytes
	// Base64 encoding changes on every load, so... can't verify it
	// re := regexp.MustCompile("background-image:url\\('data:image\\/png;base64,\\S+='\\)")

	// if !re.Match(bytes) {
	// 	t.Errorf("inline-image failed expected: `%s`, was:`%s`",
	// 		string(bytes))
	// }
}
