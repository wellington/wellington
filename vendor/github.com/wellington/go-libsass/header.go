package libsass

import (
	"strconv"
	"sync"

	"github.com/wellington/go-libsass/libs"
)

var globalHeaders []string

// RegisterHeader fifo
func RegisterHeader(body string) {
	ghMu.Lock()
	globalHeaders = append(globalHeaders, body)
	ghMu.Unlock()
}

type Header struct {
	idx     *string
	Content string
}

type Headers struct {
	wg      sync.WaitGroup
	closing chan struct{}
	sync.RWMutex
	h []Header
	// idx is a pointer for libsass to lookup these Headers
	idx int
}

// NewHeaders instantiates a Headers for prefixing Sass to input
// See: https://github.com/sass/libsass/wiki/API-Sass-Importer
func NewHeaders() *Headers {
	h := &Headers{
		closing: make(chan struct{}),
	}
	return h
}

func (hdrs *Headers) Bind(opts libs.SassOptions) {
	// Push the headers into the local array
	ghMu.RLock()
	for _, gh := range globalHeaders {
		if !hdrs.Has(gh) {
			hdrs.Add(gh)
		}
	}
	ghMu.RUnlock()

	// Loop through headers creating ImportEntry
	entries := make([]libs.ImportEntry, hdrs.Len())
	hdrs.RLock()
	for i, ent := range hdrs.h {
		uniquename := "hdr" + strconv.FormatInt(int64(i), 10)
		entries[i] = libs.ImportEntry{
			// Each entry requires a unique identifier
			// https://github.com/sass/libsass/issues/1292
			Path:   uniquename,
			Source: ent.Content,
			SrcMap: "",
		}
	}
	hdrs.RUnlock()
	// Preserve reference to libs address to these entries
	hdrs.idx = libs.BindHeader(opts, entries)
}

func (hdrs *Headers) Close() {
	// Clean up memory reserved for headers
	libs.RemoveHeaders(hdrs.idx)
	close(hdrs.closing)
	hdrs.wg.Wait()
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
