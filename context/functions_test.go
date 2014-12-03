package context

import (
	"io"
	"testing"
)

func wrapCallback(sc SassCallback, ch chan UnionSassValue) SassCallback {
	return func(c *Context, usv UnionSassValue) UnionSassValue {
		usv = sc(c, usv)
		ch <- usv
		return usv
	}
}

func setupCtx(r io.Reader, out io.Writer, cookies ...Cookie) (Context, chan UnionSassValue, chan error) {
	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		BuildDir:     "test/build",
		ImageDir:     "test/img",
		GenImgDir:    "test/build/img",
		Out:          "",
	}
	var cc chan UnionSassValue
	ec := make(chan error, 1)
	_ = ec
	// If callbacks were made, add them to the context
	// and create channels for communicating with them.
	if len(cookies) > 0 {
		cs := make([]Cookie, len(cookies))
		for i, c := range cookies {
			cs[i] = Cookie{
				c.sign,
				wrapCallback(c.fn, cc),
				&ctx,
			}
		}
	}

	ec <- ctx.Compile(r, out)

	return ctx, cc, ec
}

func TestImageUrl(t *testing.T) {
	ctx := Context{
		BuildDir: "test/build",
		ImageDir: "test/img",
	}

	usv, _ := Marshal("image.png")
	usv = ImageUrl(&ctx, usv)
	var path string
	Unmarshal(usv, &path)

	if e := "../img/image.png"; e != path {
		t.Errorf("got: %s wanted: %s", path, e)
	}
}
