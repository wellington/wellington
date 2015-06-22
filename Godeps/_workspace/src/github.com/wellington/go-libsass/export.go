package context

// The use of //export prevents being able to define any C code in the
// preamble of that file.  Export
// defines additional C code necessary for the context<->sass_context bridge.
// See: http://golang.org/cmd/cgo/#hdr-C_references_to_Go

// #cgo CFLAGS: -O2 -fPIC
// #cgo CPPFLAGS: -w
// #cgo CXXFLAGS: -g -std=c++0x -pedantic -Wno-c++11-extensions -O2 -fPIC
// #cgo darwin linux LDFLAGS: -ldl
// #cgo LDFLAGS: -lstdc++ -lm
// #include "sass_context.h"
//
import "C"
import (
	"log"
	"reflect"
	"strings"
	"unsafe"
)

// GoBridge is exported to C for linking libsass to Go.  This function
// adheres to the interface provided by libsass.
//
//export GoBridge
func GoBridge(cargs UnionSassValue, ptr unsafe.Pointer) UnionSassValue {
	// Recover the Cookie struct passed in
	ck := *(*Cookie)(ptr)
	usv := ck.Fn(ck.Ctx, cargs)
	return usv
}

// HeaderBridge is called by libsass to find available custom headers
//
//export HeaderBridge
func HeaderBridge(ptr unsafe.Pointer) C.Sass_Import_List {
	ctx := (*Context)(ptr)
	l := ctx.Headers.Len()
	list := C.sass_make_import_list(C.size_t(l))
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(list)),
		Len:  l, Cap: l,
	}
	golist := *(*[]C.Sass_Import_Entry)(unsafe.Pointer(&hdr))

	for i, head := range ctx.Headers.h {
		ent := C.sass_make_import_entry(
			nil,
			C.CString(head.Content),
			nil)
		cent := (C.Sass_Import_Entry)(ent)
		golist[i] = cent
	}
	return list
}

// ImporterBridge is called by C to pass Importer arguments into Go land. A
// Sass_Import is returned for libsass to resolve.
//
//export ImporterBridge
func ImporterBridge(url *C.char, prev *C.char, ptr unsafe.Pointer) C.Sass_Import_List {
	ctx := (*Context)(ptr)
	parent := C.GoString(prev)
	rel := C.GoString(url)
	list := C.sass_make_import_list(1)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(list)),
		Len:  1, Cap: 1,
	}
	golist := *(*[]C.Sass_Import_Entry)(unsafe.Pointer(&hdr))
	if body, err := ctx.Imports.Get(parent, rel); err == nil {
		conts := C.CString(string(body))
		ent := C.sass_make_import_entry(url, conts, nil)
		cent := (C.Sass_Import_Entry)(ent)
		golist[0] = cent
	} else if strings.HasPrefix(rel, "compass") {
		ent := C.sass_make_import_entry(url, C.CString(""), nil)
		cent := (C.Sass_Import_Entry)(ent)
		golist[0] = cent
	} else {
		ent := C.sass_make_import_entry(url, nil, nil)
		cent := (C.Sass_Import_Entry)(ent)
		golist[0] = cent
	}

	return list
}

// SampleCB example how a callback is defined
func SampleCB(ctx *Context, usv UnionSassValue) UnionSassValue {
	var sv []interface{}
	Unmarshal(usv, &sv)
	return C.sass_make_boolean(false)
}

// Error takes a Go error and returns a libsass Error
func Error(err error) UnionSassValue {
	return C.sass_make_error(C.CString(err.Error()))
}

// Warn takes a string and causes a warning in libsass
func Warn(s string) UnionSassValue {
	return C.sass_make_warning(C.CString(s))
}

// WarnHandler captures Sass warnings and redirects to stdout
func WarnHandler(ctx *Context, csv UnionSassValue) UnionSassValue {
	var s string
	Unmarshal(csv, &s)
	log.Println("WARNING: " + s)

	r, _ := Marshal("")
	return r
}

func init() {
	RegisterHandler("@warn", WarnHandler)
}
