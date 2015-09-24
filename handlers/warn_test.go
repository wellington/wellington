package handlers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	libsass "github.com/wellington/go-libsass"
)

func TestError_warn(t *testing.T) {

	oo := os.Stdout
	r, w, _ := os.Pipe()
	defer w.Close()
	os.Stdout = w

	// Disabled while new warn integration is built
	in := bytes.NewBufferString(`@warn "!";
div { color: red; }`)
	ctx := libsass.NewContext()
	libsass.RegisterHandler("@warn", WarnHandler)
	var empty bytes.Buffer
	err := ctx.Compile(in, &empty)
	if err != nil {
		t.Error(err)
	}

	e := `WARNING: !`
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = oo
	out := <-outC
	qout := fmt.Sprintf("%q", out)

	if !strings.Contains(qout, e) {
		t.Errorf("got:\n%s\nwanted:\n%s", qout, e)
	}
	os.Stdout = oo
}
