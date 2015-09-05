package libsass

import (
	"strconv"
	"sync"

	"github.com/wellington/go-libsass/libs"
)

func (ctx *Context) SetHeaders(opts libs.SassOptions) {
	// Push the headers into the local array
	for _, gh := range globalHeaders {
		if !ctx.Headers.Has(gh) {
			ctx.Headers.Add(gh)
		}
	}

	// Loop through headers creating ImportEntry
	entries := make([]libs.ImportEntry, ctx.Headers.Len())
	for i, ent := range ctx.Headers.h {
		uniquename := "hdr" + strconv.FormatInt(int64(i), 10)
		entries[i] = libs.ImportEntry{
			// Each entry requires a unique identifier
			// https://github.com/sass/libsass/issues/1292
			Path:   uniquename,
			Source: ent.Content,
			SrcMap: "",
		}
	}
	libs.BindHeader(opts, entries)
}

type Header struct {
	Content string
}

type Headers struct {
	sync.RWMutex
	h []Header
}

func (h *Headers) Add(s string) {
	h.Lock()
	defer h.Unlock()

	h.h = append(h.h, Header{
		Content: s,
	})
}

func (h *Headers) Has(s string) bool {
	for _, c := range h.h {
		if s == c.Content {
			return true
		}
	}
	return false
}

func (h *Headers) Len() int {
	return len(h.h)
}

var globalHeaders []string

// RegisterHeader fifo
func RegisterHeader(body string) {
	globalHeaders = append(globalHeaders, body)
}
