package handlers

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/go-libsass/libs"
	"github.com/wellington/spritewell"
	"github.com/wellington/wellington/payload"
)

func init() {
	os.MkdirAll("../test/build/img", 0777)
}

func wrapCallback(sc libsass.HandlerFunc, ch chan libsass.SassValue) libs.SassCallback {
	return libsass.Handler(func(v interface{}, usv libsass.SassValue, rsv *libsass.SassValue) error {
		c := v.(*libsass.Context)
		err := sc(c, usv, rsv)
		ch <- *rsv
		return err
	})
}

func testSprite(t *testing.T, comp libsass.Compiler) {
	paths := comp.(libsass.Pather)
	// Generate test sprite
	imgs := spritewell.New(&spritewell.Options{
		ImageDir:  paths.ImgDir(),
		BuildDir:  paths.BuildDir(),
		GenImgDir: paths.ImgBuildDir(),
	})
	glob := "*.png"
	err := imgs.Decode(glob)
	if err != nil {
		t.Fatal(err)
	}

}

func setupComp(t *testing.T, r io.Reader, out io.Writer) (libsass.Compiler, error) {

	comp, err := libsass.New(out, r,
		libsass.OutputStyle(libsass.NESTED_STYLE),
		libsass.BuildDir("../test/build"),
		libsass.ImgDir("../test/img"),
		libsass.FontDir("../test/font"),
		libsass.Payload(payload.New()),
		libsass.ImgBuildDir("../test/build/img"),
	)
	if err != nil {
		t.Fatal(err)
	}
	testSprite(t, comp)

	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(5 * time.Second):
			t.Fatal("setupComp timeout")
		case <-done:
			return
		}
	}()

	err = comp.Run()
	close(done)
	return comp, err
}

func oldContext() *libsass.Context {
	ctx := libsass.NewContext()
	ctx.Payload = payload.New()
	return ctx
}

func setupCtx(t *testing.T, r io.Reader, out io.Writer /*, cookies ...libsass.Cookie*/) (*libsass.Context, libsass.SassValue, error) {
	var usv libsass.SassValue

	comp, err := libsass.New(out, r,
		libsass.OutputStyle(libsass.NESTED_STYLE),
		libsass.BuildDir("../test/build"),
		libsass.ImgDir("../test/img"),
		libsass.FontDir("../test/font"),
		libsass.ImgBuildDir("../test/build/img"),
		libsass.Payload(payload.New()),
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx := comp.Context()

	testSprite(t, comp)

	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(5 * time.Second):
			t.Fatal("setupCtx timeout")
		case <-done:
			return
		}
	}()

	err = ctx.Compile(r, out)
	close(done)
	return ctx, usv, err
}

func TestFuncImageURL(t *testing.T) {
	comp, err := libsass.New(nil, nil,
		libsass.BuildDir("../test/build"),
		libsass.ImgDir("../test/img"),
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx := libsass.NewCompilerContext(comp)

	usv, _ := libsass.Marshal([]string{"139.png"})
	rsv, err := ImageURL(ctx, usv)
	if err != nil {
		t.Fatal(err)
	}
	var path string
	err = libsass.Unmarshal(*rsv, &path)
	if err != nil {
		t.Fatal(err)
	}
	if e := "url('../img/139.png')"; e != path {
		t.Errorf("got: %s wanted: %s", path, e)
	}

	// Test sending invalid date to imageURL
	usv, _ = libsass.Marshal(libs.SassNumber{Value: 1, Unit: "px"})
	_, err = ImageURL(ctx, usv)
	if err == nil {
		t.Fatal("error is nil")
	}

	e := "Invalid Sass type expected: slice got: libs.SassNumber value: 1px"

	if e != err.Error() {
		t.Errorf("got: %s wanted: %s", err, e)
	}

}

type mockImg struct {
	r io.ReadCloser
}

func (m mockImg) Do(s string) (io.ReadCloser, error) {
	return m.r, nil
}

func TestCompile_HTTP_InlineImage(t *testing.T) {
	oldImageResolver := imgResolver
	defer func() {
		imgResolver = oldImageResolver
	}()

	f, err := os.Open("../test/img/pixel/1x1.png")
	if err != nil {
		t.Fatal(err)
	}
	// overwrite the defaultimageresolver
	imgResolver = mockImg{r: f}

	in := bytes.NewBufferString(`div {
  background: #602d6c no-repeat inline-image("http://example.com/pixel/1x1.png");
}`)

	var out bytes.Buffer
	_, err = setupComp(t, in, &out)
	if err != nil {
		t.Error(err)
	}
	exp := `div {
  background: #602d6c no-repeat url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAMAAAAoyzS7AAAAA1BMVEX/TQBcNTh/AAAAAXRSTlMz/za5cAAAAA5JREFUeJxiYgAEAAD//wAGAAP60FmuAAAAAElFTkSuQmCC"); }
`
	if exp != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), exp)
	}
}

