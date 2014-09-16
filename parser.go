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
	Index, shift         int
	Chop                 []Replace
	Pwd, Input, ImageDir string
	Includes             []string
	Items                []Item
	Output               []byte
	Sprites              map[string]ImageList
	Vars                 map[string]string
}

// Parser reads the tokens from the lexer and performs
// conversions and/or substitutions for sprite*() calls.
//
// Parser creates a map of all variables and sprites
// (created via sprite-map calls).
func (p *Parser) Start(in io.Reader, pkgdir string) []byte {
	p.Vars = make(map[string]string)
	p.Sprites = make(map[string]ImageList)

	if p.ImageDir == "" {
		p.ImageDir = pkgdir
	}
	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	buf.ReadFrom(in)

	tokens, input, err := p.start(pkgdir, string(buf.Bytes()))
	p.Input = input
	p.Items = tokens
	if err != nil {
		panic(err)
	}
	var (
		def, cmd string
	)
	for i := 0; i < len(tokens); i = i + 1 {
		token := tokens[i]
		last := i
		// Generate list of vars
		if token.Type == VAR {
			def = fmt.Sprintf("%s", token)
			val := ""
			nested := false
			for {
				i++
				token = tokens[i]
				switch token.Type {
				case LPAREN:
					nested = true
				case RPAREN:
					nested = false
				case CMD:
					// Changing the behavior of CMD!
					cmd = fmt.Sprintf("%s", token)
					val += cmd
				case FILE:
					i = p.File(cmd, last, i)
					def = ""
					cmd = ""
				case SUB:
					// fmt.Println("SUB:", tokens[i-1], tokens[i], tokens[i+1])
					// fmt.Println(p.Input[tokens[i-20].Pos:tokens[i+20].Pos])
					// Cowardly give up and hope these variables do not matter
					// Cases:
					// - @for $i from 1 through $variable
					// - $variable: ($i - 1)
					// -
					fallthrough
				default:
					// fmt.Printf("Default: %s\n", token)
					val += fmt.Sprintf("%s", token)
				}

				if !nested && tokens[i].Type != CMD {
					break
				}
			}
			if def != "" {
				p.Vars[def] = val
			}
			//Replace subsitution tokens
		} else if token.Type == SUB {
			repl := ""
			switch cmd {
			case "sprite":
				//Capture sprite
				sprite := p.Sprites[fmt.Sprintf("%s", token)]
				//Capture filename
				i++
				name := fmt.Sprintf("%s", tokens[i])
				repl = sprite.CSS(name)

				p.Mark(tokens[i-3].Pos, tokens[i+2].Pos, repl)
				tokens = append(tokens[:i-3], tokens[i:]...)
				i = i - 3
				def = ""
				cmd = ""
			case "sprite-height":
				sprite := p.Sprites[fmt.Sprintf("%s", token)]
				repl = fmt.Sprintf("height: %dpx;",
					sprite.ImageHeight(tokens[i+1].String()))
				// Walk forward to file name
				i++
				p.Mark(tokens[i-4].Pos, tokens[i+3].Pos, repl)
				tokens = append(tokens[:i-4], tokens[i:]...)
				i = i - 4
				def = ""
				cmd = ""
			case "sprite-width":
				sprite := p.Sprites[fmt.Sprintf("%s", token)]
				repl = fmt.Sprintf("width: %dpx;",
					sprite.ImageWidth(tokens[i+1].String()))
				// Walk forward to file name
				i++
				p.Mark(tokens[i-4].Pos, tokens[i+3].Pos, repl)
				tokens = append(tokens[:i-4], tokens[i:]...)
				i = i - 4
				def = ""
				cmd = ""
			case "sprite-dimensions":
				sprite := p.Sprites[fmt.Sprintf("%s", token)]
				repl = sprite.Dimensions(tokens[i+1].String())
				// Walk forward to file name
				i++
				p.Mark(tokens[i-4].Pos, tokens[i+3].Pos, repl)
				tokens = append(tokens[:i-4], tokens[i:]...)
				i = i - 4
				def = ""
				cmd = ""
			default:
				tokens[i].Value = p.Vars[token.Value]
			}
		} else if token.Type == CMD {
			// Sync the index during the refactor
			p.Index = i
			cmd = fmt.Sprintf("%s", token)
			switch token.Value {
			case "inline-image":
				cmd = ""
				p.Mixin()
			default:
				//log.Fatal("Danger will robinson")
			}
			i = p.Index
		}
	}
	// I don't recall the point of this, but process
	// will result in whitespace errors in the output.
	// p.Output = process(p.Input, p.Items, 0)
	p.Output = []byte(p.Input)
	p.Replace()
	// DEBUG
	for _, item := range p.Items {
		_ = item
		// fmt.Printf("%s %s\n", item.Type, item)
	}
	return p.Output
}

