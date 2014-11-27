package context

/*
#cgo LDFLAGS: -lsass -lstdc++ -lm
#cgo CFLAGS:

#include <stdlib.h>
#include "sass_context.h"
#include "sass_functions.h"

static union Sass_Value* CallSassFunction( union Sass_Value* s_args, void* cookie);
*/
import "C"

import (
	"errors"
	"image/color"
	"io"
	"io/ioutil"

	"unsafe"
)

//export customHandler
func customHandler(args *C.union_Sass_Value, ptr unsafe.Pointer) *C.union_Sass_Value {
	// Recover the lane int from the pointer,
	// this may not be safe to do
	lane := *(*int)(ptr)
	arglen := int(C.sass_list_get_length(args))
	infs := make([]interface{}, arglen)
	for i := 0; i < arglen; i++ {
		arg := C.sass_list_get_value(args, C.size_t(i))
		infs[i] = SVtoInf(arg)
	}
	// Reference to original context
	swim := Pool[lane]
	swim.values = append(swim.values, infs)
	return C.sass_make_boolean(false)
}

// SVtoInf converts Sass Value to Go compatible data types.
func SVtoInf(arg *C.union_Sass_Value) interface{} {
	switch {
	case bool(C.sass_value_is_null(arg)):
		return nil
	case bool(C.sass_value_is_number(arg)):
		return int(C.sass_number_get_value(arg))
	case bool(C.sass_value_is_string(arg)):
		c := C.sass_string_get_value(arg)
		return C.GoString(c)
	case bool(C.sass_value_is_boolean(arg)):
		return bool(C.sass_boolean_get_value(arg))
	case bool(C.sass_value_is_color(arg)):
		col := color.RGBA{
			R: uint8(C.sass_color_get_r(arg)),
			G: uint8(C.sass_color_get_g(arg)),
			B: uint8(C.sass_color_get_b(arg)),
			A: uint8(C.sass_color_get_a(arg)),
		}
		return col
	case bool(C.sass_value_is_list(arg)):
		l := make([]interface{}, C.sass_list_get_length(arg))
		for i := range l {
			l[i] = SVtoInf(C.sass_list_get_value(arg, C.size_t(i)))
		}
		return l
	case bool(C.sass_value_is_map(arg)):
		len := int(C.sass_map_get_length(arg))
		m := make(map[interface{}]interface{}, len)
		for i := 0; i < len; i++ {
			m[SVtoInf(C.sass_map_get_key(arg, C.size_t(i)))] =
				SVtoInf(C.sass_map_get_value(arg, C.size_t(i)))
		}
		return m
	case bool(C.sass_value_is_error(arg)):
		return C.GoString(C.sass_error_get_message(arg))
	}
	return nil
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

	//TODO: Remove this likely, here now for easier testing
	values []interface{}
}

// Constants/enums for the output style.
const (
	NESTED_STYLE = iota
	EXPANDED_STYLE
	COMPACT_STYLE
	COMPRESSED_STYLE
)

var Style map[string]int

var Pool []*Context

func init() {
	Style = make(map[string]int)
	Style["nested"] = NESTED_STYLE
	Style["expanded"] = EXPANDED_STYLE
	Style["compact"] = COMPACT_STYLE
	Style["compressed"] = COMPRESSED_STYLE

}

// Init validates options in the struct and returns a Sass Options.
func (ctx *Context) Init(dc *C.struct_Sass_Data_Context) *C.struct_Sass_Options {
	if ctx.Precision == 0 {
		ctx.Precision = 5
	}
	cmt := C.bool(ctx.Comments)
	imgpath := C.CString(ctx.ImageDir)
	prec := C.int(ctx.Precision)

	opts := C.sass_data_context_get_options(dc)

	defer func() {
		C.free(unsafe.Pointer(imgpath))
		// C.free(unsafe.Pointer(cc))
		// C.sass_delete_data_context(dc)
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
	return opts
}

// Compile reads in and writes the libsass compiled result to out.
// Options and custom functions are applied as specified in Context.
func (ctx *Context) Compile(in io.Reader, out io.Writer) error {

	bs, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	if len(bs) == 0 {
		return errors.New("No input provided")
	}
	src := C.CString(string(bs))
	defer C.free(unsafe.Pointer(src))

	dc := C.sass_make_data_context(src)
	defer C.sass_delete_data_context(dc)

	opts := ctx.Init(dc)
	// TODO: Manually free options memory without throwing
	// malloc errors
	// defer C.free(unsafe.Pointer(opts))
	C.sass_data_context_set_options(dc, opts)
	cc := C.sass_data_context_get_context(dc)
	compiler := C.sass_make_data_compiler(dc)

	C.sass_compiler_parse(compiler)
	C.sass_compiler_execute(compiler)
	defer func() {
		C.sass_delete_compiler(compiler)
	}()

	cout := C.GoString(C.sass_context_get_output_string(cc))
	io.WriteString(out, cout)

	ctx.Status = int(C.sass_context_get_error_status(cc))
	errJson := C.sass_context_get_error_json(cc)
	errS := ctx.ProcessSassError([]byte(C.GoString(errJson)))

	if errS != "" {
		return errors.New(errS)
	}

	return nil
}
