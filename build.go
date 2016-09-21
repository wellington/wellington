package wellington

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"github.com/fatih/color"
	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/wellington/payload"
)

var testch chan struct{}

// BuildArgs holds universal arguments for a build that the parser
// uses during the initial build and the filewatcher passes back to
// the parser on any file changes.
type BuildArgs struct {
	// paths are the initial directories passed to a build
	// This is required to output files
	paths []string

	Payload  context.Context
	ImageDir string

	// BuildDir is the base build directory used. When recursive
	// file matching is involved, this directory will be used as the
	// parent.
	BuildDir  string
	Includes  []string
	Font      string
	Gen       string
	Style     int
	Comments  bool
	CacheBust string
	// emit source map files alongside css files
	SourceMap bool
}

// Paths retrieves the paths in the arguments
func (b *BuildArgs) Paths() []string {
	return b.paths
}

// WithPaths creates a new BuildArgs with paths applied
func (b *BuildArgs) WithPaths(paths []string) {

	// sort paths prior to running to fix #187
	sorted := stringSize(paths)
	sort.Sort(sorted)

	b.paths = sorted
}

// Init initializes the payload, this should really go away
func (b *BuildArgs) init() {
	b.Payload = payload.New()
}

// Build holds a set of read only arguments to the builder.
// Channels from this are used to communicate between the workers
// and loaders executing builds.
type Build struct {
	wg      sync.WaitGroup
	closing chan struct{}

	workwg sync.WaitGroup
	err    error
	done   chan error
	status chan error
	queue  chan work

	paths      []string
	bArgs      *BuildArgs
	partialMap *SafePartialMap
}

type work struct {
	file string
}

// NewBuild accepts arguments to reate a new Builder
func NewBuild(args *BuildArgs, pMap *SafePartialMap) *Build {
	if args.Payload == nil {
		args.init()
	}

	return &Build{
		done:   make(chan error),
		status: make(chan error),

		queue:   make(chan work),
		closing: make(chan struct{}),

		paths:      args.Paths(),
		bArgs:      args,
		partialMap: pMap,
	}
}

// ErrPartialMap when no partial map is found
var ErrPartialMap = errors.New("No partial map found")

// Run compiles all valid Sass files found in the passed paths.
// It will block until all files are compiled.
func (b *Build) Run() error {

	if b.partialMap == nil {
		return ErrPartialMap
	}

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.doBuild()
	}()

	go b.loadWork()

	return <-b.done
}

func (b *Build) loadWork() {
	files := pathsToFiles(b.paths, true)
	for _, file := range files {
		b.queue <- work{file: file}
	}
	close(b.queue)
}

func (b *Build) doBuild() {
	for {
		select {
		case <-b.closing:
			return
		case work := <-b.queue:
			path := work.file
			if len(path) == 0 {
				b.workwg.Wait()
				b.done <- nil
				return
			}
			b.workwg.Add(1)
			go func(path string) {
				err := b.build(path)
				defer b.workwg.Done()
				if err != nil {
					b.done <- err
				}
			}(path)
		}
	}
}

func (b *Build) build(path string) error {
	if len(path) == 0 {
		return errors.New("invalid path given")
	}

	if !isImportable(path) {
		return errors.New("file does not end in .sass or .scss")
	}

	out, sout, bdir, err := b.bArgs.getOut(path)
	if err != nil {
		return err
	}

	return loadAndBuild(path, b.bArgs, b.partialMap, out, sout, bdir)
}

// Close shuts down the builder ensuring all go routines have properly
// closed before returning.
func (b *Build) Close() error {
	close(b.closing)
	b.wg.Wait()
	return nil
}

var inputFileTypes = []string{".scss", ".sass"}

