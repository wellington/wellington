package context

/*
#cgo LDFLAGS: -lsass -lstdc++ -lm
#cgo CFLAGS:

#include <stdlib.h>
#include <stdio.h>
#include "sass_context.h"
#include "sass_functions.h"

static union Sass_Value* CallSassFunction( union Sass_Value* s_args, void* cookie);
*/
import "C"

import (
	"errors"
	"io"
	"io/ioutil"

	"unsafe"
)

//export customHandler
func customHandler(ptr unsafe.Pointer) {
	// Recover the lane int from the pointer,
	// this may not be safe to do
	lane := *(*int)(ptr)
	_ = Pool[lane] // Reference to original context
	return
}

// Context handles the interactions with libsass.  Context
// exposes libsass options that are available.
type Context struct {
	//Parser                        Parser
	OutputStyle                   int
	Precision                     int
	Comments                      bool
	IncludePaths                  []string
	BuildDir, ImageDir, GenImgDir string
	In, Src, Out, Map, MainFile   string
	Status                        int
	errorString                   string
	errors                        lErrors

	in      io.Reader
	out     io.Writer
	Errors  SassError
	Customs []string
	Lane    int // Reference to pool position
}

// Constants/enums for the output style.
const (
	NESTED_STYLE = iota
	EXPANDED_STYLE
	COMPACT_STYLE
	COMPRESSED_STYLE
)

var Style map[string]int

var Pool []Context

func init() {
	Style = make(map[string]int)
	Style["nested"] = NESTED_STYLE
	Style["expanded"] = EXPANDED_STYLE
	Style["compact"] = COMPACT_STYLE
	Style["compressed"] = COMPRESSED_STYLE

}

// Libsass for generating the resulting css file.
func (ctx *Context) Compile(in io.Reader, out io.Writer) error {
	if ctx.Precision == 0 {
		ctx.Precision = 5
	}
	bs, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	if len(bs) == 0 {
		return errors.New("No input provided")
	}
	src := C.CString(string(bs))
	cmt := C.bool(ctx.Comments)
	imgpath := C.CString(ctx.ImageDir)
	prec := C.int(ctx.Precision)

	dc := C.sass_make_data_context(src)
	cc := C.sass_data_context_get_context(dc)
	opts := C.sass_context_get_options(cc)

	defer func() {
		C.free(unsafe.Pointer(src))
		C.free(unsafe.Pointer(imgpath))
		//C.free(unsafe.Pointer(cc))
		C.sass_delete_data_context(dc)
	}()

	// Set custom sass functions
	if len(ctx.Customs) > 0 {
		size := C.size_t(len(ctx.Customs) + 1)
		// TODO: Does this get cleaned up by sass_delete_data_context?
		fns := C.sass_make_function_list(size)
		for i, v := range ctx.Customs {
			fn := C.sass_make_function(C.CString(v),
				C.Sass_C_Function(C.CallSassFunction),
				unsafe.Pointer(&ctx.Lane))
			C.sass_set_function(&fns, fn, C.int(i))
		}

		C.sass_option_set_c_functions(opts, fns)
	}
	C.sass_option_set_precision(opts, prec)
	C.sass_option_set_source_comments(opts, cmt)
	C.sass_data_context_set_options(dc, opts)

	// Compile
	compiler := C.sass_make_data_compiler(dc)
	C.sass_compiler_parse(compiler)
	C.sass_compiler_execute(compiler)
	//C.sass_delete_compiler(compiler)

	_ = C.sass_compile_data_context(dc)
	cout := C.GoString(C.sass_context_get_output_string(cc))
	io.WriteString(out, cout)

	ctx.Status = int(C.sass_context_get_error_status(cc))
	errS := ctx.ProcessSassError([]byte(C.GoString(C.sass_context_get_error_json(cc))))

	if errS != "" {
		return errors.New(errS)
	}

	return nil
}
