package wellington

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wellington/wellington/context"
)

// LoadAndBuild kicks off parser and compiling
// TODO: make this function testable
func LoadAndBuild(sassFile string, gba *BuildArgs, partialMap *SafePartialMap) error {

	// Remove partials
	if strings.HasPrefix(filepath.Base(sassFile), "_") {
		return nil
	}

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
		filename := strings.Replace(filepath.Base(sassFile), ".scss", ".css", 1)
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
	if gba.Includes != "" {
		ctx.IncludePaths = append(ctx.IncludePaths,
			strings.Split(gba.Includes, ",")...)
	}
	fRead, err := os.Open(sassFile)
	defer fRead.Close()
	if err != nil {
		return err
	}
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

	var pout bytes.Buffer
	par, err := StartParser(&ctx, fRead, &pout, partialMap)
	if err != nil {
		return err
	}
	err = ctx.Compile(&pout, out)

	if err != nil {
		log.Println(ctx.MainFile)
		n := ctx.ErrorLine()
		fs := par.LookupFile(n)
		log.Printf("Error encountered in: %s\n", fs)
		return err
	}
	fmt.Printf("Rebuilt: %s\n", sassFile)
	return nil
}
