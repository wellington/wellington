// +build dev

package libs

// #cgo pkg-config: libsass
// #cgo CPPFLAGS: -DUSE_LIBSASS
// #cgo LDFLAGS: -lsass
//
import "C"
