package context

// #include <stdio.h>
// #include "sass_context.h"
//
// extern struct Sass_Import** HeaderBridge(void* cookie);
//
// Sass_Import_List SassHeaders(const char* cur_path, Sass_Importer_Entry cb, struct Sass_Compiler* comp)
// {
//   void* cookie = sass_importer_get_cookie(cb);
//   Sass_Import_List list = HeaderBridge(cookie);
//   return list;
// }
//
import "C"

import (
	"reflect"
	"sync"
	"unsafe"
)

func (ctx *Context) SetHeaders(opts *C.struct_Sass_Options) {
	// Push the headers into the local array
	for _, gh := range globalHeaders {
		ctx.Headers.Add(gh)
	}

	cheads := C.sass_make_importer_list(1)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cheads)),
		Len:  1, Cap: 1,
	}
	goheads := *(*[]C.Sass_Importer_Entry)(unsafe.Pointer(&hdr))

	imper := C.sass_make_importer(
		C.Sass_Importer_Fn(C.SassHeaders),
		C.double(0),
		unsafe.Pointer(ctx))

	goheads[0] = imper

	C.sass_option_set_c_headers(opts, cheads)
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

func (h *Headers) Len() int {
	return len(h.h)
}

var globalHeaders []string

// RegisterHeader fifo
func RegisterHeader(body string) {
	globalHeaders = append(globalHeaders, body)
}
