// Main package wraps sprite_sass tool for use with the command line
// See -h for list of available options
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/wellington/cfg"
	"github.com/wellington/wellington/version"

	wt "github.com/wellington/wellington"
	_ "github.com/wellington/wellington/handlers"
)

var (
	font, dir, gen, includes      string
	mainFile, style               string
	comments, watch               bool
	cpuprofile, buildDir          string
	jsDir                         string
	ishttp, showHelp, showVersion bool
	httpPath                      string
	timeB                         bool
	config                        string
	debug                         bool
	multi                         bool

	// unused
	noLineComments bool
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
	set.BoolVar(&noLineComments, "no-line-comments", false, "UNSUPPORTED: Disable line comments")
	set.BoolVar(&relativeAssets, "relative-assets", false, "UNSUPPORTED: Make compass asset helpers generate relative urls to assets.")

	set.BoolVarP(&showVersion, "version", "v", false, "Show the app version")
	//wtCmd.PersistentFlags().BoolVarP(&showHelp, "help", "h", false, "this help")
	set.BoolVar(&debug, "debug", false, "Show detailed debug information")
	set.StringVar(&dir, "images-dir", "", "Compass Image Directory")
	set.StringVarP(&dir, "dir", "d", "", "Compass Image Directory")
	set.StringVar(&jsDir, "javascripts-dir", "", "Compass JS Directory")
	set.BoolVar(&timeB, "time", false, "Retrieve timing information")

	set.StringVarP(&buildDir, "build", "b", "", "Target directory for generated CSS, relative paths from sass-dir are preserved")
	set.StringVar(&buildDir, "css-dir", "", "Location of CSS files")

	set.StringVar(&gen, "gen", ".", "Generated images directory")

	set.StringVar(&includes, "sass-dir", "", "Compass Sass Directory")
	set.StringVarP(&includes, "proj", "p", "", "Project directory")

	set.StringVar(&font, "font", ".", "Font Directory")
	set.StringVarP(&style, "style", "s", "nested", "CSS nested style")

	set.StringVarP(&config, "config", "c", "", "Location of the config file")

	set.BoolVarP(&comments, "comment", "", true, "Turn on source comments")

	set.BoolVarP(&watch, "watch", "w", false, "File watcher that will rebuild css on file changes")

	set.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	set.BoolVarP(&multi, "multi", "", false, "Enable multi-threaded operation")
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile Sass stylesheets to CSS",
	Long: `Fast compilation of Sass stylesheets to CSS. For usage consult
the documentation at https://github.com/wellington/wellington#wellington`,
	Run: Run,
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch Sass files for changes and rebuild CSS",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		watch = true
		Run(cmd, args)
	},
}

var httpCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts a http server that will convert Sass to CSS",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ishttp = true
		Run(cmd, args)
	},
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
	Short: "wt builds Sass",
	Run:   Run,
}

func main() {
	AddCommands()
	root()

	wtCmd.Execute()
}

