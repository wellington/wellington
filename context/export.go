package context

// The use of //export prevents being able to define any C code in the
// preamble of that file.  Export
// defines additional C code necessary for the context<->sass_context bridge.
// See: http://golang.org/cmd/cgo/#hdr-C_references_to_Go

// #cgo pkg-config: --cflags --libs libsass
// #cgo LDFLAGS: -lsass -lstdc++ -lm -std=c++11
// #include "sass_context.h"
//
import "C"
import (
	"log"
	"unsafe"
)

// Cookie is used for passing context information to libsass.  Cookie is
// passed to custom handlers when libsass executes them through the go bridge.
type Cookie struct {
	Sign string
	Fn   SassCallback
	Ctx  *Context
}

// GoBridge is exported to C for linking libsass to Go.  This function adheres
// to the interface provided by libsass.
//
//export GoBridge
func GoBridge(cargs UnionSassValue, ptr unsafe.Pointer) UnionSassValue {
	// Recover the Cookie struct passed in
	ck := *(*Cookie)(ptr)
	usv := ck.Fn(ck.Ctx, cargs)
	return usv
}

// ImporterBridge is called by C to pass Importer arguments into Go land. A
// Sass_Import is returned for libsass to resolve.
//
//export ImporterBridge
func ImporterBridge(url *C.char, prev *C.char, ptr unsafe.Pointer) **C.struct_Sass_Import {
	ctx := (*Context)(ptr)
	// parent := C.GoString(prev)
	rel := C.GoString(url)
	list := C.sass_make_import_list(1)
	golist := (*[1]*C.struct_Sass_Import)(unsafe.Pointer(list))
	if ref, ok := ctx.FindImport(rel); ok {
		conts := C.CString(ref.Contents)
		srcmap := C.CString("")
		ent := C.sass_make_import_entry(url, conts, srcmap)
		golist[0] = ent
	} else {
		ent := C.sass_make_import_entry(url, nil, nil)
		golist[0] = ent
	}
	return list
}

// SassCallback defines the callback libsass eventually executes in sprite_sass
type SassCallback func(ctx *Context, csv UnionSassValue) UnionSassValue

type handler struct {
	sign     string
	callback func(ctx *Context, csv UnionSassValue) UnionSassValue
}

// handlers is the list of registered sass handlers
var handlers []handler

// SampleCB example how a callback is defined
func SampleCB(ctx *Context, usv UnionSassValue) UnionSassValue {
	var sv []interface{}
	Unmarshal(usv, &sv)
	return C.sass_make_boolean(false)
}

func Error(err error) UnionSassValue {
	return C.sass_make_error(C.CString(err.Error()))
}

// Warn takes a string and causes a warning in libsass
func Warn(s string) UnionSassValue {
	return C.sass_make_error(C.CString("@warn" + s + ";"))
}

// RegisterHandler sets the passed signature and callback to the
// handlers array.
func RegisterHandler(sign string,
	callback func(ctx *Context, csv UnionSassValue) UnionSassValue) {
	handlers = append(handlers, handler{sign, callback})
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
