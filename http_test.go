package wellington

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/wellington/wellington/context"
)

func TestFileHandler(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TempDir: %v", err)
	}
	e := "Hello world"
	if err := ioutil.WriteFile(filepath.Join(tempDir, "foo.txt"),
		[]byte(e), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ts := httptest.NewServer(FileHandler(tempDir))
	defer ts.Close()
	get := func(suffix string) string {
		res, err := http.Get(ts.URL + suffix)
		if err != nil {
			t.Fatalf("Get %s: %v", suffix, err)
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("ReadAll %s: %v", suffix, err)
		}
		res.Body.Close()
		return string(b)
	}

	if s := get("/build/foo.txt"); e != s {
		t.Fatalf("got %q want %q", s, e)
	}

}

func TestHTTPHandler(t *testing.T) {
	ctx := context.NewContext()
	hh := http.HandlerFunc(HTTPHandler(ctx))
	req, err := http.NewRequest("GET", "", nil)
	req.Header.Set("Origin", "http://foo.com")
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

	ehead := map[string][]string{
		"Access-Control-Allow-Origin":      []string{"http://foo.com"},
		"Access-Control-Allow-Methods":     []string{"POST, GET, OPTIONS, PUT, DELETE"},
		"Access-Control-Allow-Headers":     []string{"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token"},
		"Access-Control-Allow-Credentials": []string{"true"},
	}

	for i, h := range w.Header() {

		if ehead[i][0] != h[0] {
			t.Errorf("got:\n%q\nwanted:\n%q", h, ehead[i])
		}
	}

}

func TestHTTPHandler_error(t *testing.T) {
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
