package handlers

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	libsass "github.com/wellington/go-libsass"
)

func TestError_warn(t *testing.T) {
	var pout bytes.Buffer
	log.SetFlags(0)
	log.SetOutput(&pout)
	// Disabled while new warn integration is built
	in := bytes.NewBufferString(`
@warn "!";
div {
  color: red;
}`)
	var out bytes.Buffer
	ctx := libsass.Context{}
	err := ctx.Compile(in, &out)

	if err != nil {
		t.Error(err)
	}

	e := `"\x1b[33mWARNING: !\x1b[0m\n"`

	qout := fmt.Sprintf("%q", pout.String())

	if e != qout {
		t.Errorf("got:\n%s\nwanted:\n%s", qout, e)
	}
	log.SetOutput(os.Stdout)
}
