package wellington

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/wellington/wellington/context"
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
}

// HTTPHandler starts a CORS enabled web server that takes as input
// Sass and outputs CSS.
func HTTPHandler(ctx *context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			pout bytes.Buffer
			buf  bytes.Buffer
		)
		start := time.Now()
		// Set headers
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		_, err := StartParser(ctx, r.Body, &pout, NewPartialMap())
		if err != nil {
			enc := json.NewEncoder(w)
			enc.Encode(Response{
				Start: start,
				Elapsed: strconv.FormatFloat(float64(
					time.Since(start).Nanoseconds())/float64(time.Millisecond),
					'f', 3, 32) + "ms",
				Contents: "",
				Error:    fmt.Sprintf("%s", err),
			})
			return
		}
		err = ctx.Compile(&pout, &buf)
		defer func() {
			enc := json.NewEncoder(w)
			errString := ""
			if err != nil {
				errString = err.Error()
			}
			enc.Encode(Response{
				Start: start,
				Elapsed: strconv.FormatFloat(float64(
					time.Since(start).Nanoseconds())/(1000*1000),
					'f', 3, 32) + "ms",
				Contents: buf.String(),
				Error:    errString,
			})
		}()
	}
}
