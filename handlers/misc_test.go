package handlers

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	libsass "github.com/wellington/go-libsass"
)

func TestFontURLFail(t *testing.T) {
	var ignore *os.File
	old := os.Stdout
	defer func() { os.Stdout = old }()
	os.Stdout = ignore
	in := bytes.NewBufferString(`@font-face {
  src: font-url("arial.eot");
}`)
	var out bytes.Buffer
	ctx := libsass.NewContext()
	err := ctx.Compile(in, &out)

	e := "error in C function font-url: font-url: font path not set"
	if !strings.Contains(err.Error(), e) {
		t.Errorf("got:\n%s\nwanted:\n%s\n", err, e)
	}

}

func TestFontURL(t *testing.T) {
	contents := `
$path: font-url($raw: true, $path: "arial.eot");
@font-face {
  src: font-url("arial.eot");
  src: url("#{$path}");
}`
	in := bytes.NewBufferString(contents)
	var out bytes.Buffer
	comp, _, err := setupComp(t, in, &out)
	if err != nil {
		t.Fatal(err)
	}
	e := `@font-face {
  src: url("../font/arial.eot");
  src: url("../font/arial.eot"); }
`

	if e != out.String() {
		t.Errorf("got: %s wanted: %s", out.String(), e)
	}

	comp.Option(libsass.CacheBust("ts"))
	in.WriteString(contents)
	out.Reset()
	err = comp.Run()
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat("../test/font/arial.eot")
	if err != nil {
		t.Fatal(err)
	}
	qs, err := modHash(info)
	if err != nil {
		t.Fatal(err)
	}

	e = fmt.Sprintf(`@font-face {
  src: url("../font/arial.eot%s");
  src: url("../font/arial.eot%s"); }
`, qs, qs)
	if e != out.String() {
		t.Errorf("got: %s wanted: %s", out.String(), e)
	}

}

func TestFontURL_invalid(t *testing.T) {
	r, w, _ := os.Pipe()
	_, _ = r, w
	old := os.Stdout
	defer func() { os.Stdout = old }()
	//os.Stdout = w
	in := bytes.NewBufferString(`@font-face {
  src: font-url(5px);
}`)
	var out bytes.Buffer
	ctx := libsass.NewContext()
	err := ctx.Compile(in, &out)

	e := `Error > stdin:2
error in C function font-url: Invalid Sass type expected: string got: libs.SassNumber value: 5px`
	if !strings.HasPrefix(err.Error(), e) {
		t.Errorf("got:\n%s\nwanted:\n%s", err, e)
	}
}
