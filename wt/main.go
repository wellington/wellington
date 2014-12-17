// Main package wraps sprite_sass tool for use with the command line
// See -h for list of available options
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/wellington/spritewell"
	"github.com/wellington/wellington/context"

	wt "github.com/wellington/wellington"
	_ "github.com/wellington/wellington/context/handlers"
)

const version = `v0.4.0`

var (
	Font, Dir, Gen, Includes string
	MainFile, Style          string
	Comments, Watch          bool
	cpuprofile               string
	Help, ShowVersion        bool
	BuildDir                 string
)

func init() {
	flag.BoolVar(&ShowVersion, "version", false, "Show the app version")

	flag.BoolVar(&Help, "help", false, "this help")
	flag.BoolVar(&Help, "h", false, "this help")

	flag.StringVar(&BuildDir, "b", "", "Build Directory")
	flag.StringVar(&Gen, "gen", ".", "Directory for generated images")

	flag.StringVar(&Includes, "p", "", "SASS import path")
	flag.StringVar(&Dir, "dir", "", "Image directory")
	flag.StringVar(&Dir, "d", "", "Image directory")
	flag.StringVar(&Font, "font", ".", "Font Directory")

	flag.StringVar(&Style, "style", "nested", "CSS nested style")
	flag.StringVar(&Style, "s", "nested", "CSS nested style")
	flag.BoolVar(&Comments, "comment", true, "Turn on source comments")
	flag.BoolVar(&Comments, "c", true, "Turn on source comments")

	flag.BoolVar(&Watch, "watch", false, "File watcher that will rebuild css on file changes")
	flag.BoolVar(&Watch, "w", false, "File watcher that will rebuild css on file changes")

	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
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

	if Help {
		fmt.Println("Please specify input filepath.")
		fmt.Println("\nAvailable options:")
		flag.PrintDefaults()
		return
	}

	if Gen != "" {
		err := os.MkdirAll(Gen, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	style, ok := context.Style[Style]

	if !ok {
		style = context.NESTED_STYLE
	}

	partialMap := wt.NewPartialMap()

	if len(flag.Args()) == 0 {
		// Read from stdin
		log.Print("Reading from stdin, -h for help")
		out := os.Stdout
		in := os.Stdin

		var pout bytes.Buffer
		ctx := context.Context{}

		_, err := wt.StartParser(&ctx, in, &pout, "", partialMap)
		if err != nil {
			log.Println(err)
		}
		err = ctx.Compile(&pout, out)

		if err != nil {
			log.Println(err)
		}
	}

	SpriteCache := spritewell.SafeImageMap{
		M: make(map[string]spritewell.ImageList, 100)}
	ImageCache := spritewell.SafeImageMap{
		M: make(map[string]spritewell.ImageList, 100)}
	topLevelFilePaths := make([]string, len(flag.Args()))

	globalBuildArgs := wt.BuildArgs{
		Imgs:     ImageCache,
		Sprites:  SpriteCache,
		Dir:      Dir,
		BuildDir: BuildDir,
		Includes: Includes,
		Font:     Font,
		Style:    style,
		Gen:      Gen,
		Comments: Comments,
	}

	for i, f := range flag.Args() {
		topLevelFilePaths[i] = filepath.Dir(f)
		wt.LoadAndBuild(f, &globalBuildArgs, partialMap)
	}

	if Watch {
		//fmt.Println(PartialMap.M["/Users/dslininger/Projects/RetailMeNot/www/gui/sass/bourbon/css3/_hyphens.scss"])
		wt.FileWatch(partialMap, &globalBuildArgs, topLevelFilePaths)
	}
}
