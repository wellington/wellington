// Main package wraps sprite_sass tool for use with the command line
// See -h for list of available options
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	libsass "github.com/wellington/go-libsass"
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

	set.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	set.BoolVarP(&multi, "multi", "", true, "Enable multi-threaded operation")
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
	Short: "wt is a Sass project tool",
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

func parseBuildArgs(paths []string) *wt.BuildArgs {
	style, ok := libsass.Style[style]

	if !ok {
		style = libsass.NESTED_STYLE
	}
	incs := strings.Split(includes, ",")
	incs = append(incs, paths...)
	gba := &wt.BuildArgs{
		ImageDir: dir,
		BuildDir: buildDir,
		Includes: incs,
		Font:     font,
		Style:    style,
		Gen:      gen,
		Comments: comments,
	}
	gba.WithPaths(paths)

	return gba
}

func globalRun(paths []string) (*wt.SafePartialMap, *wt.BuildArgs) {

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
		fmt.Printf("      Font  Dir: %s\n", gba.Font)
		fmt.Printf("      Image Dir: %s\n", gba.ImageDir)
		fmt.Printf("      Build Dir: %s\n", gba.BuildDir)
		fmt.Printf("Build Image Dir: %s\n", gba.Gen)
		fmt.Printf(" Include Dir(s): %s\n", gba.Includes)
		fmt.Println("===================================")
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

// Serve starts a web server accepting POST calls and return CSS
func Serve(cmd *cobra.Command, paths []string) {

	_, gba := globalRun(paths)
	if len(gba.Gen) == 0 {
		log.Fatal("Must pass an image build directory to use HTTP")
	}

	addr := ":12345"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on: %s", addr)
	}

	certFile := "./server.crt"
	keyFile := "./server.key"
	_, _ = certFile, keyFile
	mux := http.NewServeMux()
	mux.Handle("/build/", wt.FileHandler(gba.Gen))
	mux.HandleFunc("/", wt.HTTPHandler(gba))
	log.Println("Web server started on :12345")

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	cfg := &tls.Config{
		// InsecureSkipVerify: true,
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{cert},
	}

	svr := &http.Server{Handler: mux, TLSConfig: cfg}
	// http2 client only works with TLS

	// Enable HTTP2 https://http2.golang.org/
	http2.ConfigureServer(svr, &http2.Server{})
	tlsLis := tls.NewListener(lis, svr.TLSConfig)
	log.Fatal(svr.Serve(tlsLis))
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
