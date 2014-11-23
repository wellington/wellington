// Main package wraps sprite_sass tool for use with the command line
// See -h for list of available options
package main

/*
#cgo LDFLAGS: -L../libsass/lib -lsass -lstdc++ -lm
#cgo CFLAGS: -I../libsass

#include <stdlib.h>
#include <sass_context.h>
*/
import "C"

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/drewwells/sprite_sass/context"

	sprite "github.com/drewwells/sprite_sass"
)

const version = `v0.2.0`

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

	style, ok := context.Style[Style]

	if !ok {
		style = context.NESTED_STYLE
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
		ctx := context.Context{
			// TODO: Most of these fields are no longer used
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
		var pout bytes.Buffer
		startParser(ctx, fRead, &pout, filepath.Dir(Input))
		err = ctx.Compile(&pout, out, filepath.Dir(Input))
		// err = ctx.Run(fRead, out, filepath.Dir(Input))
		if err != nil {
			log.Println(err)
		}
	}
}

func startParser(ctx context.Context, in io.Reader, out io.Writer, pkgdir string) error {
	// Run the sprite_sass parser prior to passing to libsass
	parser := sprite.Parser{
		ImageDir:  ctx.ImageDir,
		Includes:  ctx.IncludePaths,
		BuildDir:  ctx.BuildDir,
		GenImgDir: ctx.GenImgDir,
	}
	bs, err := parser.Start(in, pkgdir)
	if err != nil {
		return err
	}
	out.Write(bs)
	return err
}
