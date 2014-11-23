package context

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
	"io/ioutil"

	"unsafe"

	. "github.com/drewwells/spritewell"
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
	Sprites                       []ImageList
	Status                        int
	errorString                   string
	errors                        lErrors

	in     io.Reader
	out    io.Writer
	Errors SassError
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

// libsass for generating the resulting css file.
func (ctx *Context) Compile(in io.Reader, out io.Writer, s string) error {
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
