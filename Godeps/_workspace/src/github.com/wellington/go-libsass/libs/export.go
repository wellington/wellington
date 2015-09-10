package libs

// #include "sass/context.h"
//
import "C"
import "unsafe"

// SassCallback defines the callback libsass eventually executes in
// sprite_sass
type SassCallback func(v interface{}, csv UnionSassValue, rsv *UnionSassValue) error

// Cookie is used for passing context information to libsass.  Cookie is
// passed to custom handlers when libsass executes them through the go
// bridge.
type Cookie struct {
	Sign string
	Fn   SassCallback
	Ctx  interface{}
}

// GoBridge is exported to C for linking libsass to Go.  This function
// adheres to the interface provided by libsass.
//
//export GoBridge
func GoBridge(cargs UnionSassValue, ptr unsafe.Pointer) UnionSassValue {
	// Recover the Cookie struct passed in
	ck := *(*Cookie)(ptr)
	var usv UnionSassValue
	err := ck.Fn(ck.Ctx, cargs, &usv)
	_ = err
	return usv
}
