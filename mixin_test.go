package sprite_sass_test

import (
	"bytes"

	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestMixin(t *testing.T) {

	ib := []byte(`$s: sprite-map("pixel.png");
div {
    background: sprite($s, pixel);
    @include sprite-dimensions($s, pixel);
}`)
	e := `
div {
    background: url("test") 0px 0px;
    width: 1px;
height: 1px;
}`

	in := bytes.NewBuffer(ib)
	p := Parser{}
	output := rerandom.ReplaceAllString(string(p.Start(in, "test")), "")

	if e != output {
		t.Errorf("Mixin parsing failed was:\n%s\nexpected:\n%s", output, e)
	}

	// Base64 encoding changes on every load, so... can't verify it
	// re := regexp.MustCompile("background-image:url\\('data:image\\/png;base64,\\S+='\\)")

	// if !re.Match(bytes) {
	// 	t.Errorf("inline-image failed expected: `%s`, was:`%s`",
	// 		string(bytes))
	// }
}

func TestImageUrl(t *testing.T) {
	ib := []byte(`background: image-url("pixel.png");`)
	e := `background: url("test/pixel.png");`
	in := bytes.NewBuffer(ib)
	p := Parser{}
	output := string(p.Start(in, "test"))

	if e != output {
		t.Errorf("image url failed was:\n%s\nexpected:\n%s\n", output, e)
	}
}
