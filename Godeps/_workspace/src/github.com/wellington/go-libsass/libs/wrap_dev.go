// +build dev

package libs

// #cgo pkg-config: --cflags --libs libsass
// #cgo LDFLAGS: -lsass -lstdc++ -ldl -lm
//
import "C"