func TestFuncImageHeight(t *testing.T) {
	in := bytes.NewBufferString(`div {
    height: image-height("139");
}`)
	var out bytes.Buffer
	_, err := setupComp(t, in, &out)

	if err != nil {
		t.Error(err)
	}

	e := `div {
  height: 139px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestRegImageWidth(t *testing.T) {
	in := bytes.NewBufferString(`div {
    height: image-width("139");
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(t, in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  height: 96px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestRegSpriteImageHeight(t *testing.T) {
	in := bytes.NewBufferString(`$map: sprite-map("*.png");
div {
  height: image-height(sprite-file($map,"139"));
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(t, in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  height: 139px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestRegSpriteImageWidth(t *testing.T) {
	in := bytes.NewBufferString(`$map: sprite-map("*.png");
div {
  width: image-width(sprite-file($map,"139"));
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(t, in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  width: 96px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}

}

func TestImageUrl_bustsmallfile(t *testing.T) {
	// Ensure cache busting works for files smaller than header buffer
	contents := `
div {
    background: image-url("pixel/1x1.png");
}`
	in := bytes.NewBufferString(contents)
	var out bytes.Buffer
	comp, err := setupComp(t, in, &out)
	if err != nil {
		t.Error(err)
	}
	// This is annoying, but no way to configure compiler
	in.WriteString(contents)
	out.Reset()
	comp.Option(libsass.CacheBust("sum"))
	err = comp.Run()
	if err != nil {
		t.Fatal(err)
	}
	e := `div {
  background: url('../img/pixel/1x1.png?7b29fcd7'); }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestRegImageURL(t *testing.T) {
	contents := `
div {
    background: image-url("139.png");
}`
	in := bytes.NewBufferString(contents)
	var out bytes.Buffer
	comp, err := setupComp(t, in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  background: url('../img/139.png'); }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
	out.Reset()
	comp.Option(libsass.CacheBust("ts"))

	in.WriteString(contents)
	err = comp.Run()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat("../test/img/139.png")
	if err != nil {
		t.Fatal(err)
	}
	qry, err := modHash(info)
	if err != nil {
		t.Fatal(err)
	}
	e = fmt.Sprintf(`div {
  background: url('../img/139.png%s'); }
`, qry)

	if e != out.String() {
		t.Errorf("got: %s\nwanted: %s", out.String(), e)
	}

	out.Reset()
	comp.Option(libsass.CacheBust("sum"))

	in.WriteString(contents)
	err = comp.Run()
	if err != nil {
		t.Fatal(err)
	}
	e = `div {
  background: url('../img/139.png?194da102'); }
`

	if e != out.String() {
		t.Errorf("got: %s\nwanted: %s", out.String(), e)
	}

}

func TestRegInlineImage(t *testing.T) {
	in := bytes.NewBufferString(`
div {
    background: inline-image("pixel/1x1.png");
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(t, in, &out)
	if err != nil {
		t.Error(err)
	}
	e := `div {
  background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAMAAAAoyzS7AAAAA1BMVEX/TQBcNTh/AAAAAXRSTlMz/za5cAAAAA5JREFUeJxiYgAEAAD//wAGAAP60FmuAAAAAElFTkSuQmCC"); }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestInlineImage_nofile(t *testing.T) {
	in := bytes.NewBufferString(`
div {
    background: inline-image("pixel/nofile.png");
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(t, in, &out)
	if err == nil {
		t.Error("No error thrown for missing file")
	}
	e := `Error > stdin:3
error in C function inline-image: open ../test/img/pixel/nofile.png: no such file or directory

Backtrace:
	stdin:3, in function ` + "`inline-image`" + `
	stdin:3

div {
    background: inline-image("pixel/nofile.png");
}
`
	if e != err.Error() {
		t.Errorf("got:\n%s\nwanted:\n%s", err.Error(), e)
	}

}

func TestHandle_unknownmap(t *testing.T) {
	t.Skip("")
	in := bytes.NewBufferString(`
$map: sprite-map("*.png");
div {
  background: sprite("nomap", "140");
}`)
	var out bytes.Buffer
	_, _, err := setupCtx(t, in, &out)
	if err == nil {
		t.Error("no error thrown for invalid map")
	}
	_ = out

	e := `Error > stdin:4
error in C function sprite: Variable not found matching glob: nomap sprite:140

Backtrace:
	stdin:4, in function ` + "`sprite`" + `
	stdin:4

$map: sprite-map("*.png");
div {
  background: sprite("nomap", "140");
}
`

	if e != err.Error() {
		t.Errorf("got:\n%s\nwanted:\n%s", err.Error(), e)
	}
}

func ExampleSprite() {
	in := bytes.NewBufferString(`
$map: sprite-map("*.png", 10px); // One argument
div {
  background: sprite($map, "140");
}`)

	ctx := oldContext()

	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"
	var out bytes.Buffer
	err := ctx.Compile(in, &out)
	if err != nil {
		fmt.Println(err)
	}

	io.Copy(os.Stdout, &out)

	// Output:
	// div {
	//   background: url("img/4f0c6e.png") 0px -149px; }
}

func TestHandle_offset(t *testing.T) {
	in := bytes.NewBufferString(`
$map: sprite-map("*.png", 10px);
div {
  background: sprite($map, "140", 10px, 10px);
}`)

	ctx := oldContext()
	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"
	var out bytes.Buffer
	err := ctx.Compile(in, &out)

	if err != nil {
		t.Fatal("expected error")
	}

	e := `div {
  background: url("img/4f0c6e.png") 10px -139px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestHandle_erroroffset(t *testing.T) {
	in := bytes.NewBufferString(`
$map: sprite-map("*.png", 10px);
div {
  background: sprite($map, "140", 0, 0);
}`)

	ctx := oldContext()
	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"
	var out bytes.Buffer
	err := ctx.Compile(in, &out)

	e := `Error > stdin:4
error in C function sprite: Please specify unit for offset ie. (2px)

Backtrace:
	stdin:4, in function ` + "`sprite`" + `
	stdin:4

$map: sprite-map("*.png", 10px);
div {
  background: sprite($map, "140", 0, 0);
}
`
	if err == nil {
		t.Fatal("expected error")
	}

	if e != err.Error() {
		t.Errorf("got:\n%s\nwanted:\n%s", err.Error(), e)
	}

}

func TestSpriteHTTP(t *testing.T) {
	in := bytes.NewBufferString(`
$map: sprite-map("*.png", 10px);
div {
  background: sprite($map, "140");
}`)

	ctx := oldContext()
	ctx.IncludePaths = []string{"../test"}
	ctx.HTTPPath = "http://foo.com"
	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"
	var out bytes.Buffer
	err := ctx.Compile(in, &out)

	if err != nil {
		t.Error(err)
	}
	e := `div {
  background: url("http://foo.com/build/4f0c6e.png") 0px -149px; }
`
	if e != out.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}

}

func TestSpriteMany(t *testing.T) {

	in := bytes.NewBufferString(`
$map: sprite-map("many/*.jpg", 0px);
div {
  background: sprite($map, "bird");
  background: sprite($map, "in");
  background: sprite($map, "pencil");
  background: sprite($map, "rss");
  background: sprite($map, "twitt");
}`)

	ctx := oldContext()

	ctx.BuildDir = "../test/build"
	ctx.GenImgDir = "../test/build/img"
	ctx.ImageDir = "../test/img"
	var out bytes.Buffer
	err := ctx.Compile(in, &out)

	if err != nil {
		t.Error(err)
	}

	e := `div {
  background: url("img/744d97.png") 0px 0px;
  background: url("img/744d97.png") 0px -150px;
  background: url("img/744d97.png") 0px -300px;
  background: url("img/744d97.png") 0px -450px;
  background: url("img/744d97.png") 0px -600px; }
`

	if out.String() != e {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestInlineSVG(t *testing.T) {
	var in, out bytes.Buffer
	in.WriteString(`div {
  background-image: inline-image("hexane.svg");
}`)
	_, _, err := setupCtx(t, &in, &out)
	if err != nil {
		t.Error(err)
	}

	e := `div {
  background-image: url("data:image/svg+xml;utf8,%3C%3Fxml%20version=%221.0%22%20encoding=%22UTF-8%22%20standalone=%22no%22%3F%3E%3Csvg%20xmlns=%22http://www.w3.org/2000/svg%22%20version=%221.0%22%20width=%22480%22%20height=%22543.03003%22%20viewBox=%220%200%20257.002%20297.5%22%20xml:space=%22preserve%22%3E%3Cg%20transform=%22matrix%280.8526811,0,0,0.8526811,18.930632,21.913299%29%22%3E%3Cpolygon%20points=%228.003,218.496%200,222.998%200,74.497%208.003,78.999%208.003,218.496%20%22/%3E%3Cpolygon%20points=%22128.501,287.998%20128.501,297.5%200,222.998%208.003,218.496%20128.501,287.998%20%22%20/%3E%3Cpolygon%20points=%22249.004,218.496%20257.002,222.998%20128.501,297.5%20128.501,287.998%20249.004,218.496%20%22%20/%3E%3Cpolygon%20points=%22249.004,78.999%20257.002,74.497%20257.002,222.998%20249.004,218.496%20249.004,78.999%20%22%20/%3E%3Cpolygon%20points=%22128.501,9.497%20128.501,0%20257.002,74.497%20249.004,78.999%20128.501,9.497%20%22%20/%3E%3Cpolygon%20points=%228.003,78.999%200,74.497%20128.501,0%20128.501,9.497%208.003,78.999%20%22%20/%3E%3C/g%3E%3C/svg%3E"); }
`

	if out.String() != e {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}

	in.Reset()
	out.Reset()
	in.WriteString(`div {
  background-image: inline-image("hexane.svg", $encode: true);
}`)
	_, _, err = setupCtx(t, &in, &out)
	if err != nil {
		t.Error(err)
	}

	e = `div {
  background-image: url("data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiIHN0YW5kYWxvbmU9Im5vIj8+DQo8c3ZnIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgdmVyc2lvbj0iMS4wIiB3aWR0aD0iNDgwIiBoZWlnaHQ9IjU0My4wMzAwMyIgdmlld0JveD0iMCAwIDI1Ny4wMDIgMjk3LjUiIHhtbDpzcGFjZT0icHJlc2VydmUiPg0KPGcgdHJhbnNmb3JtPSJtYXRyaXgoMC44NTI2ODExLDAsMCwwLjg1MjY4MTEsMTguOTMwNjMyLDIxLjkxMzI5OSkiPg0KPHBvbHlnb24gcG9pbnRzPSI4LjAwMywyMTguNDk2IDAsMjIyLjk5OCAwLDc0LjQ5NyA4LjAwMyw3OC45OTkgOC4wMDMsMjE4LjQ5NiAiLz4NCjxwb2x5Z29uIHBvaW50cz0iMTI4LjUwMSwyODcuOTk4IDEyOC41MDEsMjk3LjUgMCwyMjIuOTk4IDguMDAzLDIxOC40OTYgMTI4LjUwMSwyODcuOTk4ICIgLz4NCjxwb2x5Z29uIHBvaW50cz0iMjQ5LjAwNCwyMTguNDk2IDI1Ny4wMDIsMjIyLjk5OCAxMjguNTAxLDI5Ny41IDEyOC41MDEsMjg3Ljk5OCAyNDkuMDA0LDIxOC40OTYgIiAvPg0KPHBvbHlnb24gcG9pbnRzPSIyNDkuMDA0LDc4Ljk5OSAyNTcuMDAyLDc0LjQ5NyAyNTcuMDAyLDIyMi45OTggMjQ5LjAwNCwyMTguNDk2IDI0OS4wMDQsNzguOTk5ICIgLz4NCjxwb2x5Z29uIHBvaW50cz0iMTI4LjUwMSw5LjQ5NyAxMjguNTAxLDAgMjU3LjAwMiw3NC40OTcgMjQ5LjAwNCw3OC45OTkgMTI4LjUwMSw5LjQ5NyAiIC8+DQo8cG9seWdvbiBwb2ludHM9IjguMDAzLDc4Ljk5OSAwLDc0LjQ5NyAxMjguNTAxLDAgMTI4LjUwMSw5LjQ5NyA4LjAwMyw3OC45OTkgIiAvPg0KPC9nPg0KPC9zdmc+"); }
`

	if out.String() != e {
		t.Errorf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

// func BenchmarkSprite(b *testing.B) {
// 	ctx := oldContext()
// 	ctx.BuildDir = "context/test/build"
// 	ctx.GenImgDir = "context/test/build/img"
// 	ctx.ImageDir = "context/test/img"
// 	// Add real arguments when sass lists can be [un]marshalled
// 	lst := []interface{}{"*.png", libs.SassNumber{Value: 5, Unit: "px"}}
// 	usv, _ := libsass.Marshal(lst)

// 	var rsv libsass.SassValue
// 	for i := 0; i < b.N; i++ {
// 		SpriteMap(ctx, usv, &rsv)
// 	}
// 	// Debug if needed
// 	// var s string
// 	// Unmarshal(usv, &s)
// 	// fmt.Println(s)
// }

func TestImgResolver(t *testing.T) {

	e := "Hello, client"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,
		r *http.Request) {
		w.Write([]byte(e))
	}))
	defer ts.Close()

	r, err := imgResolver.Do(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	bs, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if string(bs) != string(e) {
		t.Fatalf("got: %s was: %s", bs, e)
	}

	// Test some failure conditions
	_, err = imgResolver.Do("/data/img")
	if err == nil {
		t.Fatal(err)
	}

	e = `invalid image URL: /data/img`
	if err.Error() != e {
		t.Errorf("got: %s wanted: %s", err, e)
	}

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,
		r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	_, err = imgResolver.Do(ts.URL)
	if err == nil {
		t.Fatal("expected error")
	}

	if e := "could not resolve image: 500"; err.Error() != e {
		t.Errorf("got: %s wanted: %s", err, e)
	}
}
