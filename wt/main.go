// Main package wraps sprite_sass tool for use with the command line
// See -h for list of available options
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/wellington/wellington/context"

	wt "github.com/wellington/wellington"
	_ "github.com/wellington/wellington/handlers"
)

const version = `v0.6.0`

var (
	font, dir, gen, includes  string
	mainFile, style           string
	comments, watch           bool
	cpuprofile, buildDir      string
	ishttp, help, showVersion bool
)

func init() {
	flag.BoolVar(&showVersion, "version", false, "Show the app version")

	flag.BoolVar(&help, "help", false, "this help")
	flag.BoolVar(&help, "h", false, "this help")

	flag.StringVar(&buildDir, "b", "", "Build Directory")
	flag.StringVar(&gen, "gen", ".", "Directory for generated images")

	flag.StringVar(&includes, "p", "", "SASS import path")
	flag.StringVar(&dir, "dir", "", "Image directory")
	flag.StringVar(&dir, "d", "", "Image directory")
	flag.StringVar(&font, "font", ".", "Font Directory")

	flag.StringVar(&style, "style", "nested", "CSS nested style")
	flag.StringVar(&style, "s", "nested", "CSS nested style")
	flag.BoolVar(&comments, "comment", true, "Turn on source comments")
	flag.BoolVar(&comments, "c", true, "Turn on source comments")

	flag.BoolVar(&ishttp, "http", false, "Listen for http connections")
	flag.BoolVar(&watch, "watch", false, "File watcher that will rebuild css on file changes")
	flag.BoolVar(&watch, "w", false, "File watcher that will rebuild css on file changes")

	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
}

func main() {
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// Profiling code
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Starting profiler")
		pprof.StartCPUProfile(f)
		defer func() {
			pprof.StopCPUProfile()
			err := f.Close()
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Stopping Profiller")
		}()
	}

	for _, v := range flag.Args() {
		if strings.HasPrefix(v, "-") {
			log.Fatalf("Please specify flags before other arguments: %s", v)
		}
	}

	if help {
		fmt.Println("Please specify input filepath.")
		fmt.Println("\nAvailable options:")
		flag.PrintDefaults()
		return
	}

	if gen != "" {
		err := os.MkdirAll(gen, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	style, ok := context.Style[style]

	if !ok {
		style = context.NESTED_STYLE
	}

	gba := wt.NewBuildArgs()

	gba.Dir = dir
	gba.BuildDir = buildDir
	gba.Includes = includes
	gba.Font = font
	gba.Style = style
	gba.Gen = gen
	gba.Comments = comments

	pMap := wt.NewPartialMap()
	// FIXME: Copy pasta with LoadAndBuild
	ctx := &context.Context{
		Sprites:      gba.Sprites,
		Imgs:         gba.Imgs,
		OutputStyle:  gba.Style,
		ImageDir:     gba.Dir,
		FontDir:      gba.Font,
		GenImgDir:    gba.Gen,
		Comments:     gba.Comments,
		IncludePaths: []string{gba.Includes},
	}

	if ishttp {
		http.Handle(Dir,
			http.StripPrefix("/build",
				http.FileServer(http.Dir(Dir)),
			),
		)

		http.HandleFunc("/", httpHandler(ctx))
		err := http.ListenAndServe(":12345", nil)

		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}

		return
	}

	if len(flag.Args()) == 0 {

		// Read from stdin
		fmt.Println("Reading from stdin, -h for help")
		out := os.Stdout
		in := os.Stdin

		var pout bytes.Buffer
		_, err := wt.StartParser(ctx, in, &pout, wt.NewPartialMap())
		if err != nil {
			log.Println(err)
		}
		err = ctx.Compile(&pout, out)

		if err != nil {
			log.Println(err)
		}
	}

	sassPaths := make([]string, len(flag.Args()))
	for i, f := range flag.Args() {
		sassPaths[i] = filepath.Dir(f)
		err := wt.LoadAndBuild(f, gba, pMap)
		if err != nil {
			log.Println(err)
		}
	}

	if watch {
		w := wt.NewWatcher()
		w.PartialMap = pMap
		w.Dirs = sassPaths
		w.BArgs = gba
		w.Watch()

		fmt.Println("File watcher started use `ctrl+d` to exit")
		in := bufio.NewReader(os.Stdin)
		for {
			_, err := in.ReadString(' ')
			if err != nil {
				if err == io.EOF {
					os.Exit(0)
				}
				fmt.Println("error", err)
			}
		}
	}
}

func httpHandler(ctx *context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var pout bytes.Buffer

		// Set headers
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		_, err := wt.StartParser(ctx, r.Body, &pout, wt.NewPartialMap())
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}

		err = ctx.Compile(&pout, w)
		if err != nil {
			io.WriteString(w, err.Error())
		}
	}
}
