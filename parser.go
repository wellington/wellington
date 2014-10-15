package sprite_sass

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"sort"
)

/* Example sprite-map output:
$sprites: ($rel: "");

$sprites: map_merge($sprites, (
  139: (
    width: 139,
    height: 89,
    x: 0,
    y: 20,
    url: './image.png'
  )));

$sprites: map_merge($sprites,(140: (
    width: 140,
    height: 89,
    x: 0,
    y: 20,
    url: './image.png'
  )));
*/

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

type Replace struct {
	Start, End int
	Value      []byte
}

type Parser struct {
	Idx, shift           int
	Chop                 []Replace
	Pwd, Input, MainFile string
	SassDir, BuildDir,
	GenImgDir string
	StaticDir           string
	ProjDir             string
	ImageDir            string
	Includes            []string
	Items               []Item
	Output              []byte
	Line                map[int]string
	LineKeys            []int
	InlineImgs, Sprites map[string]ImageList
	Vars                map[string]string
}

func NewParser() *Parser {
	return &Parser{}
}

// Parser reads the tokens from the lexer and performs
// conversions and/or substitutions for sprite*() calls.
//
// Parser creates a map of all variables and sprites
// (created via sprite-map calls).
func (p *Parser) Start(in io.Reader, pkgdir string) ([]byte, error) {
	p.Vars = make(map[string]string)
	p.Sprites = make(map[string]ImageList)
	p.InlineImgs = make(map[string]ImageList)
	p.Line = make(map[int]string)

	// Setup paths
	if p.MainFile == "" {
		p.MainFile = "string"
	}
	if p.BuildDir == "" {
		p.BuildDir = pkgdir
	}
	if p.SassDir == "" {
		p.SassDir = pkgdir
	}
	if p.StaticDir == "" {
		p.StaticDir = pkgdir
	}
	if p.ImageDir == "" {
		p.ImageDir = p.StaticDir
	}
	if p.GenImgDir == "" {
		p.GenImgDir = p.BuildDir
	}
	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	buf.ReadFrom(in)

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
	// This call will have valid token positions
	items, input, err = p.GetItems(pkgdir, p.MainFile, input)

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
	p.Parse(p.Items)
	p.Output = []byte(p.Input)
	// Perform substitutions
	p.Replace()
	rel := []byte(fmt.Sprintf(`$rel: "%s"; /* $rel: %s */%s`,
		p.Rel(), p.Rel(), "\n"))
	return append(rel, p.Output...), nil
}

func (p *Parser) Rel() string {
	rel, _ := filepath.Rel(p.BuildDir, p.ImageDir)
	return filepath.Clean(rel)
}

// LookupFile translates line positions into line number
// and file it belongs to
func (p *Parser) LookupFile(pos int) string {
	pos = pos - 1
	for i, n := range p.LineKeys {
		if n > pos {
			if i == 0 {
				return fmt.Sprintf("%s:%d", p.MainFile, pos+1)
			}
			hit := p.LineKeys[i-1]
			return fmt.Sprintf("%s:%d", p.Line[hit], pos-p.LineKeys[i-1])
		}
	}
	return "mainfile?" + p.MainFile
}

// Find Paren that matches the current (
// func RParen(items []Item) (int, int) {
// 	if len(items) == 0 {
// 		return 0, 0
// 	}
// 	if items[0].Type != LPAREN {
// 		panic("Expected: ( was: " + items[0].Value)
// 	}
// 	pos := 1
// 	match := 1
// 	nest := false
// 	nestPos := 0

// 	for match != 0 && pos < len(items) {
// 		switch items[pos].Type {
// 		case LPAREN:
// 			match++
// 		case RPAREN:
// 			match--
// 		}
// 		if match > 1 {
// 			if !nest {
// 				nestPos = pos
// 			}
// 			// Nested command must be resolved
// 			nest = true
// 		}
// 		pos++
// 	}

// 	return pos, nestPos
// }

