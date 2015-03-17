package context

// #include <stdlib.h>
// #include "sass_context.h"
//
// extern union Sass_Value* GoBridge( union Sass_Value* s_args, void* cookie);
// union Sass_Value* CallSassFunction( union Sass_Value* s_args, void* cookie ) {
//     return GoBridge(s_args, cookie);
// }
//
//
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/wellington/spritewell"

	"unsafe"
)

// Context handles the interactions with libsass.  Context
// exposes libsass options that are available.
type Context struct {
	//Parser                        Parser
	OutputStyle  int
	Precision    int
	Comments     bool
	IncludePaths []string

	// Input directories
	FontDir, ImageDir string
	// Output/build directories
	BuildDir, GenImgDir string
	// HTTP supporting code
	HTTPPath                    string
	In, Src, Out, Map, MainFile string
	Status                      int
	errorString                 string
	errors                      lErrors

	in     io.Reader
	out    io.Writer
	Errors SassError
	// Place to keep cookies, so Go doesn't garbage collect them before C
	// is done with them
	Cookies []Cookie
	// Imports has the list of Import files currently present
	// in the calling context
	Imports Imports
	// Used for callbacks to retrieve sprite information, etc.
	Imgs, Sprites spritewell.SafeImageMap
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

func NewContext() *Context {
	c := Context{}

	// Initiailize image map(s)
	c.Sprites = spritewell.SafeImageMap{
		M: make(map[string]spritewell.ImageList)}
	c.Imgs = spritewell.SafeImageMap{
		M: make(map[string]spritewell.ImageList)}
	c.Imports.Init()

	return &c
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
			h.Sign, h.Fn, ctx,
		}
	}
	ctx.Cookies = cookies
	size := C.size_t(len(ctx.Cookies))
	fns := C.sass_make_function_list(size)
	signatures := make([]string, len(ctx.Cookies))
	// Send cookies to libsass
	// Create a slice that's backed by a C array
	length := len(ctx.Cookies) + 1
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(fns)),
		Len:  length, Cap: length,
	}

	gofns := *(*[]C.Sass_C_Function_Callback)(unsafe.Pointer(&hdr))
	for i, v := range ctx.Cookies {
		signatures[i] = ctx.Cookies[i].Sign
		_ = v
		cg := C.CString(signatures[i])
		_ = cg

		fn := C.sass_make_function(
			// sass signature
			C.CString(v.Sign),
			// C bridge
			C.Sass_C_Function(C.CallSassFunction),
			// Only pass reference to global array, so
			// GC won't clean it up.
			unsafe.Pointer(&ctx.Cookies[i]))

		gofns[i] = fn
	}

	ctx.SetImporter(opts)

	C.sass_option_set_c_functions(opts, (C.Sass_C_Function_List)(unsafe.Pointer(&gofns[0])))
	C.sass_option_set_precision(opts, prec)
	C.sass_option_set_source_comments(opts, cmt)
	return opts
}

// Compile reads in and writes the libsass compiled result to out.
// Options and custom functions are applied as specified in Context.
func (ctx *Context) Compile(in io.Reader, out io.Writer) error {

	defer ctx.Reset()
	bs, err := ioutil.ReadAll(in)

	if err != nil {
		return err
	}
	if len(bs) == 0 {
		return errors.New("No input provided")
	}
	src := C.CString(string(bs))

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
	errJSON := C.sass_context_get_error_json(cc)
	err = ctx.ProcessSassError([]byte(C.GoString(errJSON)))

	if err != nil {
		return err
	}

	if ctx.error() != "" {
		lines := bytes.Split(bs, []byte("\n"))
		var out string
		for i := -7; i < 7; i++ {
			if i+ctx.Errors.Line >= 0 && i+ctx.Errors.Line < len(lines) {
				out += fmt.Sprintf("%s\n", string(lines[i+ctx.Errors.Line]))
			}
		}
		// TODO: this is weird, make something more idiomatic
		return errors.New(ctx.error() + "\n" + out)
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
