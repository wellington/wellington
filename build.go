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

	"github.com/fatih/color"
	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/wellington/types"
)

var testch chan struct{}

// BuildArgs holds universal arguments for a build that the parser
// uses during the initial build and the filewatcher passes back to
// the parser on any file changes.
type BuildArgs struct {
	// paths are the initial directories passed to a build
	// This is required to output files
	paths []string

	// Imgs, Sprites spritewell.SafeImageMap
	Payload  types.Payloader
	ImageDir string

	// BuildDir is the base build directory used. When recursive
	// file matching is involved, this directory will be used as the
	// parent.
	BuildDir string
	Includes []string
	Font     string
	Gen      string
	Style    int
	Comments bool
}

// WithPaths creates a new BuildArgs with paths applied
func (b *BuildArgs) WithPaths(paths []string) {
	b.paths = paths
}

// Init initializes the payload, this should really go away
func (b *BuildArgs) init() {
	b.Payload = newPayload()
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
func NewBuild(paths []string, args *BuildArgs, pMap *SafePartialMap) *Build {
	if args.Payload == nil {
		args.init()
	}

	// paths should be sorted, so that the most specific relative path is
	// used for relative build path in build directory
	sort.Sort(sort.StringSlice(paths))
	return &Build{
		done:   make(chan error),
		status: make(chan error),

		queue:   make(chan work),
		closing: make(chan struct{}),

		paths:      paths,
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
		return errors.New("file is does not end in .sass or .scss")
	}

	out, bdir, err := b.bArgs.getOut(path)
	if err != nil {
		return err
	}

	return loadAndBuild(path, b.bArgs, b.partialMap, out, bdir)
}

// Close shuts down the builder ensuring all go routines have properly
// closed before returning.
func (b *Build) Close() error {
	close(b.closing)
	b.wg.Wait()
	return nil
}

var inputFileTypes = []string{".scss", ".sass"}

func (b *BuildArgs) getOut(path string) (io.WriteCloser, string, error) {
	var (
		out  io.WriteCloser
		fout string
	)
	if b == nil {
		return nil, "", errors.New("build args is nil")
	}
	if len(b.BuildDir) == 0 {
		out = os.Stdout
		return out, "", nil
	}
	rel := relative(b.paths, path)
	filename := updateFileOutputType(filepath.Base(path))
	fout = filepath.Join(b.BuildDir, rel, filename)
	dir := filepath.Dir(fout)
	// FIXME: do this once per Build instead of every file
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to create directory: %s",
			dir)
	}
	out, err = os.Create(fout)
	if err != nil {
		return nil, "", err
	}
	return out, dir, nil
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
		out, bdir, err := gba.getOut(path)
		if err != nil {
			return err
		}
		return loadAndBuild(path, gba, pMap, out, bdir)
	}

	out, bdir, err := gba.getOut(path)
	if err != nil {
		return err
	}
	err = loadAndBuild(path, gba, pMap, out, bdir)
	if err != nil {
		return err
	}

	return nil
}

// FromBuildArgs creates a compiler from BuildArgs
func FromBuildArgs(dst io.Writer, src io.Reader, gba *BuildArgs) (libsass.Compiler, error) {
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
	)
	return comp, err
}

func loadAndBuild(sassFile string, gba *BuildArgs, partialMap *SafePartialMap, out io.WriteCloser, buildDir string) error {
	defer func() {
		// BuildDir lets us know if we should closer out. If no buildDir,
		// specified out == os.Stdout and do not close. If buildDir != "",
		// then out must be something we should close.
		// This is important, since out can be many things and inspecting
		// them could be race unsafe.
		if len(buildDir) > 0 {
			out.Close()
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
	)
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
