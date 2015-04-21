package context

// #include "sass_context.h"
//
import "C"

// Version reports libsass version information
func Version() string {
	ver := C.GoString(C.libsass_version())
	return ver
}
