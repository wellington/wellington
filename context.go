package sprite_sass

/*
#cgo LDFLAGS: -Llibsass -lsass -lstdc++ -lm
#cgo CFLAGS: -Ilibsass

#include <stdlib.h>
#include <sass_interface.h>
*/
import "C"

import (
	"errors"
	"io"
	"log"
	"strings"
	"unsafe"

	. "github.com/drewwells/spritewell"
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
	errorString                   string
	errors                        lErrors
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
func (ctx *Context) Run(in io.Reader, out io.Writer, pkgdir string) error {
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
	ctx.Compile()

	io.WriteString(out, ctx.Out)

	if len(ctx.Error()) == 0 {
		return nil
	}

	return errors.New(ctx.Error())
}

// Compile passes off the sass compliant string to
// libsass for generating the resulting css file.
func (ctx *Context) Compile() {

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
	// Set the internal error string to C error return
	ctx.ProcessSassError(C.GoString(cCtx.error_message))
}