func (b *BuildArgs) getOut(path string) (io.WriteCloser, io.WriteCloser, string, error) {

	var (
		out io.WriteCloser
	)
	if b == nil {
		return nil, nil, "", errors.New("build args is nil")
	}
	if len(b.BuildDir) == 0 {
		out = os.Stdout
		return out, nil, "", nil
	}
	rel := relative(b.paths, path)
	filename := updateFileOutputType(filepath.Base(path))
	name := filepath.Join(b.BuildDir, rel, filename)
	dir := filepath.Dir(name)
	// FIXME: do this once per Build instead of every file
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, nil, "", fmt.Errorf("Failed to create directory: %s",
			dir)
	}
	out, err = os.Create(name)
	if err != nil {
		return nil, nil, "", err
	}
	var smap *os.File
	if b.SourceMap {
		smap, err = os.Create(name + ".map")
	}

	return out, smap, dir, err
}

// LoadAndBuild kicks off parser and compiling. It expands directories
// to recursively locate Sass files
// TODO: make this function testable
func LoadAndBuild(path string, gba *BuildArgs, pMap *SafePartialMap) error {
	if len(path) == 0 {
		return errors.New("invalid path passed")
	}

	// file detected!
	if isImportable(path) {
		out, sout, bdir, err := gba.getOut(path)
		if err != nil {
			return err
		}
		return loadAndBuild(path, gba, pMap, out, sout, bdir)
	}

	out, sout, bdir, err := gba.getOut(path)
	if err != nil {
		return err
	}
	err = loadAndBuild(path, gba, pMap, out, sout, bdir)
	if err != nil {
		return err
	}

	return nil
}

// FromBuildArgs creates a compiler from BuildArgs
func FromBuildArgs(dst io.Writer, dstmap io.Writer, src io.Reader, gba *BuildArgs) (libsass.Compiler, error) {
	if gba == nil {
		return libsass.New(dst, src)
	}
	if gba.Payload == nil {
		gba.init()
	}

	comp, err := libsass.New(dst, src,
		// Options overriding defaults
		// libsass.Path(sassFile), what path should be provided?
		libsass.ImgDir(gba.ImageDir),
		libsass.ImgBuildDir(gba.Gen),
		libsass.BuildDir(gba.BuildDir),
		libsass.Payload(gba.Payload),
		libsass.Comments(gba.Comments),
		libsass.OutputStyle(gba.Style),
		libsass.FontDir(gba.Font),
		libsass.IncludePaths(gba.Includes),
		libsass.CacheBust(gba.CacheBust),
		libsass.SourceMap(gba.SourceMap, dstmap),
	)
	return comp, err
}

func loadAndBuild(sassFile string, gba *BuildArgs, partialMap *SafePartialMap, out io.WriteCloser, sout io.WriteCloser, buildDir string) error {
	defer func() {
		// BuildDir lets us know if we should closer out. If no buildDir,
		// specified out == os.Stdout and do not close. If buildDir != "",
		// then out must be something we should close.
		// This is important, since out can be many things and inspecting
		// them could be race unsafe.
		if len(buildDir) > 0 {
			out.Close()
			sout.Close()
		}
	}()

	// FIXME: move this elsewhere or make it so it doesn't need to be set
	imgdir := gba.ImageDir
	if len(imgdir) == 0 {
		imgdir = filepath.Dir(sassFile)
	}

	comp, err := libsass.New(out, nil,
		// Options overriding defaults
		libsass.Path(sassFile),
		libsass.ImgDir(imgdir),
		libsass.BuildDir(buildDir),
		libsass.Payload(gba.Payload),
		libsass.Comments(gba.Comments),
		libsass.OutputStyle(gba.Style),
		libsass.FontDir(gba.Font),
		libsass.ImgBuildDir(gba.Gen),
		libsass.IncludePaths(gba.Includes),
		libsass.SourceMap(gba.SourceMap, sout),
	)

	if err != nil {
		return err
	}

	// Start Sass transformation
	err = comp.Run()
	if err != nil {
		return errors.New(color.RedString("%s", err))
	}
	for _, inc := range comp.Imports() {
		partialMap.AddRelation(sassFile, inc)
	}

	// TODO: moves this method to *Build and wait on it to finish
	// go func(file string) {
	select {
	case <-testch:
	default:

	}
	// }(sassFile)
	return nil
}

func updateFileOutputType(filename string) string {
	for _, filetype := range inputFileTypes {
		filename = strings.Replace(filename, filetype, ".css", 1)
	}
	return filename
}
