package libsass

import (
	"errors"
	"io"
)

type Compiler interface {
	Run() error
	Imports() []string
	Options(...options) error
}

func OutputStyle(style int) options {
	return func(c *Sass) error {
		c.ctx.OutputStyle = style
		return nil
	}
}

// Precision specifies the number of points beyond the decimal place is
// preserved during math calculations.
func Precision(prec int) options {
	return func(c *Sass) error {
		c.ctx.Precision = prec
		return nil
	}
}

// Comments toggles whether comments should be included in the output
func Comments(b bool) options {
	return func(c *Sass) error {
		c.ctx.Comments = b
		return nil
	}
}

// IncludePaths adds additional directories to search for Sass files
func IncludePaths(includes []string) options {
	return func(c *Sass) error {
		c.includePaths = includes
		c.ctx.IncludePaths = includes
		return nil
	}
}

// HTTPPath prefixes all sprites and generated images with this uri.
// Enabling wellington to serve images when used in HTTP mode
func HTTPPath(u string) options {
	return func(c *Sass) error {
		c.httpPath = u
		c.ctx.HTTPPath = u
		return nil
	}
}

// SourceMap behaves differently depending on compiler used. For
// compile, it will embed sourcemap into the source. For file compile,
// it will include a separate file with the source map.
func SourceMap(b bool) options {
	return func(c *Sass) error {
		c.ctx.includeMap = b
		return nil
	}
}

// FontDir specifies where to find fonts
func FontDir(path string) options {
	return func(c *Sass) error {
		c.ctx.FontDir = path
		return nil
	}
}

// BasePath sets the internal path provided to handlers requiring
// a base path for http calls. This is useful for hosted solutions that
// need to provided absolute paths to assets.
func BasePath(basePath string) options {
	return func(c *Sass) error {
		c.httpPath = basePath
		// FIXME: remove from context
		c.ctx.HTTPPath = basePath
		return nil
	}
}

// Path specifies a file to read instead of using the provided
// io.Reader. This activates file compiling that includes line numbers
// in the resulting output.
func Path(path string) options {
	return func(c *Sass) error {
		c.srcFile = path
		c.ctx.MainFile = path
		return nil
	}
}

// annoying options for handlers to work

func Payload(load interface{}) options {
	return func(c *Sass) error {
		c.ctx.Payload = load
		return nil
	}
}

// ImgBuildDir specifies where to place images
func ImgBuildDir(path string) options {
	return func(c *Sass) error {
		c.ctx.GenImgDir = path
		return nil
	}
}

// ImgDir specifies where to locate images for spriting
func ImgDir(path string) options {
	return func(c *Sass) error {
		c.ctx.ImageDir = path
		return nil
	}
}

// BuildDir only used for spriting, how terrible!
func BuildDir(path string) options {
	return func(c *Sass) error {
		c.ctx.BuildDir = path
		return nil
	}
}

type options func(*Sass) error

func New(dst io.Writer, src io.Reader, opts ...options) (Compiler, error) {

	c := &Sass{
		dst: dst,
		src: src,
		ctx: NewContext(),
	}
	c.ctx.in = src
	c.ctx.out = dst

	return c, c.Options(opts...)
}

// Sass implements compiler interface for Sass and Scss stylesheets. To
// configure the compiler, use the options method.
type Sass struct {
	ctx          *Context
	dst          io.Writer
	src          io.Reader
	srcFile      string
	httpPath     string
	includePaths []string
	imports      []string
	// payload is passed around for handlers to have context
	payload interface{}
}

var _ Compiler = &Sass{}

func (c *Sass) run() error {
	defer func() {
		c.imports = c.ctx.ResolvedImports
	}()

	if len(c.srcFile) > 0 {
		return c.ctx.FileCompile(c.srcFile, c.dst)
	}
	return c.ctx.Compile(c.src, c.dst)
}

func (c *Sass) Options(opts ...options) error {
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// Run starts transforming S[c|a]ss to CSS
func (c *Sass) Run() error {
	return c.run()
}

// Imports returns the full list of partials used to build the output
// stylesheet.
func (c *Sass) Imports() []string {
	return c.imports
}

var ErrNoCompile = errors.New("No compile has occurred")
