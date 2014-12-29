package wellington

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestLoadAndBuild(t *testing.T) {
	oo := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	err := LoadAndBuild("test/sass/file.scss", &BuildArgs{}, NewPartialMap())
	if err != nil {
		t.Error(err)
	}
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = oo
	out := <-outC

	e := `div {
  color: black; }
Rebuilt: test/sass/file.scss
`
	if e != out {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}
}
