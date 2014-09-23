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
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	sprite "github.com/drewwells/sprite_sass"
)

var (
	Dir, Input, Output, Includes, Style string
	Comments                            bool
)

func init() {
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
	for _, v := range flag.Args() {
		if strings.HasPrefix(v, "-") {
			log.Fatalf("Please specify flags before other arguments: %s", v)
		}
	}

	if len(flag.Args()) == 0 {
		fmt.Println("Please specify input and output filepaths.")
		fmt.Println("\nAvailable options:")
		flag.PrintDefaults()
		return
	} else if len(flag.Args()) > 1 {
		log.Fatal("Only one input file is supported at this time")
	}

	style, ok := sprite.Style[Style]

	if !ok {
		style = sprite.NESTED_STYLE
	}

	ctx := sprite.Context{
		OutputStyle:  style,
		ImageDir:     Dir,
		Comments:     Comments,
		IncludePaths: []string{filepath.Dir(flag.Arg(0))},
	}

	fRead, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	if Includes != "" {
		ctx.IncludePaths = append(ctx.IncludePaths,
			strings.Split(Includes, ",")...)
	}

	var output io.WriteCloser

	if Output == "" {
		output = os.Stdout
	} else {
		output, err = os.Create(Output)
		if err != nil {
			panic(err)
		}
	}

	err = ctx.Run(fRead, output, filepath.Dir(Input))
	if err != nil {
		log.Fatal(err)
	}
}
