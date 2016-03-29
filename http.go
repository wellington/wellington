package wellington

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"time"

	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/wellington/version"
)

// FileHandler starts a file server serving files out of the specified
// build directory.
func FileHandler(gen string) http.Handler {
	abs, err := filepath.Abs(gen)
	if err != nil {
		log.Fatalf("Can not resolve relative path: %s", gen)
	}

	return http.StripPrefix("/build/",
		http.FileServer(http.Dir(abs)),
	)
}

// Response is the object returned on HTTP responses from wellington
type Response struct {
	Contents string    `json:"contents"`
	Start    time.Time `json:"start"`
	Elapsed  string    `json:"elapsed"`
	Error    string    `json:"error"`
	Version  string    `json:"version"`
}

func setDefaultHeaders(w http.ResponseWriter, r *http.Request) {
	// Set headers
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// HTTPHandler starts a CORS enabled web server that takes as input
// Sass and outputs CSS.
func HTTPHandler(gba *BuildArgs, httpPath string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeaders(w, r)
		start := time.Now()
		resp := Response{
			Start:   start,
			Version: version.Version,
		}
		var (
			err  error
			pout bytes.Buffer
		)
		enc := json.NewEncoder(w)
		defer func() {
			resp.Contents = pout.String()
			resp.Elapsed = time.Since(start).String()
			if err != nil {
				resp.Error = err.Error()
			}
			enc.Encode(resp)
		}()
		if r.Body == nil {
			err = errors.New("request is empty")
			return
		}
		defer r.Body.Close()

		comp, err := FromBuildArgs(&pout, nil, r.Body, gba)
		if err != nil {
			resp.Contents = ""
			return
		}
		comp.Option(libsass.HTTPPath(httpPath))
		err = comp.Run()
	}
}