// func RBracket(items []Item, pos int) (int, int) {
// 	if items[pos].Type != LBRACKET && items[pos].Type != INTP {
// 		panic("Expected: { was: " + items[0].Value)
// 	}

// 	// Move to next item and set match to 1
// 	pos++
// 	match := 1
// 	nest := false
// 	nestPos := 0
// 	for match != 0 && pos < len(items) {
// 		switch items[pos].Type {
// 		case LBRACKET, INTP:
// 			match++
// 		case RBRACKET:
// 			match--
// 		}
// 		if match > 1 {
// 			if !nest {
// 				nestPos = pos
// 			}
// 			// Nested command must be resolved
// 			nest = true
// 		}
// 		pos++
// 	}
// 	return pos, nestPos
// }

func (p *Parser) Parse(items []Item) []byte {
	var (
		out []byte
		eoc int
	)
	_ = eoc
	if len(items) == 0 {
		return []byte("")
	}
	j := 1
	item := items[0]
	switch item.Type {
	case VAR:
		if items[1].Value != ":" {
			log.Fatal(": expected after variable declaration")
		}
		for j < len(items) && items[j].Type != SEMIC {
			j++
		}
		if items[2].Type != CMDVAR {
			// Hackery for empty sass maps
			val := string(p.Parse(items[2:j]))
			// TODO: $var: $anothervar doesnt work
			// setting other things like $var: darken(#123, 10%)
			if val != "()" && val != "" {
				// fmt.Println("SETTING", item, val)
				p.Vars[item.String()] = val
			}
		} else if items[2].Value == "sprite-map" {
			// Special parsing of sprite-maps
			imgs := ImageList{
				ImageDir:  p.ImageDir,
				BuildDir:  p.BuildDir,
				GenImgDir: p.GenImgDir,
				Vertical:  true,
			}
			name := fmt.Sprintf("%s", items[0])
			glob := fmt.Sprintf("%s", items[4])
			imgs.Decode(glob)
			imgs.Combine()
			p.Sprites[name] = imgs
			//TODO: Generate filename
			p.Mark(items[2].Pos,
				items[j].Pos+len(items[j].Value), imgs.Map(name))
			_, err := imgs.Export()
			if err != nil {
				log.Printf("Failed to save sprite: %s", name)
				log.Println(err)
			}
		}
	}

	return append(out, p.Parse(items[j:])...)
}

// Import recursively resolves all imports.  It lexes the input
// adding the tokens to the Parser object.
// TODO: Convert this to byte slice in/out
func (p *Parser) GetItems(pwd, filename, input string) ([]Item, string, error) {

	var (
		status    []Item
		importing bool
		output    []byte
		pos       int
		last      *Item
		lastname  string
		lineCount int
	)

	lex := New(func(lex *Lexer) StateFn {
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
		case ItemEOF:
			output = append(output, input[pos:]...)
			return status, string(output), nil
		case IMPORT:
			output = append(output, input[pos:item.Pos]...)
			last = item
			importing = true
		case INCLUDE, CMT:
			output = append(output, input[pos:item.Pos]...)
			pos = item.Pos
			status = append(status, *item)
		default:
			if importing {
				lastname = filename
				filename = fmt.Sprintf("%s", *item)
				for _, nl := range output {
					if nl == '\n' {
						lineCount++
					}
				}
				p.Line[lineCount] = filename
				pwd, contents, err := p.ImportPath(pwd, filename)
				if err != nil {
					return nil, "", err
				}
				//Eat the semicolon
				item := lex.Next()
				pos = item.Pos + len(item.Value)
				if item.Type != SEMIC {
					log.Println("@import statement must be followed by ;")
				}

				moreTokens, moreOutput, err := p.GetItems(
					pwd,
					filename,
					contents)
				// If importing was successful, each token must be moved
				// forward by the position of the @import call that made
				// it available.
				for i, _ := range moreTokens {
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
				p.Line[lineCount+1] = filename
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
