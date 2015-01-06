package main

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/wellington/wellington/context"
)

func TestStdinImport(t *testing.T) {
	fh, err := os.Open("../test/sass/import.scss")
	if err != nil {
		t.Error(err)
	}

	oldStd := os.Stdin
	oldOut := os.Stdout

	os.Stdin = fh
	r, w, _ := os.Pipe()
	os.Stdout = w
	flag.Set("p", "../test/sass")
	main()

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdin = oldStd
	os.Stdout = oldOut

	out := <-outC

	e := `Reading from stdin, -h for help
/* line 8, stdin */
div {
  background: #00FF00;
  font-size: 10pt; }
`

	if e != out {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestStdinSprite(t *testing.T) {
	fh, err := os.Open("../test/sass/sprite.scss")
	if err != nil {
		t.Error(err)
	}

	oldStd := os.Stdin
	oldOut := os.Stdout

	os.Stdin = fh
	r, w, _ := os.Pipe()
	os.Stdout = w
	flag.Set("dir", "../test/img")
	flag.Set("gen", "../test/img/build")
	main()

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdin = oldStd
	os.Stdout = oldOut

	out := <-outC

	e := `Reading from stdin, -h for help
/* line 8, stdin */
div {
  height: 139px;
  width: 96px;
  background: url("../test/img/build/91300a.png") -0px -0px; }
`
	if e != out {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestHttp(t *testing.T) {
	ctx := context.NewContext()
	hh := http.HandlerFunc(httpHandler(ctx))
	// nil causes panic, is this a problem?
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	hh.ServeHTTP(w, req)

	if e := 200; w.Code != e {
		t.Errorf("got: %d wanted: %d", w.Code, e)
	}

	if e := "input is empty"; w.Body.String() != e {
		t.Errorf("got: %s wanted: %s", w.Body.String(), e)
	}

	req, err = http.NewRequest("GET", "", bytes.NewBufferString(`div { p { color: red; } }`))
	if err != nil {
		t.Error(err)
	}
	w.Body.Reset()
	hh.ServeHTTP(w, req)

	if e := 200; w.Code != e {
		t.Errorf("got: %d wanted: %d", w.Code, e)
	}
	e := `div p {
  color: red; }
`
	if w.Body.String() != e {
		t.Errorf("got: %s wanted: %s", w.Body.String(), e)
	}
}

func TestHttpError(t *testing.T) {
	ctx := context.NewContext()
	hh := http.HandlerFunc(httpHandler(ctx))
	// nil causes panic, is this a problem?
	req, err := http.NewRequest("GET", "",
		bytes.NewBufferString(`div { p { color: darken(); } };`))
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	hh.ServeHTTP(w, req)

	if e := 200; w.Code != e {
		t.Errorf("got: %d wanted: %d", w.Code, e)
	}

	e := `Error > stdin:6
required parameter $color is missing in call to function darken
@mixin sprite-dimensions($map, $name) {
  $file: sprite-file($map, $name);
  height: image-height($file);
  width: image-width($file);
}
div { p { color: darken(); } };
`
	if w.Body.String() != e {
		t.Errorf("got:\n%s\nwanted:\n%s", w.Body.String(), e)
	}

	req, err = http.NewRequest("GET", "",
		bytes.NewBufferString(`div { p { color: red; } }`))
	if err != nil {
		t.Error(err)
	}
	w.Body.Reset()
	hh.ServeHTTP(w, req)

	if e := 200; w.Code != e {
		t.Errorf("got: %d wanted: %d", w.Code, e)
	}
	e = `div p {
  color: red; }
`
	if w.Body.String() != e {
		t.Errorf("got:\n%s\nwanted:\n%s", w.Body.String(), e)
	}

	// Second run shouldn't have an error in it
}

func TestFile(t *testing.T) {
	// TODO: Tests for file importing here
}
