package context

// #include <stdlib.h>
// #include "sass_context.h"
//
// extern union Sass_Value* goBridge( union Sass_Value* s_args, void* cookie);
// union Sass_Value* CallSassFunction( union Sass_Value* s_args, void* cookie ) {
//     return goBridge(s_args, cookie);
// }
import "C"

import (
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/drewwells/spritewell"

	"unsafe"
)

// Cookie is used for passing context information to libsass.  Cookie is
// passed to custom handlers when libsass executes them through the go bridge.
type Cookie struct {
	sign string
	fn   SassCallback
	ctx  *Context
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

	in     io.Reader
	out    io.Writer
	Errors SassError
	// Place to keep cookies, so Go doesn't garbage collect them before C
	// is done with them
	Cookies []Cookie

	// Used for callbacks to retrieve sprite information, etc.
	InlineImgs, Sprites map[string]spritewell.ImageList

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
	cookies := make([]Cookie, len(handlers)+len(ctx.Cookies))
	// Append registered handlers to cookie array
	for i, h := range handlers {
		cookies[i] = Cookie{
			h.sign, h.callback, ctx,
		}
	}
	for i, h := range ctx.Cookies {
		cookies[i+len(handlers)] = Cookie{
			h.sign, h.fn, ctx,
		}
	}

	size := C.size_t(len(ctx.Cookies) + 1)
	fns := C.sass_make_function_list(size)
	// Send cookies to libsass
	if len(cookies) > 0 {
		for i, v := range ctx.Cookies {
			fn := C.sass_make_function(
				// sass signature
				C.CString(v.sign),
				// C bridge
				C.Sass_C_Function(C.CallSassFunction),
				// Only pass reference to global array, so
				// GC won't clean it up.
				unsafe.Pointer(&ctx.Cookies[i]))
			C.sass_set_function(&fns, fn, C.int(i))
		}
	}

	C.sass_option_set_c_functions(opts, fns)

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
	errS, err := ctx.ProcessSassError([]byte(C.GoString(errJson)))

	if err != nil {
		return err
	}

	if errS != "" {
		return errors.New(errS)
	}

	return nil
}

// Rel creates relative paths between the build directory where the CSS lives
// and the image directory that is being linked.  This is not compatible
// with generated images like sprites.
func (p *Context) RelativeImage() string {
	rel, _ := filepath.Rel(p.BuildDir, p.ImageDir)
	return filepath.Clean(rel)
}
