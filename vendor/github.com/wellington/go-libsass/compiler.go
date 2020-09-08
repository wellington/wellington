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

	// Syntax represents the style of code Sass or SCSS
	Syntax() Syntax
}

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
	// FIXME: old context for storing state, use compiler instead
	ctx     *compctx
	dst     io.Writer
	mappath string

	// src is the input stream to compile
	src io.Reader
	// path to the input file
	srcFile string

	// cachebust instructs the compiler to generate new paths
	// preventing browser caching
	cachebust     string
	httpPath      string
	includePaths  []string
	imports       []string
	cmt           bool
	sourceMapRoot string
	// payload is passed around for handlers to have context
	payload context.Context

	// current syntax of the compiler, Sass or SCSS
	syntax Syntax
}

var _ Compiler = &sass{}

func (c *sass) run() error {
	defer func() {
		c.imports = c.ctx.ResolvedImports
	}()

	if len(c.srcFile) > 0 {
		return c.ctx.fileCompile(c.srcFile, c.dst, c.mappath, c.sourceMapRoot)
	}
	return c.ctx.compile(c.dst, c.src)
}

// BuildDir is where CSS is written to disk
func (s *sass) BuildDir() string {
	return s.ctx.BuildDir
}

// CacheBust reveals the current cache busting state
func (c *sass) CacheBust() string {
	return c.cachebust
}

func (s *sass) HTTPPath() string {
	return s.ctx.HTTPPath
}

// ImgBuildDir fetch the image build directory
func (s *sass) ImgBuildDir() string {
	return s.ctx.GenImgDir
}

// ImgDir returns the Image Directory used for locating images
func (c *sass) ImgDir() string {
	return c.ctx.ImageDir
}

// Imports returns the full list of partials used to build the
// output stylesheet.
func (c *sass) Imports() []string {
	return c.imports
}

// FontDir returns the font directory option
func (c *sass) FontDir() string {
	return c.ctx.FontDir
}

// LineComments returns the source comment status
func (c *sass) LineComments() bool {
	return c.cmt
}

func (c *sass) Payload() context.Context {
	return c.ctx.Payload
}

// Run starts transforming S[c|a]ss to CSS
func (c *sass) Run() error {
	return c.run()
}

func (c *sass) Syntax() Syntax {
	return c.syntax
}
