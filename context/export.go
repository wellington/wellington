package context

// The use of //export prevents being able to define any C code in the preamble of that file.  Export
// defines additional C code necessary for the context<->sass_context bridge.
// See: http://golang.org/cmd/cgo/#hdr-C_references_to_Go

// #cgo LDFLAGS: -lsass -lstdc++ -lm
// #cgo CFLAGS:
//
// #include "sass_context.h"
import "C"
import "unsafe"

// goBridge is exported to C for providing a Go function reference that can
// be executed by C.
//export goBridge
func goBridge(cargs UnionSassValue, ptr unsafe.Pointer) UnionSassValue {
	// Recover the Cookie struct passed in
	ck := *(*Cookie)(ptr)
	usv := ck.fn(cargs)
	return usv
}

// CookieCb defines the callback libsass eventually executes in sprite_sass
type SassCallback func(csv UnionSassValue) UnionSassValue

func SampleCB(usv UnionSassValue) UnionSassValue {
	var sv []SassValue
	Unmarshal(usv, &sv)
	return C.sass_make_boolean(false)
}
