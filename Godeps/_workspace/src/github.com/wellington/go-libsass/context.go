package context

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/wellington/go-libsass/libs"
)

// Context handles the interactions with libsass.  Context
// exposes libsass options that are available.
type Context struct {
	// Options
	OutputStyle  int
	Precision    int
	Comments     bool
	IncludePaths []string
	// Input directories
	FontDir  string
	ImageDir string
	// Output/build directories
	BuildDir  string
	GenImgDir string

	// HTTP supporting code
	HTTPPath                    string
	In, Src, Out, Map, MainFile string
	Status                      int //libsass status code

	// many error parameters some are unnecessary and should be removed
	libsassError string
	errorString  string
	errors       lErrors
	Errors       SassError

	in  io.Reader
	out io.Writer
	// Place to keep cookies, so Go doesn't garbage collect them before C
	// is done with them
	Cookies []Cookie
	// Imports has the list of Import files currently present
	// in the calling context
	Imports Imports
	Headers Headers
	// Has list of compiler included files
	ResolvedImports []string

	// Attach additional data to a context for use by custom
	// handlers/mixins
	Payload interface{}
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

	return &c
}

// Init validates options in the struct and returns a Sass Options.
func (ctx *Context) Init(goopts libs.SassOptions) libs.SassOptions {
	if ctx.Precision == 0 {
		ctx.Precision = 5
	}

	Mixins(ctx)

	ctx.SetHeaders(goopts)
	ctx.SetImporters(goopts)
	ctx.SetFunc(goopts)
	libs.SetIncludePaths(goopts, ctx.IncludePaths)
	libs.SassOptionSetPrecision(goopts, ctx.Precision)
	libs.SassOptionSetOutputStyle(goopts, ctx.OutputStyle)
	libs.SassOptionSetSourceComments(goopts, ctx.Comments)
	return goopts
}

func (ctx *Context) FileCompile(path string, out io.Writer) error {
	defer ctx.Reset()
	gofc := libs.SassMakeFileContext(path)
	goopts := libs.SassFileContextGetOptions(gofc)
	ctx.Init(goopts)
	//os.PathListSeparator
	incs := strings.Join(ctx.IncludePaths, string(os.PathListSeparator))
	libs.SassOptionSetIncludePath(goopts, incs)
	libs.SassFileContextSetOptions(gofc, goopts)
	gocc := libs.SassFileContextGetContext(gofc)
	gocompiler := libs.SassMakeFileCompiler(gofc)
	libs.SassCompilerParse(gocompiler)
	ctx.ResolvedImports = libs.GetImportList(gocc)
	libs.SassCompilerExecute(gocompiler)
	defer libs.SassDeleteCompiler(gocompiler)

	goout := libs.SassContextGetOutputString(gocc)
	io.WriteString(out, goout)
	ctx.Status = libs.SassContextGetErrorStatus(gocc)
	errJSON := libs.SassContextGetErrorJSON(gocc)
	// Yet another property for storing errors
	ctx.libsassError = errJSON
	err := ctx.ProcessSassError([]byte(errJSON))
	if err != nil {
		return err
	}
	if ctx.error() != "" {
		// TODO: this is weird, make something more idiomatic*/
		return errors.New(ctx.error())
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

	godc := libs.SassMakeDataContext(string(bs))
	goopts := libs.SassDataContextGetOptions(godc)
	ctx.Init(goopts)
	libs.SassDataContextSetOptions(godc, goopts)
	goctx := libs.SassDataContextGetContext(godc)
	gocompiler := libs.SassMakeDataCompiler(godc)
	libs.SassCompilerParse(gocompiler)
	libs.SassCompilerExecute(gocompiler)
	defer libs.SassDeleteCompiler(gocompiler)

	goout := libs.SassContextGetOutputString(goctx)
	io.WriteString(out, goout)

	ctx.Status = libs.SassContextGetErrorStatus(goctx)
	errJSON := libs.SassContextGetErrorJSON(goctx)
	ctx.libsassError = errJSON
	err = ctx.ProcessSassError([]byte(errJSON))

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
