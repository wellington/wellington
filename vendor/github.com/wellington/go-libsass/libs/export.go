package libs

// #include "sass/context.h"
//
import "C"
import (
	"fmt"
	"sync"
)

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

// gate gobridge, it has some unknown race conditions
var gobridgeMu sync.Mutex

// GoBridge is exported to C for linking libsass to Go.  This function
// adheres to the interface provided by libsass.
//
//export GoBridge
func GoBridge(cargs UnionSassValue, cidx C.int) UnionSassValue {
	// Recover the Cookie struct passed in
	idx := int(cidx)
	ck, ok := globalFuncs.Get(idx).(Cookie)
	if !ok {
		fmt.Printf("failed to resolve Cookie %d\n", idx)
		return MakeNil()
	}
	// ck := *(*Cookie)(ptr)

	var usv UnionSassValue
	err := ck.Fn(ck.Ctx, cargs, &usv)
	_ = err
	return usv
}
