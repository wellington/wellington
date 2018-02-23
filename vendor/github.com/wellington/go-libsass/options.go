package libsass

import "context"

type option func(*sass) error

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

// BasePath sets the internal path provided to handlers requiring
// a base path for http calls. This is useful for hosted solutions
// that need to provided absolute paths to assets.
func BasePath(basePath string) option {
	return func(c *sass) error {
		c.httpPath = basePath
		// FIXME: remove from context
		c.ctx.HTTPPath = basePath
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

// Comments toggles whether comments should be included in the output
func Comments(b bool) option {
	return func(c *sass) error {
		c.ctx.Comments = b
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

// HTTPPath prefixes all sprites and generated images with this uri.
// Enabling wellington to serve images when used in HTTP mode
func HTTPPath(u string) option {
	return func(c *sass) error {
		c.httpPath = u
		c.ctx.HTTPPath = u
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

// ImportsOption specifies configuration for import resolution
func ImportsOption(imports *Imports) option {
	return func(c *sass) error {
		c.ctx.Imports = imports
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

// LineComments removes the line by line playby of the Sass compiler
func LineComments(b bool) option {
	return func(c *sass) error {
		c.cmt = b
		return nil
	}
}

// OutputStyle controls the presentation of the CSS available
// option: nested, expanded, compact, compressed
func OutputStyle(style int) option {
	return func(c *sass) error {
		c.ctx.OutputStyle = style
		return nil
	}
}

// Precision specifies the number of points beyond the decimal
// place is preserved during math calculations.
func Precision(prec int) option {
	return func(c *sass) error {
		c.ctx.Precision = prec
		return nil
	}
}

// SourceMap behaves differently depending on compiler used. For
// compile, it will embed sourcemap into the source. For file
// compile, it will include a separate file with the source map.
func SourceMap(b bool, path, sourceMapRoot string) option {
	return func(c *sass) error {
		c.ctx.includeMap = b
		c.mappath = path
		if len(sourceMapRoot) > 0 {
			c.sourceMapRoot = sourceMapRoot
		}
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

// Syntax lists that available syntaxes for the compiler
type Syntax int

const (
	SCSSSyntax Syntax = iota
	SassSyntax
)

func WithSyntax(mode Syntax) option {
	return func(c *sass) error {
		c.syntax = mode
		return nil
	}
}
