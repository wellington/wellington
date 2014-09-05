package sprite_sass

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	//. "github.com/kr/pretty"
)

type Parser struct {
	Output  string
	Vars    map[string]string
	Sprites map[string]ImageList
}

// Parser reads the tokens from the lexer and performs
// conversions and/or substitutions for sprite*() calls.
//
// Parser creates a map of all variables and sprites
// (created via sprite-map calls).
func (p Parser) Start(f string) []byte {
	p.Vars = make(map[string]string)
	p.Sprites = make(map[string]ImageList)
	fvar, _ := ioutil.ReadFile(f)
	i := string(fvar)
	tokens, input, err := parser(i, filepath.Dir(f))

	if err != nil {
		log.Fatal(err)
	}
	var (
		t, cmd string
	)
	for i := 0; i < len(tokens); i = i + 1 {
		token := tokens[i]
		// Generate list of vars
		if token.Type == VAR {
			t = fmt.Sprintf("%s", token)
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
					cmd = fmt.Sprintf("%s", token)
					val += cmd
				case FILE:
					if cmd == "sprite-map" {
						imgs := ImageList{}
						glob := fmt.Sprintf("%s", token)
						imgs.Decode(glob)
						imgs.Vertical = true
						imgs.Combine()
						p.Sprites[t] = imgs

						//TODO: Generate filename
						//imgs.Export("generated.png")
						cmd = ""
					}
				case SUB:
					// Can this ever happen, do we care?
					fmt.Println("SUB")
				default:
					//fmt.Printf("Default: %s\n", token)
					val += fmt.Sprintf("%s", token)
				}
				if !nested && token.Type != CMD {
					break
				}
			}
			p.Vars[t] = val
			//Replace subsitution tokens
		} else if token.Type == SUB {
			if cmd == "sprite" {
				//Capture sprite
				sprite := p.Sprites[fmt.Sprintf("%s", token)]
				//Capture filename
				i++
				tokens[i].Value = sprite.CSS(fmt.Sprintf("%s",
					tokens[i]))
				tokens[i].Write = true
				tokens = append(tokens[:i-3], tokens[i:]...)
				i = i - 3
				cmd = ""
			} else {
				tokens[i].Value = p.Vars[token.Value]
			}
		} else if token.Type == CMD {
			cmd = fmt.Sprintf("%s", token)
		}
	}

	return process(input, tokens, 0)
}

func process(in string, items []Item, pos int) []byte {

	var out []byte
	l := len(items)

	if pos >= len(in) {
		return []byte("")
	}

	if items[0].Type == CMD && items[0].Value == "sprite" {
		i := 1
		//out = append(out, items[0].Value...)
		//Skip to semicolon
		for ; items[i].Write || i > l; i++ {
		}
		return append(out, process(in, items[i:], items[i].Pos)...)
	}

	if items[0].Write {
		i := 1
		out = append(out, items[0].Value...)
		//Skip to semicolon
		for ; items[i].Type != SEMIC || i > l; i++ {
		}
		return append(out, process(in, items[i:], pos)...)
	}

	if l > 1 {
		if items[1].Write {
			out = append(out, items[0].Value...)
			out = append(out, ':', ' ')
		} else {
			out = append(out, in[items[0].Pos:items[1].Pos]...)
		}
		out = append(out, process(in, items[1:], pos)...)
	} else {
		out = append(out, in[items[0].Pos:]...)
	}

	return out
}

// parser recrusively resolves all imports and tokenizes the
// input string
func parser(input, path string) ([]Item, string, error) {

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

		if err != nil {
			return nil, string(output), fmt.Errorf("Error: %v (pos %d)", err, item.Pos)
		}
		if item.Type == ItemEOF {
			output = append(output, input[pos:]...)
			return status, string(output), nil
		} else if item.Type == IMPORT {
			output = append(output, input[pos:item.Pos]...)
			last = item
			importing = true
		} else {
			if importing {
				//Load and retrieve all tokens from imported file
				path := fmt.Sprintf(
					"%s/_%s.scss",
					path, *item)

				file, err := ioutil.ReadFile(path)
				if err != nil {
					fullpath, _ := filepath.Abs(path)
					log.Fatal("Cannot import path: ", fullpath)
				}
				//pos = item.Pos + len(item.Value) + 2 //Adjust for ";
				//Eat the semicolon
				item := lex.Next()
				pos = item.Pos + len(item.Value)
				if item.Type != SEMIC {
					log.Fatal("@import must be followed by ;")
				}
				//pos = item.Pos + len(item.Value)
				moreTokens, moreOutput, err := parser(string(file),
					filepath.Dir(path))
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
