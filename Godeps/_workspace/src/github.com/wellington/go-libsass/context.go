package context

// #include <stdlib.h>
// #include "sass_context.h"
//
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

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
	Headers Headers
	// Has list of compiler included files
	ResolvedImports []string
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

type SassOptions C.struct_Sass_Options

func NewSassOptions() *SassOptions {
	copts := C.sass_make_options()
	return (*SassOptions)(copts)
}

// Init validates options in the struct and returns a Sass Options.
func (ctx *Context) Init(goopts *SassOptions) *C.struct_Sass_Options {
	opts := (*C.struct_Sass_Options)(goopts)
	if ctx.Precision == 0 {
		ctx.Precision = 5
	}
	cmt := C.bool(ctx.Comments)
	imgpath := C.CString(ctx.ImageDir)
	prec := C.int(ctx.Precision)

	defer func() {
		C.free(unsafe.Pointer(imgpath))
		// C.free(unsafe.Pointer(cc))
		// C.sass_delete_data_context(dc)
	}()
	Mixins(ctx)
	ctx.SetHeaders(opts)
	ctx.SetImporter(opts)
	ctx.SetIncludePaths(opts)
	ctx.SetFunc(opts)

	C.sass_option_set_precision(opts, prec)
	C.sass_option_set_source_comments(opts, cmt)
	return opts
}

func (c *Context) SetIncludePaths(opts *C.struct_Sass_Options) {
	for _, inc := range c.IncludePaths {
		C.sass_option_set_include_path(opts, C.CString(inc))
	}
}

func GetImportList(cctx *C.struct_Sass_Context) []string {
	len := int(C.sass_context_get_included_files_size(cctx))
	imps := C.sass_context_get_included_files(cctx)
	list := make([]string, len, len)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(imps)),
		Len:  len, Cap: len,
	}
	goimps := *(*[]*C.char)(unsafe.Pointer(&hdr))
	for i := range goimps {
		list[i] = C.GoString(goimps[i])
	}
	return list
}

func (c *Context) FileCompile(path string, out io.Writer) error {
	defer c.Reset()
	cpath := C.CString(path)
	fc := C.sass_make_file_context(cpath)
	defer C.sass_delete_file_context(fc)
	fcopts := C.sass_file_context_get_options(fc)
	goopts := (*SassOptions)(fcopts)
	opts := c.Init(goopts)
	//os.PathListSeparator
	incs := strings.Join(c.IncludePaths, string(os.PathListSeparator))
	C.sass_option_set_include_path(opts, C.CString(incs))
	C.sass_file_context_set_options(fc, opts)
	cc := C.sass_file_context_get_context(fc)
	compiler := C.sass_make_file_compiler(fc)
	C.sass_compiler_parse(compiler)
	c.ResolvedImports = GetImportList(cc)
	C.sass_compiler_execute(compiler)
	defer C.sass_delete_compiler(compiler)
	cout := C.GoString(C.sass_context_get_output_string(cc))
	io.WriteString(out, cout)
	c.Status = int(C.sass_context_get_error_status(cc))
	errJson := C.sass_context_get_error_json(cc)
	err := c.ProcessSassError([]byte(C.GoString(errJson)))
	if err != nil {
		return err
	}
	if c.error() != "" {
		/*lines := bytes.Split(bs, []byte("\n"))
		var out string
		for i := -7; i < 7; i++ {
			if i+c.Errors.Line >= 0 && i+c.Errors.Line < len(lines) {
				out += fmt.Sprintf("%s\n", string(lines[i+c.Errors.Line]))
			}
		}
		// TODO: this is weird, make something more idiomatic*/
		return errors.New(c.error())
	}

	return nil
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

	options := C.sass_data_context_get_options(dc)
	opts := ctx.Init((*SassOptions)(options))

	// TODO: Manually free options memory without throwing
	// malloc errors
	// defer C.free(unsafe.Pointer(opts))
	C.sass_data_context_set_options(dc, opts)
	cc := C.sass_data_context_get_context(dc)
	compiler := C.sass_make_data_compiler(dc)

	C.sass_compiler_parse(compiler)
	C.sass_compiler_execute(compiler)
	defer C.sass_delete_compiler(compiler)

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
