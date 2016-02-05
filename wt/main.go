// Main package wraps sprite_sass tool for use with the command line
// See -h for list of available options
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/fatih/color"
	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/spritewell"
	"github.com/wellington/wellington/payload"
	"github.com/wellington/wellington/version"

	wt "github.com/wellington/wellington"
	_ "github.com/wellington/wellington/handlers"
)

var (
	proj                 string
	includes             []string
	font, dir, gen       string
	style                string
	comments             bool
	cpuprofile, buildDir string
	httpPath             string
	timeB                bool
	config               string
	debug                bool
	cachebust            string

	// unused
	relativeAssets bool

	paths []string
)

/*
   --app APP                    Tell compass what kind of application it is integrating with. E.g. rails
   --fonts-dir FONTS_DIR        The directory where you keep your fonts.
*/

func flags(app *kingpin.Application) {
	// Unused cli args
	app.Flag("build", "Path to target directory to place generated CSS, relative paths inside project directory are preserved").
		Short('b').StringVar(&buildDir)
	app.Flag("cachebust", "Defeat cache by appending timestamps to static assets ie. ts, sum, timestamp").StringVar(&cachebust)
	app.Flag("comment", "Turn on source comments").BoolVar(&comments)
	app.Flag("config", "Temporarily disabled: Location of the config file").Short('c').ExistingFileVar(&config)
	app.Flag("cpuprofile", "Go runtime cpu profilling for debugging").StringVar(&cpuprofile)
	app.Flag("css-dir", "Deprecated: Use -b instead").Hidden().ExistingDirVar(&buildDir)
	app.Flag("debug", "Show detailed debug information").Hidden().BoolVar(&debug)
	app.Flag("debug-info", "Deprecated: Use --debug instead").Hidden().BoolVar(&debug)

	app.Flag("dir", "Path to locate images for spriting and image functions").Short('d').ExistingDirVar(&dir)
	app.Flag("font", "Path to directory containing fonts").Default(".").ExistingDirVar(&font)
	app.Flag("gen", "Path to place generated images").Default(".").StringVar(&gen)
	app.Flag("generated-images-path", "Deprecated: Use --gen instead").Hidden().Default(".").ExistingDirVar(&gen)
	app.Flag("images-dir", "Deprecated: Use -d instead").Hidden().ExistingDirVar(&dir)
	app.Flag("includes", "Include Sass from additional directories").Short('I').ExistingDirsVar(&includes)
	app.Flag("output-style", "Deprecated: Use --style instead").Hidden().Short('s').Default("nested").EnumVar(&style, "nested", "expanded", "compact", "compressed")
	app.Flag("proj", "Path to directory containing Sass stylesheets").Short('p').ExistingDirVar(&proj)
	app.Flag("relative-assets", "UNSUPPORTED: Make compass asset helpers generate relative urls to assets.").BoolVar(&relativeAssets)
	app.Flag("sass-dir", "Deprecated: Use --includes instead").Hidden().ExistingDirsVar(&includes)
	app.Flag("style", "Nested style of output CSS. Available options: nested, expanded, compact, compressed").Short('s').Default("nested").EnumVar(&style, "nested", "expanded", "compact", "compressed")
	app.Flag("time", "Retrieve timing information").BoolVar(&timeB)
}

func hostname() string {
	if host := os.Getenv("HOSTNAME"); len(host) > 0 {
		if !strings.HasPrefix(host, "http") {
			return "http://" + host
		}
		return host
	}

	if host, err := os.Hostname(); err == nil {
		return "http://" + host
	}

	return ""
}

func main() {
	app := kingpin.New("wt", "wt is a Sass project tool made to handle large projects. It uses the libSass compiler for efficiency and speed.").
		Version(fmt.Sprintf("   libsass: %s\nWellington: %s", libsass.Version(), version.Version))
	flags(app)

	serveCmd := app.Command("serve", "Starts a http server that will convert Sass to CSS")
	serveCmd.Flag("httppath", "Only for HTTP, overrides generated sprite paths to support http").Default(hostname()).StringVar(&httpPath)

	compileCmd := app.Command("compile", "Fast compilation of Sass stylesheets to CSS. For usage consult the documentation at https://github.com/wellington/wellington#wellington").Alias("")
	watchCmd := app.Command("watch", "Watch Sass files for changes and rebuild CSS")

	for _, cmd := range []*kingpin.CmdClause{serveCmd, compileCmd, watchCmd} {
		cmd.Arg("paths", "Target file or directories").ExistingFilesOrDirsVar(&paths)
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case serveCmd.FullCommand():
		Serve()
	case compileCmd.FullCommand():
		Compile()
	case watchCmd.FullCommand():
		Watch()
	}
}

func makeabs(wd string, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(wd, path)
}

