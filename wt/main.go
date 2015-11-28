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

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/wellington/version"

	wt "github.com/wellington/wellington"
	_ "github.com/wellington/wellington/handlers"
)

var (
	proj                          string
	includes                      []string
	font, dir, gen                string
	mainFile, style               string
	comments, watch               bool
	cpuprofile, buildDir          string
	jsDir                         string
	ishttp, showHelp, showVersion bool
	httpPath                      string
	timeB                         bool
	config                        string
	debug                         bool
	cachebust                     string

	// unused
	relativeAssets bool
	cssDir         string
)

/*
   --app APP                    Tell compass what kind of application it is integrating with. E.g. rails
   --fonts-dir FONTS_DIR        The directory where you keep your fonts.
*/
func init() {

	// Interoperability args
}

func flags(set *pflag.FlagSet) {
	// Unused cli args
	set.StringVarP(&buildDir, "build", "b", "",
		"Path to target directory to place generated CSS, relative paths inside project directory are preserved")
	set.BoolVarP(&comments, "comment", "", false, "Turn on source comments")
	set.BoolVar(&debug, "debug", false, "Show detailed debug information")

	var nothingb bool
	set.BoolVar(&debug, "debug-info", false, "")
	set.MarkDeprecated("debug-info", "Use --debug instead")

	set.StringVarP(&dir, "dir", "d", "",
		"Path to locate images for spriting and image functions")
	set.StringVar(&dir, "images-dir", "", "")
	set.MarkDeprecated("images-dir", "Use -d instead")

	set.StringVar(&font, "font", ".", "Path to directory containing fonts")
	set.StringVar(&gen, "gen", ".", "Path to place generated images")

	set.StringVarP(&proj, "proj", "p", "",
		"Path to directory containing Sass stylesheets")
	set.BoolVar(&nothingb, "no-line-comments", false, "UNSUPPORTED: Disable line comments, use comments")
	set.MarkDeprecated("no-line-comments", "Use --comments instead")
	set.BoolVar(&relativeAssets, "relative-assets", false, "UNSUPPORTED: Make compass asset helpers generate relative urls to assets.")

	set.BoolVarP(&showVersion, "version", "v", false, "Show the app version")
	set.StringVar(&cachebust, "cachebust", "", "Defeat cache by appending timestamps to static assets ie. ts, sum, timestamp")
	set.StringVarP(&style, "style", "s", "nested",
		`nested style of output CSS
                        available options: nested, expanded, compact, compressed`)
	set.StringVar(&style, "output-style", "nested", "")
	set.MarkDeprecated("output-style", "Use --style instead")
	set.BoolVar(&timeB, "time", false, "Retrieve timing information")

	var nothing string
	set.StringVar(&nothing, "require", "", "")
	set.MarkDeprecated("require", "Compass backwards compat, Not supported")
	set.MarkDeprecated("require", "Not supported")
	set.StringVar(&nothing, "environment", "", "")
	set.MarkDeprecated("environment", "Not supported")
	set.StringSliceVar(&includes, "includes", nil, "Include Sass from additional directories")
	set.StringSliceVarP(&includes, "", "I", nil, "")
	set.MarkDeprecated("I", "Compass backwards compat, use --includes instead")
	set.StringVar(&buildDir, "css-dir", "",
		"Compass backwards compat. Reference locations relative to Sass project directory")
	set.MarkDeprecated("css-dir", "Use -b instead")
	set.StringVar(&jsDir, "javascripts-dir", "", "")
	set.MarkDeprecated("javascripts-dir", "Compass backwards compat, ignored")
	set.StringSliceVar(&includes, "sass-dir", nil,
		"Compass backwards compat, use --includes instead")
	set.StringVarP(&config, "config", "c", "",
		"Temporarily disabled: Location of the config file")

	set.StringVar(&cpuprofile, "cpuprofile", "", "Go runtime cpu profilling for debugging")
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile Sass stylesheets to CSS",
	Long: `Fast compilation of Sass stylesheets to CSS. For usage consult
the documentation at https://github.com/wellington/wellington#wellington`,
	Run: Compile,
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch Sass files for changes and rebuild CSS",
	Long:  ``,
	Run:   Watch,
}

var httpCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts a http server that will convert Sass to CSS",
	Long:  ``,
	Run:   Serve,
}

func init() {
	hostname := os.Getenv("HOSTNAME")
	if len(hostname) > 0 {
		if !strings.HasPrefix(hostname, "http") {
			hostname = "http://" + hostname
		}
	} else if host, err := os.Hostname(); err == nil {
		hostname = "http://" + host
	}
	httpCmd.Flags().StringVar(&httpPath, "httppath", hostname,
		"Only for HTTP, overrides generated sprite paths to support http")

}

func root() {
	flags(wtCmd.PersistentFlags())
}

// AddCommands attaches the cli subcommands ie. http, compile to the
// main cli entrypoint.
func AddCommands() {
	wtCmd.AddCommand(httpCmd)
	wtCmd.AddCommand(compileCmd)
	wtCmd.AddCommand(watchCmd)
}

var wtCmd = &cobra.Command{
	Use:   "wt",
	Short: "wt is a Sass project tool made to handle large projects. It uses the libSass compiler for efficiency and speed.",
	Run:   Compile,
}

func main() {
	AddCommands()
	root()
	wtCmd.Execute()
}

func argExit() bool {

	if showVersion {
		fmt.Printf("   libsass: %s\n", libsass.Version())
		fmt.Printf("Wellington: %s\n", version.Version)
		return true
	}

	if showHelp {
		fmt.Println("Please specify input filepath.")
		fmt.Println("\nAvailable options:")
		//flag.PrintDefaults()
		return true
	}
	return false

}

func makeabs(wd string, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(wd, path)
}

func parseBuildArgs(paths []string) *wt.BuildArgs {
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

func globalRun(paths []string) (*wt.SafePartialMap, *wt.BuildArgs) {
	// fmt.Printf("paths: %s args: % #v\n", paths, pflag.Args())
	if argExit() {
		return nil, nil
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

	for _, v := range paths {
		if strings.HasPrefix(v, "-") {
			log.Fatalf("Please specify flags before other arguments: %s", v)
		}
	}

	if gen != "" {
		err := os.MkdirAll(gen, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	pMap := wt.NewPartialMap()
	gba := parseBuildArgs(paths)
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
func Watch(cmd *cobra.Command, paths []string) {
	pMap, gba := globalRun(paths)
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
func Serve(cmd *cobra.Command, paths []string) {

	_, gba := globalRun(paths)
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
func Compile(cmd *cobra.Command, paths []string) {
	start := time.Now()
	pMap, gba := globalRun(paths)
	if gba == nil {
		return
	}

	defer func() {
		log.Printf("Compilation took: %s\n", time.Since(start))
	}()

	run(paths, pMap, gba)
}

// Run is the main entrypoint for the cli.
func run(paths []string, pMap *wt.SafePartialMap, gba *wt.BuildArgs) {

	// No paths given, read from stdin and wait
	if len(paths) == 0 {

		fmt.Println("Reading from stdin, -h for help")
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
	for _, s := range gba.Payload.Sprite().M {
		img.Add(1)
		err := s.Wait()
		img.Done()
		if err != nil {
			log.Printf("error writing sprite: %s\n", err)
		}
	}
	img.Wait()
	pMap.RUnlock()

}
