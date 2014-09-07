// Main package wraps sprite_sass tool for use with the command line
// See -h for list of available options
package main

/*
#cgo LDFLAGS: -Llibsass -lsass -lstdc++
#cgo CFLAGS: -Ilibsass

#include <stdlib.h>
#include <sass_interface.h>
*/
import "C"

import (
	"flag"
	"fmt"
	"strings"

	sprite "github.com/drewwells/sprite_sass"
)

var (
	Dir, Input, Output, Includes, Style string
	Comments                            bool
)

func init() {
	flag.StringVar(&Input, "input", "", "Input file")
	flag.StringVar(&Input, "i", "", "Input file")
	flag.StringVar(&Output, "output", "", "Output file")
	flag.StringVar(&Output, "o", "", "Output file")
	flag.StringVar(&Includes, "p", "", "SASS import path")
	flag.StringVar(&Dir, "dir", ".", "Image directory")
	flag.StringVar(&Dir, "d", ".", "Image directory")
	flag.StringVar(&Style, "style", "nested", "CSS nested style")
	flag.StringVar(&Style, "s", "nested", "CSS nested style")
	flag.BoolVar(&Comments, "comment", true, "Turn on source comments")
	flag.BoolVar(&Comments, "c", true, "Turn on source comments")
}

func main() {
	flag.Parse()

	if Input == "" {
		fmt.Println("Please specify input and output filepaths.")
		fmt.Println("\nAvailable options:")
		flag.PrintDefaults()
		return
	}

	style, ok := sprite.Style[Style]

	if !ok {
		style = sprite.NESTED_STYLE
	}

	ctx := sprite.Context{
		OutputStyle: style,
		ImageDir:    Dir,
		Comments:    Comments,
	}

	if Includes != "" {
		ctx.IncludePaths = strings.Split(Includes, ",")
	}

	err := ctx.Run(Input, Output)
	if err != nil {
		panic(err)
	}
}
