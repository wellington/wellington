package wellington

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wellington/wellington/context"
)

func TestHttp(t *testing.T) {
	ctx := context.NewContext()
	hh := http.HandlerFunc(HTTPHandler(ctx))
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
	e := `div p {
  color: red; }
`
	if w.Body.String() != e {
		t.Errorf("got: %s wanted: %s", w.Body.String(), e)
	}
}

func TestHttpError(t *testing.T) {
	ctx := context.NewContext()
	hh := http.HandlerFunc(HTTPHandler(ctx))
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
