// +build !dev

package libs

// #cgo CFLAGS: -O2 -fPIC
// #cgo CPPFLAGS: -I../libsass-build -I../libsass-build/include
// #cgo CXXFLAGS: -g -std=c++0x -O2 -fPIC
// #cgo LDFLAGS: -lstdc++ -lm
// #cgo darwin linux LDFLAGS: -ldl
//
import "C"
