// +build !dev

package libs

// #cgo CFLAGS: -O2 -fPIC
// #cgo CPPFLAGS: -w
// #cgo CXXFLAGS: -g -std=c++0x -pedantic -Wno-c++11-extensions -O2 -fPIC
// #cgo LDFLAGS: -lstdc++ -lm
//
import "C"
