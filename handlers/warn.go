package handlers

import (
	"log"

	"github.com/fatih/color"
	libsass "github.com/wellington/go-libsass"
)

// WarnHandler captures Sass warnings and redirects to stdout
func WarnHandler(v interface{}, csv libsass.SassValue, rsv *libsass.SassValue) error {
	var s string
	libsass.Unmarshal(csv, &s)
	log.Println(color.YellowString("WARNING: " + s))

	r, _ := libsass.Marshal("")
	*rsv = r
	return nil
}

func init() {
	libsass.RegisterHandler("@warn", WarnHandler)
}