// Mixin processes tokens in the format @include mixin(args...)
// and perform requested actions.
func (p *Parser) Mixin() {

	// Commands always end in ); else panic
	start := p.Index
	cmd := p.Items[start]

	var file Item
	for {
		cur := p.Items[p.Index]
		if cur.Type == RPAREN {
			p.Index++
			if p.Items[p.Index].Type != SEMIC {
				f, l := p.Items[start].Pos, p.Items[p.Index].Pos
				log.Fatal("Commands must end with semicolon",
					fmt.Sprintf(p.Input[f:l]))
			} else {
				break
			}
		} else if cur.Type == FILE {
			file = cur
		}
		p.Index++
	}
	if file.Type != FILE {
		log.Fatal("File for command was not found")
	}
	if cmd.Value == "inline-image" {
		img := ImageList{}
		img.Decode(p.ImageDir + "/" + file.Value)
		repl := img.Inline()
		p.Index-- // Preserve the final semic
		p.Mark(cmd.Pos-1, p.Items[p.Index].Pos+1, repl)
	}
}

// Replace iterates through the list of substrings to
// cut or replace, adjusting for shift of the output
// buffer as a result of these ops.
func (p *Parser) Replace() {

	for _, c := range p.Chop {

		begin := c.Start - p.shift
		end := c.End - p.shift
		// fmt.Println(string(p.Input[begin:end]), "~>`"+string(c.Value)+"`")
		// fmt.Println("*******")
		// Adjust shift for number of bytes deleted and inserted
		p.shift += end - begin
		p.shift -= len(c.Value)
		suf := append(c.Value, p.Output[end:]...)

		p.Output = append(p.Output[:begin], suf...)
	}
}

// Mark segments of the input string for future deletion.
func (p *Parser) Mark(start, end int, val string) {
	// fmt.Println("Mark:", string(p.Input[start:end]), "~")
	p.Chop = append(p.Chop, Replace{start, end, []byte(val)})
}

// Processes file which usually mean cutting some of the input
// text.
func (p *Parser) File(cmd string, start, end int) int {
	first := p.Items[start]
	item := p.Items[end]
	// Find the next newline, failing that find the semicolon
	i := end
	if cmd == "sprite-map" {
		for ; p.Items[i].Type != RPAREN; i++ {
		}
		i = i - 1
		// Verify that the statement ends with semicolon
		interest := p.Items[i+3]
		// Mark this area for deletion, since doing so now would
		// invalidate all subsequent tokens positions.
		p.Mark(first.Pos, interest.Pos, "")
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
	return i
}

func process(in string, items []Item, pos int) []byte {

	var out []byte
	l := len(items)

	if pos >= len(in) {
		return []byte("")
	}

	// TODO: There's an error where items[1] has an invalid
	// position.
	if l > 1 && items[1].Pos > items[0].Pos {
		out = append(out, in[items[0].Pos:items[1].Pos]...)
		out = append(out, process(in, items[1:], pos)...)
	} else {
		out = append(out, in[items[0].Pos:]...)
	}

	return out
}

// parser recrusively resolves all imports and tokenizes the
// input string
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
				// Lexer needs to be adjusted for current
				// position of end of @import
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
