// +build dev

package libs

// #cgo CPPFLAGS: -DUSE_LIBSASS
// #cgo CPPFLAGS: -I../libsass-build -I../libsass-build/include
// #cgo LDFLAGS: -lsass
// #cgo darwin linux LDFLAGS: -ldl
//
import "C"
