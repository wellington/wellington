package sprite_sass

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

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
	Pwd, Input, ImageDir string
	Includes             []string
	Items                []Item
	Output               []byte
	Sprites              map[string]ImageList
	NewVars, Vars        map[string]string
}

func NewParser() *Parser {
	return &Parser{}
}

// Parser reads the tokens from the lexer and performs
// conversions and/or substitutions for sprite*() calls.
//
// Parser creates a map of all variables and sprites
// (created via sprite-map calls).
func (p *Parser) Start(in io.Reader, pkgdir string) []byte {
	p.Vars = make(map[string]string)
	p.NewVars = make(map[string]string)
	p.Sprites = make(map[string]ImageList)

	if p.ImageDir == "" {
		p.ImageDir = pkgdir
	}
	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	buf.ReadFrom(in)

	// This pass resolves all the imports, but positions will
	// be off due to @import calls
	tokens, input, err := p.start(pkgdir, string(buf.Bytes()))
	// This call will have valid token positions
	tokens, input, err = p.start(pkgdir, input)

	p.Input = input
	p.Items = tokens
	if err != nil {
		panic(err)
	}
	// DEBUG
	// for _, item := range p.Items {
	// 	fmt.Printf("%s %s\n", item.Type, item)
	// }
	p.Parse(p.Items)

	p.Output = []byte(p.Input)
	p.Replace()
	// fmt.Printf("out: % #v\n", p.Sprites)
	return p.Output
}

// Find Paren that matches the current (
func RParen(items []Item) (int, int) {
	if items[0].Type != LPAREN {
		panic("Expected: ( was: " + items[0].Value)
	}
	pos := 1
	match := 1
	nest := false
	nestPos := 0
	for match != 0 {
		switch items[pos].Type {
		case LPAREN:
			match++
		case RPAREN:
			match--
		}
		if match > 1 {
			if !nest {
				nestPos = pos
			}
			// Nested command must be resolved
			nest = true
		}
		pos++
	}

	return pos, nestPos
}

func RBracket(items []Item, pos int) (int, int) {
	if items[pos].Type != LBRACKET {
		panic("Expected: { was: " + items[0].Value)
	}

	// Move to next item and set match to 1
	pos++
	match := 1
	nest := false
	nestPos := 0
	for match != 0 && pos < len(items) {
		switch items[pos].Type {
		case LBRACKET, INT:
			match++
		case RBRACKET:
			match--
		}
		if match > 1 {
			if !nest {
				nestPos = pos
			}
			// Nested command must be resolved
			nest = true
		}
		pos++
	}
	return pos, nestPos
}

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
		for items[j].Type != SEMIC {
			j++
		}
		if items[1].Type != CMDVAR {
			// Hackery for empty sass maps
			val := string(p.Parse(items[1:j]))
			if val != "()" {
				p.NewVars[item.String()] = val
			}
		} else {
			// Special parsing of sprite-maps
			p.Mark(items[0].Pos,
				items[j].Pos+len(items[j].Value), "")
			imgs := ImageList{}

			name := fmt.Sprintf("%s", items[0])
			glob := fmt.Sprintf("%s", items[3])
			imgs.Decode(p.ImageDir + "/" + glob)
			imgs.Vertical = true
			imgs.Combine()
			p.Sprites[name] = imgs
			//TODO: Generate filename
			//imgs.Export("generated.png")
		}
	case SUB:
		/*for items[j].Type != SEMIC {
			j++
			if j >= len(items) {
				fmt.Println(items)
				panic(fmt.Sprintf("Did not find ; for %s\n", item))
			}
		}*/
		val, ok := p.NewVars[item.Value]
		// Do not replace if nothing was found
		if !ok {
			val = item.Value
		}
		p.Mark(item.Pos, item.Pos+len(item.Value), val)
	case CMD:
		for j < len(items) && items[j].Type != SEMIC {
			j++
		}
		out, eoc = p.Command(items[0:j])
	case TEXT:
		out = append(out, item.Value...)
	case MIXIN, FUNC, IF, ELSE, EACH:
		// Ignore the entire mixin and move to the next line
		lpos := 0
		for {
			if items[lpos].Type == LBRACKET {
				break
			}
			lpos++
		}
		pos, _ := RBracket(items, lpos)
		for i := 0; i < pos; i++ {
			out = append(out, items[i].Value...)
		}
		// fmt.Println(">>", item.Type, items[lpos:pos], "<<")
		j = pos
	default:
		out = append(out, item.Value...)
	}

	return append(out, p.Parse(items[j:])...)
}

