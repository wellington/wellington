package wellington

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	context "github.com/wellington/libsass"
)

var inputFileTypes = []string{".scss", ".sass"}

// LoadAndBuild kicks off parser and compiling
// TODO: make this function testable
func LoadAndBuild(sassFile string, gba *BuildArgs, partialMap *SafePartialMap) error {

	if gba == nil {
		return fmt.Errorf("build args are nil")
	}

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
	ctx := context.Context{
		Sprites:     gba.Sprites,
		Imgs:        gba.Imgs,
		OutputStyle: gba.Style,
		ImageDir:    gba.Dir,
		FontDir:     gba.Font,
		// Assumption that output is a file
		BuildDir:     filepath.Dir(fout),
		GenImgDir:    gba.Gen,
		MainFile:     sassFile,
		Comments:     gba.Comments,
		IncludePaths: []string{filepath.Dir(sassFile)},
	}
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
		log.Println(sassFile)
		return err
	}

	for _, inc := range ctx.ResolvedImports {
		partialMap.AddRelation(ctx.MainFile, inc)
	}

	fmt.Printf("Rebuilt: %s\n", sassFile)
	return nil
}

func updateFileOutputType(filename string) string {
	for _, filetype := range inputFileTypes {
		filename = strings.Replace(filename, filetype, ".css", 1)
	}
	return filename
}
