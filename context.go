package sprite_sass

/*
#cgo LDFLAGS: -Llibsass -lsass -lstdc++
#cgo CFLAGS: -Ilibsass

#include <stdlib.h>
#include <sass_interface.h>
*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"unsafe"

	"github.com/seateam/color"
)

// Context handles the interactions with libsass.  Context
// exposes libsass options that are available.
type Context struct {
	Parser                        Parser
	OutputStyle                   int
	Precision                     int
	Comments                      bool
	IncludePaths                  []string
	BuildDir, ImageDir, GenImgDir string
	Src, Out, Map, MainFile       string
	Sprites                       []ImageList
}

// Constants/enums for the output style.
const (
	NESTED_STYLE = iota
	EXPANDED_STYLE
	COMPACT_STYLE
	COMPRESSED_STYLE
)

var Style map[string]int

func init() {
	Style = make(map[string]int)
	Style["nested"] = NESTED_STYLE
	Style["expanded"] = EXPANDED_STYLE
	Style["compact"] = COMPACT_STYLE
	Style["compressed"] = COMPRESSED_STYLE
}

// Run uses the specified pathnames to read in sass and
// export out css with generated spritesheets based on
// the ImageDir option.  WriteCloser is necessary to
// notify readers when the stream is finished.
func (ctx *Context) Run(in io.Reader, out io.WriteCloser, pkgdir string) error {

	if in == nil {
		return errors.New("Input or output files were not specified")
	}

	if pkgdir != "" {
		ctx.IncludePaths = append(ctx.IncludePaths, pkgdir)
	}

	if ctx.GenImgDir == "" {
		ctx.GenImgDir = ctx.BuildDir
	}

	// Run the sprite_sass parser prior to passing to libsass
	ctx.Parser = Parser{
		ImageDir:  ctx.ImageDir,
		Includes:  ctx.IncludePaths,
		BuildDir:  ctx.BuildDir,
		GenImgDir: ctx.GenImgDir,
	}
	bs, err := ctx.Parser.Start(in, pkgdir)
	if err != nil {
		return err
	}
	ctx.Src = string(bs)

	err = ctx.Compile()

	obuf := bytes.NewBufferString(ctx.Out)
	defer out.Close()
	io.Copy(out, obuf)

	if err != nil {
		return err
	}

	return nil
}

// Compile passes off the sass compliant string to
// libsass for generating the resulting css file.
func (ctx *Context) Compile() error {

	if ctx.Precision == 0 {
		ctx.Precision = 5
	}

	if ctx.Src == "" {
		log.Fatal("No input string specified")
	}
	// set up the underlying C context struct
	cCtx := C.sass_new_context()
	cCtx.source_string = C.CString(ctx.Src)
	cCtx.options.output_style = C.int(ctx.OutputStyle)
	cCtx.options.source_comments = C.bool(ctx.Comments)
	cCtx.options.include_paths = C.CString(strings.Join(ctx.IncludePaths, ":"))
	cCtx.options.image_path = C.CString(ctx.ImageDir)
	cCtx.options.precision = C.int(ctx.Precision)
	defer func() {
		C.free(unsafe.Pointer(cCtx.source_string))
		C.free(unsafe.Pointer(cCtx.options.include_paths))
		C.free(unsafe.Pointer(cCtx.options.image_path))
		C.sass_free_context(cCtx)
	}()

	// Call the libsass compile function to populate the C context
	C.sass_compile(cCtx)

	// Populate Gocontext with results from c compiler
	ctx.Out = C.GoString(cCtx.output_string)
	ctx.Map = C.GoString(cCtx.source_map_string)
	errString := strings.TrimSpace(C.GoString(cCtx.error_message))
	err := errors.New(errString)
	if err.Error() == "" {
		err = nil
	} else {
		// Attempt to find the source error
		split := strings.Split(err.Error(), ":")
		if len(split) == 0 {
			return err
		}
		pos, lerr := strconv.Atoi(split[1])
		if lerr != nil {
			return lerr
		}
		lines := strings.Split(ctx.Src, "\n")
		// Line number is off by one from libsass
		// Find previous lines to maximum available
		errLines := "error in " + ctx.Parser.LookupFile(pos)
		red := color.NewStyle(color.BlackPaint, color.RedPaint).Brush()
		first := pos - 7
		if first < 0 {
			first = 0
		}
		last := pos + 7
		if last > len(lines) {
			last = len(lines)
		}
		for i := first; i < last; i++ {
			// translate 0 index to 1 index
			str := fmt.Sprintf("\n%3d: %s", i+1, lines[i])
			if i == pos-1 {
				str = red(str)
			}
			errLines += str
		}

		err = errors.New(err.Error() + "\n" + errLines)
	}
	return err
}
