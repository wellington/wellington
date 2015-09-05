package wellington

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	libsass "github.com/wellington/go-libsass"
	// TODO: Remove dot imports
)

func init() {
	//log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
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
	Imports    libsass.Imports
	ImageDir   string
	Includes   []string
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

	// Send original byte slice
	p.Output = buf.Bytes() //[]byte(p.Input)

	return p.Output, nil
}

// StartParser accepts build arguments
// TODO: Should this be called StartParser or NewParser?
// TODO: Should this function create the partialMap or is this
// the right way to inject one?
func StartParser(ctx *libsass.Context, in io.Reader, out io.Writer, partialMap *SafePartialMap) (*Parser, error) {
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
