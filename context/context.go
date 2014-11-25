package context

// Callback gateway of type
// union Sass_Value call_sass_function(const union Sass_Value s_args, void *cookie)

/*
#cgo LDFLAGS: -lsass -lstdc++ -lm
#cgo CFLAGS:


#include <stdlib.h>
#include "sass_context.h"
#include "sass_functions.h"
typedef void (*Callback)(int argnum);

void CallMe( Callback f ) {
  f(0);
}

*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"unsafe"
)

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

// export Custom
func Custom() {
	log.Print("HELLO THERE")
	return
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

	opts := C.sass_make_options()
	if len(ctx.Customs) > 0 {
		// Find out the size of a function
		dummy := C.sass_make_function(C.CString(""), (*[0]byte)(C.CallMe), nil)
		size := C.size_t(len(ctx.Customs) + 1)
		fns_len := size * C.size_t(unsafe.Sizeof(dummy))
		fns := C.sass_make_function_list(fns_len)

		fmt.Printf("Size: %d Val: % #v Address: %s\n", unsafe.Sizeof(fns), fns, &fns)
		bfns := C.GoBytes(unsafe.Pointer(fns), C.int(fns_len))
		for i, v := range ctx.Customs {
			fn := C.sass_make_function(C.CString(v), (*[0]byte)(C.CallMe), nil)
			bfns[i] = byte(*(*C.int)(unsafe.Pointer(fn)))
		}

		fmt.Printf("Size: %d Val: % #v\n", unsafe.Sizeof(bfns), bfns)
		// c := new(CustomList)
		// &c[0] = C.sass_make_function(C.CString("foo($bar,$baz)"), (*[0]byte)(C.CallMe), nil)

		//
		ptr := unsafe.Pointer(fns)
		ptr = unsafe.Pointer(&bfns[0])
		//C.sass_option_set_c_functions(opts, ptr)
		cpfns := (**C.struct_Sass_C_Function_Descriptor)(ptr)

		C.sass_option_set_c_functions(opts, cpfns)
	}

	dc := C.sass_make_data_context(src)
	defer func() {
		C.free(unsafe.Pointer(src))

		C.free(unsafe.Pointer(imgpath))

		C.sass_delete_data_context(dc)
	}()

	C.sass_option_set_precision(opts, prec)

	C.sass_option_set_source_comments(opts, cmt)
	C.sass_data_context_set_options(dc, opts)
	_ = C.sass_compile_data_context(dc)
	cc := C.sass_data_context_get_context(dc)
	cout := C.GoString(C.sass_context_get_output_string(cc))
	io.WriteString(out, cout)

	ctx.Status = int(C.sass_context_get_error_status(cc))
	errS := ctx.ProcessSassError([]byte(C.GoString(C.sass_context_get_error_json(cc))))

	if errS != "" {
		return errors.New(errS)
	}

	return nil
}