func parseBuildArgs() *wt.BuildArgs {
	style, ok := libsass.Style[style]

	if !ok {
		style = libsass.NESTED_STYLE
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("could not find working directory", err)
	}

	proj = makeabs(wd, proj)

	incs := make([]string, len(includes))
	for i := range includes {
		incs[i] = makeabs(wd, includes[i])
	}

	dir = makeabs(wd, dir)
	font = makeabs(wd, font)
	if len(buildDir) > 0 {
		buildDir = makeabs(wd, buildDir)
		// If buildDir specified, make relative to that
		gen = makeabs(buildDir, gen)
	} else {
		gen = makeabs(wd, gen)
	}
	incs = append(incs, paths...)

	gba := &wt.BuildArgs{
		ImageDir:  dir,
		BuildDir:  buildDir,
		Includes:  append([]string{proj}, incs...),
		Font:      font,
		Style:     style,
		Gen:       gen,
		Comments:  comments,
		CacheBust: cachebust,
	}
	gba.WithPaths(paths)
	return gba
}

func globalRun() (*wt.SafePartialMap, *wt.BuildArgs) {
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

	if gen != "" {
		err := os.MkdirAll(gen, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	pMap := wt.NewPartialMap()
	gba := parseBuildArgs()
	if debug {
		log.Printf("      Font  Dir: %s\n", gba.Font)
		log.Printf("      Image Dir: %s\n", gba.ImageDir)
		log.Printf("      Build Dir: %s\n", gba.BuildDir)
		log.Printf("Build Image Dir: %s\n", gba.Gen)
		log.Printf(" Include Dir(s): %s\n", gba.Includes)
		log.Println("===================================")
	}
	return pMap, gba

}

// Watch accepts a set of paths starting a recursive file watcher
func Watch() {
	pMap, gba := globalRun()
	var err error
	bOpts := wt.NewBuild(paths, gba, pMap)
	err = bOpts.Run()
	if err != nil {
		log.Fatal(err)
	}
	w, err := wt.NewWatcher(&wt.WatchOptions{
		Paths:      paths,
		BArgs:      gba,
		PartialMap: pMap,
	})
	if err != nil {
		log.Fatal("failed to start watcher: ", err)
	}
	err = w.Watch()
	if err != nil {
		log.Fatal("filewatcher error: ", err)
	}

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

// lis is exposed so test suite can shut it down
var lis net.Listener

// Serve starts a web server accepting POST calls and return CSS
func Serve() {
	_, gba := globalRun()
	if len(gba.Gen) == 0 {
		log.Fatal("Must pass an image build directory to use HTTP")
	}

	http.Handle("/build/", wt.FileHandler(gba.Gen))
	log.Println("Web server started on :12345")

	var err error
	lis, err = net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatal("Error listening on :12345", err)
	}

	http.HandleFunc("/", wt.HTTPHandler(gba, httpPath))
	http.Serve(lis, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	log.Println("Server closed")
}

// Compile handles compile files and stdin operations.
func Compile() {
	start := time.Now()
	pMap, gba := globalRun()
	if gba == nil {
		return
	}

	defer func() {
		log.Printf("Compilation took: %s\n", time.Since(start))
	}()

	run(pMap, gba)
}

// Run is the main entrypoint for the cli.
func run(pMap *wt.SafePartialMap, gba *wt.BuildArgs) {

	// No paths given, read from stdin and wait
	if len(paths) == 0 {
		log.Println("Reading from stdin, -h for help")
		out := os.Stdout
		in := os.Stdin
		comp, err := wt.FromBuildArgs(out, in, gba)
		if err != nil {
			log.Fatal(err)
		}
		err = comp.Run()
		if err != nil {
			color.Red(err.Error())
		}
		return
	}

	bOpts := wt.NewBuild(paths, gba, pMap)

	err := bOpts.Run()
	if err != nil {
		log.Fatal(err)
	}

	// FIXME: move this to a Payload.Close() method

	// Before shutting down, check that every sprite has been
	// flushed to disk.
	img := sync.WaitGroup{}
	pMap.RLock()
	// It's not currently possible to wait on Image. This is often
	// to inline images, so it shouldn't be a factor...
	// for _, s := range gba.Payload.Image().M {
	// 	img.Add(1)
	// 	err := s.Wait()
	// 	img.Done()
	// 	if err != nil {
	// 		log.Printf("error writing image: %s\n", err)
	// 	}
	// }
	sprites := payload.Sprite(gba.Payload)
	sprites.ForEach(func(k string, sprite *spritewell.Sprite) {
		img.Add(1)
		err := sprite.Wait()
		img.Done()
		if err != nil {
			log.Printf("error writing sprite: %s\n", err)
		}
	})
	img.Wait()
	pMap.RUnlock()

}
