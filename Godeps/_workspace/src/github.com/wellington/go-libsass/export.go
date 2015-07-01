package context

import (
	"log"

	"github.com/wellington/go-libsass/libs"
)

// Error takes a Go error and returns a libsass Error
func Error(err error) SassValue {
	return SassValue{value: libs.MakeError(err.Error())}
}

// Warn takes a string and causes a warning in libsass
func Warn(s string) SassValue {
	return SassValue{value: libs.MakeWarning(s)}
}

// WarnHandler captures Sass warnings and redirects to stdout
func WarnHandler(v interface{}, csv SassValue, rsv *SassValue) error {
	var s string
	Unmarshal(csv, &s)
	log.Println("WARNING: " + s)

	r, _ := Marshal("")
	*rsv = r
	return nil
}

func init() {
	RegisterHandler("@warn", WarnHandler)
}
