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
	"time"

	"github.com/wellington/wellington/context"

	wt "github.com/wellington/wellington"
	_ "github.com/wellington/wellington/handlers"
)

const version = `v0.7.0`

var (
	font, dir, gen, includes  string
	mainFile, style           string
	comments, watch           bool
	cpuprofile, buildDir      string
	jsDir                     string
	ishttp, help, showVersion bool
	httpPath                  string
	timeB                     bool
)

/*
   -c, --config CONFIG_FILE         Specify the location of the configuration file explicitly.
       --app APP                    Tell compass what kind of application it is integrating with. E.g. rails
       --fonts-dir FONTS_DIR        The directory where you keep your fonts.
*/
func init() {
	flag.BoolVar(&showVersion, "version", false, "Show the app version")
	flag.BoolVar(&showVersion, "v", false, "Show the app version")

	flag.BoolVar(&help, "help", false, "this help")
	flag.BoolVar(&help, "h", false, "this help")

	// Interoperability args
	flag.StringVar(&gen, "css-dir", "", "Compass Build Directory")
	flag.StringVar(&dir, "images-dir", "", "Compass Image Directory")
	flag.StringVar(&includes, "sass-dir", "", "Compass Sass Directory")
	flag.StringVar(&jsDir, "javascripts-dir", "", "Compass JS Directory")
	flag.BoolVar(&timeB, "time", false, "Retrieve timing information")

	flag.StringVar(&buildDir, "b", "", "Build Directory")
	flag.StringVar(&gen, "gen", ".", "Generated images directory")

	flag.StringVar(&includes, "proj", "", "Project directory")
	flag.StringVar(&includes, "p", "", "Project directory")
	flag.StringVar(&dir, "dir", "", "Image directory")
	flag.StringVar(&dir, "d", "", "Image directory")
	flag.StringVar(&font, "font", ".", "Font Directory")

	flag.StringVar(&style, "style", "nested", "CSS nested style")
	flag.StringVar(&style, "s", "nested", "CSS nested style")
	flag.BoolVar(&comments, "comment", true, "Turn on source comments")
	flag.BoolVar(&comments, "c", true, "Turn on source comments")

	flag.BoolVar(&ishttp, "http", false, "Listen for http connections")
	flag.StringVar(&httpPath, "httppath", "",
		"Only for HTTP, overrides generated sprite paths to support http")
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

	var start time.Time
	if timeB {
		start = time.Now()
	}
	defer func() {
		diff := time.Since(start)
		log.Printf("Compilation took: %v\n", diff)
	}()

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
		BuildDir:     gba.BuildDir,
		ImageDir:     gba.Dir,
		FontDir:      gba.Font,
		GenImgDir:    gba.Gen,
		Comments:     gba.Comments,
		HTTPPath:     httpPath,
		IncludePaths: []string{gba.Includes},
	}

	if ishttp {
		if len(gba.Gen) == 0 {
			log.Fatal("Must pass an image build directory to use HTTP")
		}
		http.Handle("/build/", wt.FileHandler(gba.Gen))
		log.Println("Web server started on :12345")
		http.HandleFunc("/", wt.HTTPHandler(ctx))
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
