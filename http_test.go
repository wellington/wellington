package wellington

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func decResp(t *testing.T, r io.Reader) Response {
	dec := json.NewDecoder(r)
	var resp Response
	err := dec.Decode(&resp)
	if err != nil {
		t.Errorf("decodeHTTPResponse: %s", err)
	}
	return resp
}

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
	gba := &BuildArgs{}
	hh := http.HandlerFunc(HTTPHandler(gba))
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

	resp := decResp(t, w.Body)
	if e := "request is empty"; resp.Error != e {
		t.Errorf("got: %s wanted: %s", resp, e)
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
	resp = decResp(t, w.Body)
	if resp.Contents != e {
		t.Errorf("got: %s wanted: %s", resp.Contents, e)
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
	hh := http.HandlerFunc(HTTPHandler(&BuildArgs{}))
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

	e := `Error > stdin:1
required parameter $color is missing in call to Function darken
div { p { color: darken(); } };
`
	resp := decResp(t, w.Body)

	if resp.Error != e {
		t.Errorf("got:\n%s\nwanted:\n%s", resp.Error, e)
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
	resp = decResp(t, w.Body)

	if resp.Contents != e {
		t.Errorf("got:\n%s\nwanted:\n%s", resp.Contents, e)
	}

	// Second run shouldn't have an error in it
}