// Passed sass-command( args...)
func (p *Parser) Command(items []Item) ([]byte, int) {

	i := 0
	_ = i
	cmd := items[0]
	item := items[0]
	repl := ""
	eoc, nPos := RParen(items[1:])
	// Determine our offset from the source items
	if false && nPos != 0 {
		rightPos, _ := RParen(items[nPos:])
		p.Command(items[nPos:rightPos])
	}

	switch cmd.Value {
	case "sprite":
		//Capture sprite
		sprite := p.Sprites[fmt.Sprintf("%s", items[2])]
		//Capture filename
		name := fmt.Sprintf("%s", items[3])
		repl = sprite.CSS(name)
		p.Mark(items[0].Pos, items[4].Pos+len(items[4].Value), repl)
	case "sprite-height":
		sprite := p.Sprites[fmt.Sprintf("%s", items[2])]
		repl = fmt.Sprintf("%dpx",
			sprite.ImageHeight(items[3].String()))
		// Walk forward to file name
		p.Mark(cmd.Pos, items[eoc].Pos+len(items[eoc].Value), repl)
	case "sprite-width":
		sprite := p.Sprites[fmt.Sprintf("%s", items[2])]
		repl = fmt.Sprintf("%dpx",
			sprite.ImageWidth(items[3].String()))
		// Walk forward to file name
		p.Idx++
		p.Mark(cmd.Pos, items[eoc].Pos+len(items[eoc].Value), repl)
	case "asprite-dimensions":
		sprite := p.Sprites[fmt.Sprintf("%s", item)]
		repl = sprite.Dimensions(items[p.Idx+1].String())
		// Walk forward to file name
		p.Idx++
		p.Mark(items[p.Idx-4].Pos, items[p.Idx+3].Pos, repl)
		items = append(items[:p.Idx-4], items[p.Idx:]...)
		p.Idx = p.Idx - 4

	default:
		//items[p.Idx].Value = p.Vars[item.Value]
	}

	return []byte(""), eoc
}

// Mixin processes tokens in the format @include mixin(args...)
// and perform requested actions.
func (p *Parser) Mixin() {

	// Commands always end in ); else panic
	start := p.Idx
	cmd := p.Items[start]

	var file Item
	for {
		cur := p.Items[p.Idx]
		if cur.Type == RPAREN {
			p.Idx++
			if p.Items[p.Idx].Type != SEMIC {
				f, l := p.Items[start].Pos, p.Items[p.Idx].Pos
				log.Fatal("Commands must end with semicolon",
					fmt.Sprintf(p.Input[f:l]))
			} else {
				break
			}
		} else if cur.Type == FILE {
			file = cur
		}
		p.Idx++
	}
	if file.Type != FILE {
		log.Fatal("File for command was not found")
	}
	if cmd.Value == "inline-image" {
		img := ImageList{}
		img.Decode(p.ImageDir + "/" + file.Value)
		repl := img.Inline()
		p.Idx-- // Preserve the final semic
		p.Mark(cmd.Pos-1, p.Items[p.Idx].Pos+1, repl)
	}
	fmt.Println("Mixin", cmd.Value)
}

// Processes file which usually mean cutting some of the input
// text.
func (p *Parser) File(cmd string, start, end int) int {
	first := p.Items[start]
	item := p.Items[end]
	i := end
	if cmd == "sprite-map" {
		// Find the next semicolon and remove it
		for ; p.Items[i].Type != RPAREN; i++ {
		}
		if p.Items[i+1].Type != SEMIC {
			panic("Statements must end in semicolon")
		}
		i++
		// Verify that the statement ends with semicolon
		interest := p.Items[i]
		// Mark the entire line plus semicolon for deletion
		p.Mark(first.Pos, interest.Pos+1, "")
		imgs := ImageList{}
		glob := fmt.Sprintf("%s", item)
		name := fmt.Sprintf("%s", p.Items[start])
		imgs.Decode(p.ImageDir + "/" + glob)
		imgs.Vertical = true
		imgs.Combine()
		p.Sprites[name] = imgs
		//TODO: Generate filename
		//imgs.Export("generated.png")
	}
	return i + 1
}

// func process(in string, items []Item, pos int) []byte {

// 	var out []byte
// 	l := len(items)

// 	if pos >= len(in) {
// 		return []byte("")
// 	}

// 	// TODO: There's an error where items[1] has an invalid
// 	// position.
// 	if l > 1 && items[1].Pos > items[0].Pos {
// 		out = append(out, in[items[0].Pos:items[1].Pos]...)
// 		out = append(out, process(in, items[1:], pos)...)
// 	} else {
// 		out = append(out, in[items[0].Pos:]...)
// 	}

// 	return out
// }

// start recursively resolves all imports.  It lexes the input
// adding the tokens to the Parser object.
// TODO: Convert this to byte slice in/out
func (p *Parser) start(pwd, input string) ([]Item, string, error) {

	var (
		status    []Item
		importing bool
		output    []byte
		pos       int
		last      *Item
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

				pwd, contents, err := p.ImportPath(pwd, fmt.Sprintf("%s", *item))
				if err != nil {
					log.Fatal(err)
				}
				//Eat the semicolon
				item := lex.Next()
				pos = item.Pos + len(item.Value)
				if item.Type != SEMIC {
					panic("@import statement must be followed by ;")
				}

				moreTokens, moreOutput, err := p.start(
					pwd,
					contents)
				// If importing was successful, each token must be moved forward
				// by the position of the @import call that made it available.
				for i, _ := range moreTokens {
					moreTokens[i].Pos += last.Pos
				}

				if err != nil {
					panic(err)
				}
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
