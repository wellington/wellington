package handlers

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	libsass "github.com/wellington/go-libsass"
)

func TestError_warn(t *testing.T) {
	oo := os.Stdout
	defer func() {
		os.Stdout = oo
	}()

	r, w, _ := os.Pipe()
	defer w.Close()
	os.Stdout = w

	// Disabled while new warn integration is built
	in := bytes.NewBufferString(`@warn "!";
div { color: red; }`)

	libsass.RegisterHandler("@warn", WarnHandler)

	var out bytes.Buffer
	comp, err := libsass.New(&out, in,
		libsass.OutputStyle(libsass.NESTED_STYLE),
		libsass.BuildDir("../test/build"),
		libsass.ImgDir("../test/img"),
		libsass.ImgBuildDir("../test/build/img"),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = comp.Run()
	if err != nil {
		t.Fatal(err)
	}

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()
	w.Close()

	warnout := <-outC
	if len(warnout) == 0 {
		t.Fatal("no error reported")
	}
	e := `WARNING: !`
	if !strings.Contains(warnout, e) {
		t.Errorf("got: %q wanted: %q", warnout, e)
	}
}
