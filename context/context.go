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
	"log"

	"unsafe"
)

//export customHandler
func customHandler() C.int {
	log.Print("HELLO THERE")
	return C.int(1)
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

// Unused, currently
type CustomList C.Sass_C_Function_List

//type CustomDesc C.Sass_C_Function_Descriptor
type CustomFn C.Sass_C_Function_Callback

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

		C.sass_delete_data_context(dc)
	}()

	if len(ctx.Customs) > 0 {
		// Find out the size of a function
		dummy := C.sass_make_function(C.CString(""), C.Sass_C_Function(C.CallSassFunction), nil)
		size := C.size_t(len(ctx.Customs) + 1)
		_, _ = dummy, size
		fns_len := size * C.size_t(unsafe.Sizeof(dummy))
		fns := C.sass_make_function_list(fns_len)

		var fn C.Sass_C_Function_Callback
		for i, v := range ctx.Customs {
			_ = i
			bs := []byte(v)
			cv := (*C.char)(unsafe.Pointer(&bs[0]))
			_ = cv // const *char

			fn = C.sass_make_function(C.CString(v), C.Sass_C_Function(C.CallSassFunction), nil)
			//bfns[i] = byte(*(*C.int)(unsafe.Pointer(fn)))
			C.sass_set_function(&fns, fn, C.int(i))
		}

		C.sass_option_set_c_functions(opts, fns)
	}

	C.sass_option_set_precision(opts, prec)

	C.sass_option_set_source_comments(opts, cmt)

	C.sass_data_context_set_options(dc, opts)
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
