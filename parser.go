package wellington

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"path/filepath"
	"sort"

	"github.com/wellington/wellington/context"
	// TODO: Remove dot imports
	"github.com/wellington/wellington/lexer"
	"github.com/wellington/wellington/token"
)

var weAreNeverGettingBackTogether = []byte(`@mixin sprite-dimensions($map, $name) {
  $file: sprite-file($map, $name);
  height: image-height($file);
  width: image-width($file);
}
`)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

// Replace holds token values for replacing source input with parsed input.
// DEPRECATED
type Replace struct {
	Start, End int
	Value      []byte
}

// Parser represents a parser engine that returns parsed and imported code
// from the input useful for doing text manipulation before passing to libsass.
type Parser struct {
	Idx, shift           int
	Chop                 []Replace
	Pwd, Input, MainFile string
	SassDir, BuildDir,

	ProjDir string
	Imports    context.Imports
	ImageDir   string
	Includes   []string
	Items      []lexer.Item
	Output     []byte
	Line       map[int]string
	LineKeys   []int
	PartialMap *SafePartialMap
}

// NewParser returns a pointer to a Parser object.
func NewParser() *Parser {
	p := &Parser{PartialMap: NewPartialMap()}
	p.Imports.Init()
	return p
}

// Start reads the tokens from the lexer and performs
// conversions and/or substitutions for sprite*() calls.
//
// Start creates a map of all variables and sprites
// (created via sprite-map calls).
// TODO: Remove pkgdir, it can be put on Parser
func (p *Parser) Start(r io.Reader, pkgdir string) ([]byte, error) {

	if r == nil {
		return []byte{}, errors.New("input is empty")
	}

	var in io.Reader
	var err error

	in, err = ToScssReader(r)

	if err != nil {
		return nil, err
	}

	p.Line = make(map[int]string)

	// Setup paths
	if p.MainFile == "" {
		p.MainFile = "stdin"
	}
	if p.BuildDir == "" {
		p.BuildDir = pkgdir
	}
	if p.SassDir == "" {
		p.SassDir = pkgdir
	}
	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	if in == nil {
		return []byte{}, fmt.Errorf("input is empty")
	}
	_, err = buf.ReadFrom(in)
	if err != nil {
		return []byte{}, err
	}

	// This pass resolves all the imports, but positions will
	// be off due to @import calls
	items, input, err := p.GetItems(pkgdir, p.MainFile, string(buf.Bytes()))
	if err != nil {
		return []byte(""), err
	}
	for i := range p.Line {
		p.LineKeys = append(p.LineKeys, i)
	}
	sort.Ints(p.LineKeys)
	// Try removing this and see if it works
	// This call will have valid token positions
	// items, input, err = p.GetItems(pkgdir, p.MainFile, input)

	p.Input = input
	p.Items = items
	if err != nil {
		panic(err)
	}
	// DEBUG
	// for _, item := range p.Items {
	// 	fmt.Printf("%s %s\n", item.Type, item)
	// }
	// Process sprite calls and gen

	// Send original byte slice
	p.Output = buf.Bytes() //[]byte(p.Input)
	// Perform substitutions
	// p.Replace()
	// rel := []byte(fmt.Sprintf(`$rel: "%s";%s`,
	//   p.Rel(), "\n"))

	// Code that we will never support, ever

	return append(weAreNeverGettingBackTogether, p.Output...), nil
}

// LookupFile translates line positions into line number
// and file it belongs to
func (p *Parser) LookupFile(position int) string {
	// Shift to 0 index
	pos := position - 1
	// Adjust for shift from preamble
	shift := bytes.Count(weAreNeverGettingBackTogether, []byte{'\n'})
	pos = pos - shift
	if pos < 0 {
		return "mixin"
	}
	for i, n := range p.LineKeys {
		if n > pos {
			if i == 0 {
				// Return 1 index line numbers
				return fmt.Sprintf("%s:%d", p.Line[i], pos+1)
			}
			hit := p.LineKeys[i-1]
			filename := p.Line[hit]
			// Catch for mainimport errors
			if filename == "string" {
				filename = p.MainFile
			}
			return fmt.Sprintf("%s:%d", filename, pos-hit+1)
		}
	}
	// Either this is invalid or outside of all imports, assume it's valid
	return fmt.Sprintf("%s:%d", p.MainFile, pos-p.LineKeys[len(p.LineKeys)-1]+1)
}

