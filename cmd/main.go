package main

import (
	"flag"
	"fmt"

	sprite "github.com/drewwells/sprite-sass"
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
