package sprite_sass

/*
#cgo LDFLAGS: -Llibsass -lsass -lstdc++
#cgo CFLAGS: -Ilibsass

#include <stdlib.h>
#include <sass_interface.h>
*/
import "C"

import (
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"unsafe"
)

// Context handles the interactions with libsass.  Context
// exposes libsass options that are available.
type Context struct {
	OutputStyle   int
	Precision     int
	Comments      bool
	IncludePaths  []string
	ImageDir      string
	Src, Out, Map string
	Sprites       []ImageList
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
// the ImageDir option.
func (ctx *Context) Run(ipath, opath string) error {

	if ipath == "" || opath == "" {
		log.Fatal("Input or output files were not specified")
	}
	ctx.IncludePaths = append(ctx.IncludePaths, filepath.Dir(ipath))

	// Run the sprite_sass parser prior to passing to libsass
	parser := Parser{
		ImageDir: ctx.ImageDir,
		Includes: ctx.IncludePaths,
	}
	bytes := parser.Start(ipath)
	ctx.Src = string(bytes)

	err := ctx.Compile()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(opath, []byte(ctx.Out), 0777)
	return err
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
	if ctx.Comments {
		cCtx.options.source_comments = C.int(1)
	} else {
		cCtx.options.source_comments = C.int(0)
	}
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
	// Create Go style errors
	err := errors.New(errString)
	if err.Error() == "" {
		err = nil
	}

	return err
}
