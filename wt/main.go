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
	"time"

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
)

/*
   -c, --config CONFIG_FILE         Specify the location of the configuration file explicitly.
       --app APP                    Tell compass what kind of application it is integrating with. E.g. rails
       --fonts-dir FONTS_DIR        The directory where you keep your fonts.
*/
func init() {

	// Interoperability args
}

func flags(set *pflag.FlagSet) {
	set.BoolVarP(&showVersion, "version", "v", false, "Show the app version")
	//wtCmd.PersistentFlags().BoolVarP(&showHelp, "help", "h", false, "this help")

	set.StringVar(&gen, "css-dir", "", "Compass Build Directory")
	set.StringVar(&dir, "images-dir", "", "Compass Image Directory")
	set.StringVar(&includes, "sass-dir", "", "Compass Sass Directory")
	set.StringVar(&jsDir, "javascripts-dir", "", "Compass JS Directory")
	set.BoolVar(&timeB, "time", false, "Retrieve timing information")

	set.StringVar(&buildDir, "b", "", "Build Directory")
	set.StringVar(&gen, "gen", ".", "Generated images directory")

	set.StringVar(&includes, "proj", "", "Project directory")
	set.StringVar(&includes, "p", "", "Project directory")
	set.StringVar(&dir, "dir", "", "Image directory")
	set.StringVar(&dir, "d", "", "Image directory")
	set.StringVar(&font, "font", ".", "Font Directory")

	set.StringVar(&style, "style", "nested", "CSS nested style")
	set.StringVar(&style, "s", "nested", "CSS nested style")
	set.BoolVar(&comments, "comment", true, "Turn on source comments")
	set.BoolVar(&comments, "c", true, "Turn on source comments")

	set.BoolVar(&ishttp, "http", false, "Listen for http connections")
	set.StringVar(&httpPath, "httppath", "",
		"Only for HTTP, overrides generated sprite paths to support http")
	set.BoolVar(&watch, "watch", false, "File watcher that will rebuild css on file changes")
	set.BoolVar(&watch, "w", false, "File watcher that will rebuild css on file changes")

	set.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")

}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile Sass stylesheets to CSS",
	Long: `Fast compilation of Sass stylesheets to CSS. For usage consult
the documentation at https://github.com/wellington/wellington#wellington`,
	Run: Run,
}

func root() {
	flags(wtCmd.PersistentFlags())
}

func AddCommands() {
	wtCmd.AddCommand(compileCmd)
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

	if showVersion {
		fmt.Printf("Wellington: %s\n", version.Version)
		fmt.Printf("   libsass: %s\n", libsass.Version())
		os.Exit(0)
	}

	if showHelp {
		fmt.Println("Please specify input filepath.")
		fmt.Println("\nAvailable options:")
		//flag.PrintDefaults()
		return
	}
}

func Run(cmd *cobra.Command, files []string) {
	start := time.Now()

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

	for _, v := range files {
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
	ctx := &libsass.Context{
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

	if len(files) == 0 {

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
		return
	}

	sassPaths := make([]string, len(files))
	for i, f := range files {
		sassPaths[i] = filepath.Dir(f)
		err := wt.LoadAndBuild(f, gba, pMap)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}

	if watch {
		w := wt.NewWatcher()
		w.PartialMap = pMap
		w.Dirs = sassPaths
		w.BArgs = gba
		//w.Watch()

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
