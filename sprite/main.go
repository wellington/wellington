// Main package wraps sprite_sass tool for use with the command line
// See -h for list of available options
package main

/*
#cgo LDFLAGS: -Llibsass -lsass -lstdc++ -lm
#cgo CFLAGS: -Ilibsass

#include <stdlib.h>
#include <sass_context.h>
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

const version = `v0.1.0`

var (
	Dir, Gen, Input, Output, Includes string
	MainFile, Style                   string
	Comments                          bool
	cpuprofile                        string
	ShowVersion                       bool
	BuildDir                          string
)

func init() {
	flag.StringVar(&Output, "output", "", "Output file")
	flag.StringVar(&Output, "o", "", "Output file")
	flag.StringVar(&BuildDir, "b", "", "Build Directory")
	flag.StringVar(&Includes, "p", "", "SASS import path")
	flag.StringVar(&Dir, "dir", "", "Image directory")
	flag.StringVar(&Dir, "d", "", "Image directory")
	flag.StringVar(&Gen, "gen", ".", "Directory for generated images")
	flag.StringVar(&Style, "style", "nested", "CSS nested style")
	flag.StringVar(&Style, "s", "nested", "CSS nested style")
	flag.BoolVar(&Comments, "comment", true, "Turn on source comments")
	flag.BoolVar(&Comments, "c", true, "Turn on source comments")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.BoolVar(&ShowVersion, "version", false, "Show the app version")
}

func main() {
	flag.Parse()

	if ShowVersion {
		fmt.Println(version)
		os.Exit(0)
	}

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
		fmt.Println("Please specify input filepath.")
		fmt.Println("\nAvailable options:")
		flag.PrintDefaults()
		return
	}

	style, ok := sprite.Style[Style]

	if !ok {
		style = sprite.NESTED_STYLE
	}
	for _, f := range flag.Args() {
		// Remove partials
		if strings.HasPrefix(filepath.Base(f), "_") {
			continue
		}
		log.Println("Open:", f)

		// If no imagedir specified, assume relative to the input file
		if Dir == "" {
			Dir = filepath.Dir(f)
		}
		rel, _ := filepath.Rel(Includes, filepath.Dir(f))
		filename := strings.Replace(filepath.Base(f), ".scss", ".css", 1)
		output := filepath.Join(BuildDir, rel, filename)
		ctx := sprite.Context{
			OutputStyle: style,
			ImageDir:    Dir,
			// Assumption that output is a file
			BuildDir:     filepath.Dir(output),
			GenImgDir:    Gen,
			MainFile:     f,
			Comments:     Comments,
			IncludePaths: []string{filepath.Dir(f)},
		}
		if Includes != "" {
			ctx.IncludePaths = append(ctx.IncludePaths,
				strings.Split(Includes, ",")...)
		}
		fRead, err := os.Open(f)
		if err != nil {
			panic(err)
		}

		var out io.WriteCloser

		if output == "" {
			out = os.Stdout
		} else {
			dir := filepath.Dir(output)
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				log.Fatalf("Failed to create directory: %s", dir)
			}

			out, err = os.Create(output)
			log.Println("Created:", output)
			if err != nil {
				log.Fatalf("Failed to create file: %s", Output)
			}
		}

		err = ctx.Run(fRead, out, filepath.Dir(Input))
		if err != nil {
			log.Println(err)
		}
	}
}
