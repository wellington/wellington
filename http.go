package wellington

import (
	"bytes"
	"io"
	"net/http"

	"github.com/wellington/wellington/context"
)

func HTTPHandler(ctx *context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var pout bytes.Buffer

		// Set headers
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		_, err := StartParser(ctx, r.Body, &pout, NewPartialMap())
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}

		err = ctx.Compile(&pout, w)
		if err != nil {
			io.WriteString(w, err.Error())
		}
	}
}