// Run is the main entrypoint for the cli.
func Run(cmd *cobra.Command, paths []string) {

	start := time.Now()

	if showVersion {
		fmt.Printf("   libsass: %s\n", libsass.Version())
		fmt.Printf("Wellington: %s\n", version.Version)
		os.Exit(0)
	}

	if showHelp {
		fmt.Println("Please specify input filepath.")
		fmt.Println("\nAvailable options:")
		//flag.PrintDefaults()
		os.Exit(0)
	}

	defer func() {
		diff := float64(time.Since(start).Nanoseconds()) / float64(time.Millisecond)
		log.Printf("Compilation took: %sms\n",
			strconv.FormatFloat(diff, 'f', 3, 32))
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

	style, ok := libsass.Style[style]

	if !ok {
		style = libsass.NESTED_STYLE
	}

	if len(config) > 0 {
		cfg, err := cfg.Parse(config)
		if err != nil {
			log.Fatal(err)
		}
		// Manually walk through known variables looking for matches
		// These do not override the cli flags
		if p, ok := cfg["css_dir"]; ok && len(buildDir) == 0 {
			buildDir = p
		}

		if p, ok := cfg["images_dir"]; ok && len(dir) == 0 {
			dir = p
		}

		if p, ok := cfg["sass_dir"]; ok && len(includes) == 0 {
			includes = p
		}

		if p, ok := cfg["generated_images_dir"]; ok && len(gen) == 0 {
			gen = p
		}

		// As of yet, unsupported
		if p, ok := cfg["http_path"]; ok {
			_ = p
		}

		if p, ok := cfg["http_generated_images_path"]; ok {
			_ = p
		}

		if p, ok := cfg["fonts_dir"]; ok {
			font = p
		}
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
	ctx := libsass.NewContext()
	ctx.Payload = gba.Payload
	ctx.OutputStyle = gba.Style
	ctx.BuildDir = gba.BuildDir
	ctx.ImageDir = gba.Dir
	ctx.FontDir = gba.Font
	ctx.GenImgDir = gba.Gen
	ctx.Comments = gba.Comments
	ctx.HTTPPath = httpPath
	ctx.IncludePaths = []string{gba.Includes}

	if debug {
		fmt.Printf("      Font  Dir: %s\n", gba.Font)
		fmt.Printf("      Image Dir: %s\n", gba.Dir)
		fmt.Printf("      Build Dir: %s\n", gba.BuildDir)
		fmt.Printf("Build Image Dir: %s\n", gba.Gen)
		fmt.Printf(" Include Dir(s): %s\n", gba.Includes)
		fmt.Println("===================================")
	}
	wt.InitializeContext(ctx)
	ctx.Imports.Init()

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

	// Only inject files when a config is passed. Otherwise,
	// assume we are waiting for input from stdin
	if len(includes) > 0 && len(config) > 0 {
		rot := filepath.Join(includes, "*.scss")
		pat := filepath.Join(includes, "**/*.scss")
		rotFiles, _ := filepath.Glob(rot)
		patFiles, _ := filepath.Glob(pat)
		paths = append(rotFiles, patFiles...)
		// Probably a better way to do this, but I'm impatient

		clean := make([]string, 0, len(paths))

		for _, p := range paths {
			if !strings.HasPrefix(filepath.Base(p), "_") {
				clean = append(clean, p)
			}
		}
		paths = clean
	}

	if !watch && len(paths) == 0 && len(config) == 0 {

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
			color.Red(err.Error())
		}
		return
	}
	sassPaths := paths
	bOpts := &wt.BuildOptions{
		Async:      multi,
		Paths:      paths,
		BArgs:      gba,
		PartialMap: pMap,
	}

	err := bOpts.Build()
	if err != nil {
		log.Fatal(err)
	}

	if watch {
		w := wt.NewWatcher(&wt.WatchOptions{

			PartialMap: pMap,
			Paths:      sassPaths,
			BArgs:      gba,
		})
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
	} else {

		// Before shutting down, check that every sprite has been
		// flushed to disk.
		flush := time.Now()
		img := sync.WaitGroup{}
		pMap.RLock()
		// It's not currently possible to wait on Image. This is often
		// to inline images, so it shouldn't be a factor...
		/*for _, s := range gba.Payload.Image().M {
			img.Add(1)
			err := s.Wait()
			img.Done()
			if err != nil {
				log.Printf("error writing image: %s\n", err)
			}
		}*/
		for _, s := range gba.Payload.Sprite().M {
			img.Add(1)
			err := s.Wait()
			img.Done()
			if err != nil {
				log.Printf("error writing sprite: %s\n", err)
			}
		}
		img.Wait()
		_ = flush
		// log.Println("Extra time spent flushing images to disk: ", time.Since(flush))
		pMap.RUnlock()
	}
}