// GetItems recursively resolves all imports.  It lexes the input
// adding the tokens to the Parser object.
// TODO: Convert this to byte slice in/out
func (p *Parser) GetItems(pwd, filename, input string) ([]lexer.Item, string, error) {

	var (
		status    []lexer.Item
		importing bool
		output    []byte
		pos       int
		last      *lexer.Item
		lastname  string
		lineCount int
	)

	lex := lexer.New(func(lex *lexer.Lexer) lexer.StateFn {
		return lex.Action()
	}, input)

	for {
		item := lex.Next()
		err := item.Error()
		//fmt.Println(item.Type, item.Value)
		if err != nil {
			return nil, string(output),
				fmt.Errorf("Error: %v (pos %d)", err, item.Pos)
		}
		switch item.Type {
		case token.ItemEOF:
			if filename == p.MainFile {
				p.Line[lineCount+bytes.Count([]byte(input[pos:]), []byte("\n"))] = filename
			}
			output = append(output, input[pos:]...)
			return status, string(output), nil
		case token.IMPORT:
			output = append(output, input[pos:item.Pos]...)
			last = item
			importing = true
		case token.INCLUDE, token.CMT:
			output = append(output, input[pos:item.Pos]...)
			pos = item.Pos
			status = append(status, *item)
		default:
			if importing {
				lastname = filename
				// Found import, mark parent's current position
				p.Line[lineCount] = filename
				filename = fmt.Sprintf("%s", *item)
				for _, nl := range output {
					if nl == '\n' {
						lineCount++
					}
				}
				p.Line[lineCount] = filename
				pwd, contents, err := p.ImportPath(pwd, filename)
				// FIXME: hack for top level file
				ln := lastname
				if path.IsAbs(ln) {
					ln = "stdin"
				}
				p.Imports.Add(ln,
					filename, contents)

				if err != nil {
					return nil, "", err
				}

				//Eat the semicolon
				item := lex.Next()
				if item.Type != token.SEMIC {
					log.Printf("@import in %s:%d must be followed by ;\n", filename, lineCount)
					log.Printf("        ~~~> @import %s", filename)
				}
				// Set position to token after
				// FIXME: Hack to delete newline, hopefully this doesn't break stuff
				// then readd it to the linecount
				pos = item.Pos + len(item.Value)
				moreTokens, moreOutput, err := p.GetItems(
					pwd,
					filename,
					string(contents))
				// If importing was successful, each token must be moved
				// forward by the position of the @import call that made
				// it available.
				for i := range moreTokens {
					moreTokens[i].Pos += last.Pos
				}

				if err != nil {
					return nil, "", err
				}
				for _, nl := range moreOutput {
					if nl == '\n' {
						lineCount++
					}
				}
				filename = lastname

				output = append(output, moreOutput...)
				status = append(status, moreTokens...)
				importing = false
			} else {
				output = append(output, input[pos:item.Pos]...)
				pos = item.Pos
				status = append(status, *item)
			}
		}
	}

}

// StartParser accepts build arguments
// TODO: Should this be called StartParser or NewParser?
// TODO: Should this function create the partialMap or is this
// the right way to inject one?
func StartParser(ctx *context.Context, in io.Reader, out io.Writer, partialMap *SafePartialMap) (*Parser, error) {
	// Run the sprite_sass parser prior to passing to libsass
	parser := NewParser()

	parser.ImageDir = ctx.ImageDir
	parser.Includes = ctx.IncludePaths
	parser.BuildDir = ctx.BuildDir
	parser.MainFile = ctx.MainFile
	parser.Imports = ctx.Imports

	// Save reference to parser in context
	bs, err := parser.Start(in, filepath.Dir(ctx.MainFile))
	if err != nil {
		return parser, err
	}
	out.Write(bs)
	return parser, err
}
