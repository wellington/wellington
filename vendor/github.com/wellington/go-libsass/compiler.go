package libsass

import (
	"errors"
	"io"

	"golang.org/x/net/context"
)

var (
	ErrPayloadEmpty          = errors.New("empty payload")
	ErrNoCompile             = errors.New("No compile has occurred")
	_               Pather   = &sass{}
	_               Compiler = &sass{}
)

// Pather describes the file system paths necessary for a project
type Pather interface {
	ImgDir() string
	BuildDir() string
	HTTPPath() string
	ImgBuildDir() string
	FontDir() string
}

// Compiler interface is used to translate input Sass network, filepath,
// or otherwise and transforms it to CSS. The interface includes methods
// for adding imports and specifying build options necessary to do the
// transformation.
//
type Compiler interface {
	// Run does a synchronous build via cgo. It is thread safe, but there is
	// no guarantee that the cgo calls will always be that way.
	Run() error
	// Imports returns the imports used for a compile. This is built
	// at parser time in libsass
	Imports() []string
	// Option allows the configuration of the compiler. The option is
	// unexported to encourage use of preconfigured option functions.
	Option(...option) error

	// CacheBust specifies the cache bust method used by the compiler
	// Available options: ts, sum
	CacheBust() string

	// LineComments specifies whether line comments were inserted into
	// output CSS
	LineComments() bool

	// Payload returns the attached spritewell information attached
	// to the compiler context
	Payload() context.Context
}

// CacheBust append timestamps to static assets to prevent caching
func CacheBust(t string) option {
	return func(c *sass) error {
		if t == "ts" {
			t = "timestamp"
		}
		c.cachebust = t
		return nil
	}
}

// LineComments removes the line by line playby of the Sass compiler
func LineComments(b bool) option {
	return func(c *sass) error {
		c.cmt = b
		return nil
	}
}

// OutputStyle controls the presentation of the CSS available option:
// nested, expanded, compact, compressed
func OutputStyle(style int) option {
	return func(c *sass) error {
		c.ctx.OutputStyle = style
		return nil
	}
}

// Precision specifies the number of points beyond the decimal place is
// preserved during math calculations.
func Precision(prec int) option {
	return func(c *sass) error {
		c.ctx.Precision = prec
		return nil
	}
}

// Comments toggles whether comments should be included in the output
func Comments(b bool) option {
	return func(c *sass) error {
		c.ctx.Comments = b
		return nil
	}
}

// IncludePaths adds additional directories to search for Sass files
func IncludePaths(includes []string) option {
	return func(c *sass) error {
		c.includePaths = includes
		c.ctx.IncludePaths = includes
		return nil
	}
}

// HTTPPath prefixes all sprites and generated images with this uri.
// Enabling wellington to serve images when used in HTTP mode
func HTTPPath(u string) option {
	return func(c *sass) error {
		c.httpPath = u
		c.ctx.HTTPPath = u
		return nil
	}
}

// SourceMap behaves differently depending on compiler used. For
// compile, it will embed sourcemap into the source. For file compile,
// it will include a separate file with the source map.
func SourceMap(b bool) option {
	return func(c *sass) error {
		c.ctx.includeMap = b
		return nil
	}
}

// FontDir specifies where to find fonts
func FontDir(path string) option {
	return func(c *sass) error {
		c.ctx.FontDir = path
		return nil
	}
}

// BasePath sets the internal path provided to handlers requiring
// a base path for http calls. This is useful for hosted solutions that
// need to provided absolute paths to assets.
func BasePath(basePath string) option {
	return func(c *sass) error {
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
	return func(c *sass) error {
		c.srcFile = path
		c.ctx.MainFile = path
		return nil
	}
}

// Payload gives access to sprite and image information for handlers
// to perform spriting functions.
func Payload(load context.Context) option {
	return func(c *sass) error {
		c.ctx.Payload = load
		return nil
	}
}

// ImgBuildDir specifies the destination directory for images
func ImgBuildDir(path string) option {
	return func(c *sass) error {
		c.ctx.GenImgDir = path
		return nil
	}
}

// ImgDir specifies where to locate images for spriting
func ImgDir(path string) option {
	return func(c *sass) error {
		c.ctx.ImageDir = path
		return nil
	}
}

// BuildDir only used for spriting, how terrible!
func BuildDir(path string) option {
	return func(c *sass) error {
		c.ctx.BuildDir = path
		return nil
	}
}

type option func(*sass) error

func New(dst io.Writer, src io.Reader, opts ...option) (Compiler, error) {

	c := &sass{
		dst: dst,
		src: src,
		ctx: newContext(),
	}

	c.ctx.in = src
	c.ctx.out = dst
	c.ctx.compiler = c
	err := c.Option(opts...)

	return c, err
}

// sass implements compiler interface for Sass and Scss stylesheets. To
// configure the compiler, use the option method.
type sass struct {
	ctx     *compctx
	dst     io.Writer
	src     io.Reader
	srcFile string

	cachebust    string
	httpPath     string
	includePaths []string
	imports      []string
	cmt          bool
	// payload is passed around for handlers to have context
	payload context.Context
}

var _ Compiler = &sass{}

func (c *sass) run() error {
	defer func() {
		c.imports = c.ctx.ResolvedImports
	}()

	if len(c.srcFile) > 0 {
		return c.ctx.FileCompile(c.srcFile, c.dst)
	}
	return c.ctx.Compile(c.src, c.dst)
}

func (c *sass) CacheBust() string {
	return c.cachebust
}

func (s *sass) BuildDir() string {
	return s.ctx.BuildDir
}

func (s *sass) HTTPPath() string {
	return s.ctx.HTTPPath
}

func (s *sass) ImgBuildDir() string {
	return s.ctx.GenImgDir
}

// ImgDir returns the Image Directory used for locating images
func (c *sass) ImgDir() string {
	return c.ctx.ImageDir
}

// FontDir returns the font directory option
func (c *sass) FontDir() string {
	return c.ctx.FontDir
}

// Option allows the modifying of internal compiler state
func (c *sass) Option(opts ...option) error {
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *sass) Payload() context.Context {
	return c.ctx.Payload
}

// Run starts transforming S[c|a]ss to CSS
func (c *sass) Run() error {
	return c.run()
}

// Imports returns the full list of partials used to build the output
// stylesheet.
func (c *sass) Imports() []string {
	return c.imports
}

func (c *sass) LineComments() bool {
	return c.cmt
}
