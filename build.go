package wellington

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	libsass "github.com/wellington/go-libsass"
)

type BuildOptions struct {
	wg      sync.WaitGroup
	closing chan struct{}

	workwg sync.WaitGroup
	err    error
	done   chan error
	status chan error
	queue  chan string

	Async      bool
	Paths      []string
	BArgs      BuildArgs
	PartialMap *SafePartialMap
}

// Build compiles all valid Sass files found in the passed paths.
// If not async, this call will block until all files are compiled.
func (b *BuildOptions) Build() error {
	b.done = make(chan error)
	b.queue = make(chan string)
	b.closing = make(chan struct{})

	if b.PartialMap == nil {
		return errors.New("No partial map found")
	}

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.doBuild()
	}()
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.workwg.Wait()
		close(b.done)
	}()

	go b.loadWork()

	return <-b.done
}

func (b *BuildOptions) loadWork() {
	for _, path := range b.Paths {
		b.queue <- path
	}
}

func (b *BuildOptions) doBuild() {
	var err error
	for {
		if err != nil {
			return
		}
		select {
		case <-b.closing:
			return
		case path := <-b.queue:
			b.workwg.Add(1)
			go func() {
				err = b.build(path)
				b.workwg.Done()
				if err != nil {
					b.done <- err
				}
			}()
		}
	}
}

func (b *BuildOptions) build(path string) error {
	return LoadAndBuild(path, b.BArgs, b.PartialMap)
}

// Close shuts down the builder ensuring all go routines have properly
// closed before returning.
func (b *BuildOptions) Close() error {
	close(b.closing)
	b.wg.Wait()
	return nil
}

var inputFileTypes = []string{".scss", ".sass"}

// LoadAndBuild kicks off parser and compiling
// TODO: make this function testable
func LoadAndBuild(path string, gba BuildArgs, pMap *SafePartialMap) error {
	var files []string
	// file detected!
	if isImportable(path) {
		return loadAndBuild(path, gba, pMap)
	}

	// Expand directory to all non-partial sass files
	files, err := recursePath(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = loadAndBuild(file, gba, pMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadAndBuild(sassFile string, gba BuildArgs, partialMap *SafePartialMap) error {

	// If no imagedir specified, assume relative to the input file
	if gba.Dir == "" {
		gba.Dir = filepath.Dir(sassFile)
	}
	var (
		out  io.WriteCloser
		fout string
	)
	if gba.BuildDir != "" {
		// Build output file based off build directory and input filename
		rel, _ := filepath.Rel(gba.Includes, filepath.Dir(sassFile))
		filename := updateFileOutputType(filepath.Base(sassFile))
		fout = filepath.Join(gba.BuildDir, rel, filename)
	} else {
		out = os.Stdout
	}
	ctx := libsass.NewContext()
	ctx.Payload = gba.Payload
	ctx.OutputStyle = gba.Style
	ctx.ImageDir = gba.Dir
	ctx.FontDir = gba.Font
	// Assumption that output is a file
	ctx.BuildDir = filepath.Dir(fout)
	ctx.GenImgDir = gba.Gen
	ctx.MainFile = sassFile
	ctx.Comments = gba.Comments
	ctx.IncludePaths = []string{filepath.Dir(sassFile)}

	ctx.Imports.Init()
	if gba.Includes != "" {
		ctx.IncludePaths = append(ctx.IncludePaths,
			strings.Split(gba.Includes, ",")...)
	}
	fRead, err := os.Open(sassFile)
	if err != nil {
		return err
	}
	defer fRead.Close()
	if fout != "" {
		dir := filepath.Dir(fout)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("Failed to create directory: %s", dir)
		}

		out, err = os.Create(fout)
		defer out.Close()
		if err != nil {
			return fmt.Errorf("Failed to create file: %s", sassFile)
		}
		// log.Println("Created:", fout)
	}
	err = ctx.FileCompile(sassFile, out)
	if err != nil {
		return errors.New(color.RedString("%s", err))
	}

	for _, inc := range ctx.ResolvedImports {
		partialMap.AddRelation(ctx.MainFile, inc)
	}

	go fmt.Printf("Rebuilt: %s\n", sassFile)
	return nil
}

func updateFileOutputType(filename string) string {
	for _, filetype := range inputFileTypes {
		filename = strings.Replace(filename, filetype, ".css", 1)
	}
	return filename
}
