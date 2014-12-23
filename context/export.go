package context

// The use of //export prevents being able to define any C code in the preamble of that file.  Export
// defines additional C code necessary for the context<->sass_context bridge.
// See: http://golang.org/cmd/cgo/#hdr-C_references_to_Go

// #cgo pkg-config: --cflags libsass
// #cgo LDFLAGS: -lsass -lstdc++ -lm
// #include "sass_context.h"
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

// CookieCb defines the callback libsass eventually executes in sprite_sass
type SassCallback func(ctx *Context, csv UnionSassValue) UnionSassValue

type handler struct {
	sign     string
	callback func(ctx *Context, csv UnionSassValue) UnionSassValue
}

// handlers is the list of registered sass handlers
var handlers []handler

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
