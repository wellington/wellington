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
	"runtime/pprof"
	"strings"

	sprite "github.com/drewwells/sprite_sass"
)

var (
	Dir, Gen, Input, Output, Includes, Style string
	Comments                                 bool
	cpuprofile                               string
)

func init() {
	flag.StringVar(&Output, "output", "", "Output file")
	flag.StringVar(&Output, "o", "", "Output file")
	flag.StringVar(&Includes, "p", "", "SASS import path")
	flag.StringVar(&Dir, "dir", ".", "Image directory")
	flag.StringVar(&Dir, "d", ".", "Image directory")
	flag.StringVar(&Gen, "gen", ".", "Directory for generated images")
	flag.StringVar(&Style, "style", "nested", "CSS nested style")
	flag.StringVar(&Style, "s", "nested", "CSS nested style")
	flag.BoolVar(&Comments, "comment", true, "Turn on source comments")
	flag.BoolVar(&Comments, "c", true, "Turn on source comments")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")

}

func main() {
	flag.Parse()

	// Profiling code
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer func() {
			err := f.Close()
			if err != nil {
				log.Fatal(err)
			}
			pprof.StopCPUProfile()
		}()
	}

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
		OutputStyle: style,
		ImageDir:    Dir,
		// Assumption that output is a file
		BuildDir:     filepath.Dir(Output),
		GenImgDir:    Gen,
		Comments:     Comments,
		IncludePaths: []string{filepath.Dir(flag.Arg(0))},
	}

	if Includes != "" {
		ctx.IncludePaths = append(ctx.IncludePaths,
			strings.Split(Includes, ",")...)
	}

	fRead, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	var output io.WriteCloser

	if Output == "" {
		output = os.Stdout
	} else {
		dir := filepath.Dir(Output)
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			log.Fatalf("Failed to create directory: %s", dir)
		}

		output, err = os.Create(Output)
		if err != nil {
			log.Fatalf("Failed to create file: %s", Output)
		}
	}

	err = ctx.Run(fRead, output, filepath.Dir(Input))
	if err != nil {
		log.Fatal(err)
	}
}
