package sprite_sass

/*
#cgo LDFLAGS: -Llibsass/lib -lsass -lstdc++ -lm
#cgo CFLAGS: -Ilibsass

#include <stdlib.h>
#include <sass_context.h>
*/
import "C"

import (
	"errors"
	"io"
	"log"
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
	Status                        int
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
	// Setup params for delivery to C
	src := C.CString(ctx.Src)
	// style := C.int(ctx.OutputStyle)
	cmt := C.bool(ctx.Comments)
	// This might be interesting instead of stringifying everything
	// inc := C.CString(strings.Join(ctx.IncludePaths, ":"))
	imgpath := C.CString(ctx.ImageDir)
	prec := C.int(ctx.Precision)

	// Create a data context from source string
	opts := C.sass_make_options()
	dc := C.sass_make_data_context(src)
	defer func() {
		C.free(unsafe.Pointer(src))
		// C.free(unsafe.Pointer(inc))
		C.free(unsafe.Pointer(imgpath))
		// How do you release these?
		// C.free(unsafe.Pointer(style))
		// C.free(unsafe.Pointer(cmt))
		// C.free(unsafe.Pointer(prec))
		C.sass_delete_data_context(dc)
	}()

	// Set passed options
	C.sass_option_set_precision(opts, prec)
	// C.sass_option_set_output_style(opts, style)
	C.sass_option_set_source_comments(opts, cmt)
	C.sass_data_context_set_options(dc, opts)
	_ = C.sass_compile_data_context(dc)
	cc := C.sass_data_context_get_context(dc)
	ctx.Out = C.GoString(C.sass_context_get_output_string(cc))
	ctx.Status = int(C.sass_context_get_error_status(cc))
	ctx.ProcessSassError([]byte(C.GoString(C.sass_context_get_error_json(cc))))
}
