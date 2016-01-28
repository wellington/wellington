package libsass

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"

	"github.com/wellington/go-libsass/libs"
)

// Context handles the interactions with libsass.  Context
// exposes libsass options that are available.
type compctx struct {
	// TODO: hack to give handlers Access to the Compiler
	compiler Compiler

	options    libs.SassOptions
	context    libs.SassContext
	includeMap bool

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
	errorString string
	err         SassError

	in  io.Reader
	out io.Writer

	// Funcs is a slice of Sass function handlers. Registry these globally
	// by calling RegisterHandler
	Funcs *Funcs
	// Imports is a map of overridden imports. When Sass attempts to
	// import a path matching on in this map, it will include the import
	// found in the map before looking for a file on the system.
	Imports *Imports
	// Headers are a map of strings to start any Sass project with. Any
	// header listed here will be present before any other Sass code is
	// compiled.
	Headers *Headers

	// ResolvedImports is the list of files libsass used to compile this
	// Sass sheet.
	ResolvedImports []string

	// Attach additional data to a context for use by custom
	// handlers/mixins
	Payload context.Context
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

func newContext() *compctx {
	c := &compctx{
		Headers: NewHeaders(),
		Imports: NewImports(),
	}
	// FIXME: this doesn't actually work for new options being added
	// to just the compiler
	c.compiler = &sass{ctx: c}
	c.Funcs = NewFuncs(c)
	return c
}

// Init validates options in the struct and returns a Sass Options.
func (ctx *compctx) Init(goopts libs.SassOptions) libs.SassOptions {
	if ctx.Precision == 0 {
		ctx.Precision = 5
	}
	ctx.options = goopts
	ctx.Headers.Bind(goopts)
	ctx.Imports.Bind(goopts)
	ctx.Funcs.Bind(goopts)
	libs.SassOptionSetSourceComments(goopts, ctx.compiler.LineComments())
	libs.SetIncludePaths(goopts, ctx.IncludePaths)
	libs.SassOptionSetPrecision(goopts, ctx.Precision)
	libs.SassOptionSetOutputStyle(goopts, ctx.OutputStyle)
	libs.SassOptionSetSourceComments(goopts, ctx.Comments)
	return goopts
}

func (ctx *compctx) FileCompile(path string, out io.Writer) error {
	defer ctx.Reset()
	gofc := libs.SassMakeFileContext(path)
	goopts := libs.SassFileContextGetOptions(gofc)
	ctx.Init(goopts)
	//os.PathListSeparator
	incs := strings.Join(ctx.IncludePaths, string(os.PathListSeparator))
	libs.SassOptionSetIncludePath(goopts, incs)

	// libs.SassOptionSetSourceMapContents(goopts, true)
	if ctx.includeMap {
		libs.SassOptionSetSourceMapFile(goopts, "boom.map")
	}

	libs.SassFileContextSetOptions(gofc, goopts)
	gocc := libs.SassFileContextGetContext(gofc)
	ctx.context = gocc
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
	err := ctx.ProcessSassError([]byte(errJSON))
	if err != nil {
		return err
	}
	if ctx.Error() != "" {
		// TODO: this is weird, make something more idiomatic*/
		return errors.New(ctx.Error())
	}

	return nil
}

// Compile reads in and writes the libsass compiled result to out.
// Options and custom functions are applied as specified in Context.
func (ctx *compctx) Compile(in io.Reader, out io.Writer) error {

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
	libs.SassOptionSetSourceComments(goopts, true)

	ctx.Init(goopts)
	libs.SassDataContextSetOptions(godc, goopts)
	goctx := libs.SassDataContextGetContext(godc)
	ctx.context = goctx
	gocompiler := libs.SassMakeDataCompiler(godc)
	libs.SassCompilerParse(gocompiler)
	libs.SassCompilerExecute(gocompiler)
	if ctx.includeMap {
		libs.SassOptionSetSourceMapEmbed(goopts, true)
	}
	defer libs.SassDeleteCompiler(gocompiler)

	goout := libs.SassContextGetOutputString(goctx)
	io.WriteString(out, goout)

	ctx.Status = libs.SassContextGetErrorStatus(goctx)
	errJSON := libs.SassContextGetErrorJSON(goctx)
	err = ctx.ProcessSassError([]byte(errJSON))

	if err != nil {
		return err
	}

	if ctx.Error() != "" {
		lines := bytes.Split(bs, []byte("\n"))
		var out string
		for i := -7; i < 7; i++ {
			if i+ctx.err.Line >= 0 && i+ctx.err.Line < len(lines) {
				out += fmt.Sprintf("%s\n", string(lines[i+ctx.err.Line]))
			}
		}
		// TODO: this is weird, make something more idiomatic
		return errors.New(ctx.Error() + "\n" + out)
	}

	return nil
}

// Rel creates relative paths between the build directory where the CSS lives
// and the image directory that is being linked.  This is not compatible
// with generated images like sprites.
func (p *compctx) RelativeImage() string {
	rel, _ := filepath.Rel(p.BuildDir, p.ImageDir)
	return filepath.ToSlash(filepath.Clean(rel))
}
