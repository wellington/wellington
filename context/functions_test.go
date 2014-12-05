package context

import "testing"

func TestImageUrl(t *testing.T) {
	ctx := Context{
		BuildDir: "test/build",
		ImageDir: "test/image",
	}

	usv, _ := Marshal("image.png")
	usv = ImageUrl(&ctx, usv)
	var path string
	Unmarshal(usv, &path)

	if e := "../image/image.png"; e != path {
		t.Errorf("got: %s wanted: %s", path, e)
	}
}
