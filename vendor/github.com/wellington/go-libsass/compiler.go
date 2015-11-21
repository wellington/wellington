package libsass

import (
	"errors"
	"io"
)

type Compiler interface {
	Run() error
	Imports() []string
	Option(...option) error

	ImgDir() string
	CacheBust() bool

	// Context() is deprecated, provided here as a bridge to the future
	Context() *Context
}

// CacheBust append timestamps to static assets to prevent caching
func CacheBust(t bool) option {
	return func(c *Sass) error {
		c.cachebust = t
		return nil
	}
}

// OutputStyle controls the presentation of the CSS available option:
// nested, expanded, compact, compressed
func OutputStyle(style int) option {
	return func(c *Sass) error {
		c.ctx.OutputStyle = style
		return nil
	}
}

// Precision specifies the number of points beyond the decimal place is
// preserved during math calculations.
func Precision(prec int) option {
	return func(c *Sass) error {
		c.ctx.Precision = prec
		return nil
	}
}

// Comments toggles whether comments should be included in the output
func Comments(b bool) option {
	return func(c *Sass) error {
		c.ctx.Comments = b
		return nil
	}
}

// IncludePaths adds additional directories to search for Sass files
func IncludePaths(includes []string) option {
	return func(c *Sass) error {
		c.includePaths = includes
		c.ctx.IncludePaths = includes
		return nil
	}
}

// HTTPPath prefixes all sprites and generated images with this uri.
// Enabling wellington to serve images when used in HTTP mode
func HTTPPath(u string) option {
	return func(c *Sass) error {
		c.httpPath = u
		c.ctx.HTTPPath = u
		return nil
	}
}

// SourceMap behaves differently depending on compiler used. For
// compile, it will embed sourcemap into the source. For file compile,
// it will include a separate file with the source map.
func SourceMap(b bool) option {
	return func(c *Sass) error {
		c.ctx.includeMap = b
		return nil
	}
}

// FontDir specifies where to find fonts
func FontDir(path string) option {
	return func(c *Sass) error {
		c.ctx.FontDir = path
		return nil
	}
}

// BasePath sets the internal path provided to handlers requiring
// a base path for http calls. This is useful for hosted solutions that
// need to provided absolute paths to assets.
func BasePath(basePath string) option {
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
func Path(path string) option {
	return func(c *Sass) error {
		c.srcFile = path
		c.ctx.MainFile = path
		return nil
	}
}

// annoying option for handlers to work

func Payload(load interface{}) option {
	return func(c *Sass) error {
		c.ctx.Payload = load
		return nil
	}
}

// ImgBuildDir specifies where to place images
func ImgBuildDir(path string) option {
	return func(c *Sass) error {
		c.ctx.GenImgDir = path
		return nil
	}
}

// ImgDir specifies where to locate images for spriting
func ImgDir(path string) option {
	return func(c *Sass) error {
		c.ctx.ImageDir = path
		return nil
	}
}

// BuildDir only used for spriting, how terrible!
func BuildDir(path string) option {
	return func(c *Sass) error {
		c.ctx.BuildDir = path
		return nil
	}
}

type option func(*Sass) error

func New(dst io.Writer, src io.Reader, opts ...option) (Compiler, error) {

	c := &Sass{
		dst: dst,
		src: src,
		ctx: NewContext(),
	}

	c.ctx.in = src
	c.ctx.out = dst
	c.ctx.compiler = c
	err := c.Option(opts...)

	return c, err
}

// Sass implements compiler interface for Sass and Scss stylesheets. To
// configure the compiler, use the option method.
type Sass struct {
	ctx     *Context
	dst     io.Writer
	src     io.Reader
	srcFile string

	cachebust    bool
	httpPath     string
	includePaths []string
	imports      []string
	// payload is passed around for handlers to have context
	payload interface{}
}

var _ Compiler = &Sass{}

func (c *Sass) Context() *Context {
	return c.ctx
}

func (c *Sass) run() error {
	defer func() {
		c.imports = c.ctx.ResolvedImports
	}()

	if len(c.srcFile) > 0 {
		return c.ctx.FileCompile(c.srcFile, c.dst)
	}
	return c.ctx.Compile(c.src, c.dst)
}

func (c *Sass) CacheBust() bool {
	return c.cachebust
}

// ImgDir returns the Image Directory used for locating images
func (c *Sass) ImgDir() string {
	return c.ctx.ImageDir
}

// Option allows the modifying of internal compiler state
func (c *Sass) Option(opts ...option) error {
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
