package main

/*
#cgo LDFLAGS: -L../libsass -lsass -lstdc++
#cgo CFLAGS: -I../libsass

#include <stdlib.h>
#include <sass_interface.h>
*/
import "C"

import (
	"flag"
	"fmt"

	sprite "github.com/drewwells/sprite_sass"
)

var (
	Input, Output string
)

func init() {
	flag.StringVar(&Input, "i", "nodefault", "Input file")
	flag.StringVar(&Output, "o", "nodef", "Output file")
}

func main() {
	flag.Parse()

	ctx := sprite.Context{}

	ctx.Input = Input
	ctx.Output = Output
	fmt.Printf("%+v\n", ctx)
}
