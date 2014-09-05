package sprite_sass

/*
#cgo LDFLAGS: -Llibsass -lsass -lstdc++
#cgo CFLAGS: -Ilibsass

#include <stdlib.h>
#include <sass_interface.h>
*/
import "C"

import (
	"log"
	"strings"
	"unsafe"
)

type Options struct {
	OutputStyle    int
	SourceComments bool
	IncludePaths   []string
	ImagePath      string
	// eventually gonna' have things like callbacks and whatnot
}

type Context struct {
	Options
	SourceString string
	OutputString string
	ErrorStatus  int
	ErrorMessage string
}

type FileContext struct {
	Options
	InputPath    string
	OutputString string
	ErrorStatus  int
	ErrorMessage string
}

type SassContext struct {
	Context       *Context
	Sprites       []ImageList
	Input, Output string
}

func (ctx SassContext) Compile() {
	if ctx.Context.SourceString == "" {
		log.Fatal("No input/output file specified")
	}
	libCompile(ctx.Context)
}

// Constants/enums for the output style.
const (
	NESTED_STYLE = iota
	EXPANDED_STYLE
	COMPACT_STYLE
	COMPRESSED_STYLE
)

func libCompile(goCtx *Context) {
	// set up the underlying C context struct
	cCtx := C.sass_new_context()
	cCtx.source_string = C.CString(goCtx.SourceString)
	defer C.free(unsafe.Pointer(cCtx.source_string))
	cCtx.options.output_style = C.int(goCtx.Options.OutputStyle)
	if goCtx.Options.SourceComments {
		cCtx.options.source_comments = C.int(1)
	} else {
		cCtx.options.source_comments = C.int(0)
	}
	cCtx.options.include_paths = C.CString(strings.Join(goCtx.Options.IncludePaths, ":"))
	defer C.free(unsafe.Pointer(cCtx.options.include_paths))
	cCtx.options.image_path = C.CString(goCtx.Options.ImagePath)
	defer C.free(unsafe.Pointer(cCtx.options.image_path))
	// call the underlying C compile function to populate the C context
	C.sass_compile(cCtx)
	// extract values from the C context to populate the Go context object
	goCtx.OutputString = C.GoString(cCtx.output_string)
	goCtx.ErrorStatus = int(cCtx.error_status)
	goCtx.ErrorMessage = C.GoString(cCtx.error_message)
	// don't forget to free the C context!
	C.sass_free_context(cCtx)
}
