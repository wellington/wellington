package sprite_sass

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

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
func (p Parser) Start(f string) string {
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
					token)) + ";"
				tokens[i].Write = true
				cmd = ""
			} else {
				tokens[i].Value = p.Vars[token.Value]
			}
		} else if token.Type == CMD {
			cmd = fmt.Sprintf("%s", token)
		}
	}

	//Iterate through tokens looking for ones to write out
	var (
		output []byte
		pos    int
	)
	//reader := strings.NewReader(input)
	_ = output
	for i, token := range tokens {
		//fmt.Printf("%s ", token)
		//These tokens get replaced
		if token.Write {
			//fmt.Printf("WRITE: %s %s\n", token, token.Type)
			output = append(output, '~')
			output = append(output, token.Value...)
			output = append(output, '~')
			fmt.Printf("\n%s\n", input[pos:])
			pos = pos + strings.IndexRune(input[pos:], ';') - 1

		} else {
			//fmt.Printf("NOWRITE: %s TYPE:%s\n", token, token.Type)
			if token.Type == CMD && token.Value == "sprite" {
				//Don't write out CMDs
				if i < len(tokens)-1 {
					//output = append(output, '#')
					output = append(output, input[pos:token.Pos]...)
					//output = append(output, '#')
					pos = tokens[i+1].Pos + 1
				} else {
					break
				}
			} else if pos < len(input) && token.Type != SUB && token.Type != LPAREN && token.Type != RPAREN { //&& token.Pos >= pos {
				text := fmt.Sprintf("%s", token)
				//fmt.Printf("NOWRITE: %s %s\n", text, token.Type)
				l := len(text)
				//output = append(output, '/', '|')
				output = append(output, input[pos:token.Pos]...)
				//output = append(output, '|', '/')
				//output = append(output, '^')
				output = append(output, text...)
				//output = append(output, '^')
				pos = token.Pos + l
				if pos > len(input) {
					panic("NOTHERE")
					break
				}
			} else if token.Pos <= pos {
				pos = token.Pos
				//fmt.Printf("ERROR: %s %d < %d\n", token, token.Pos, pos)
			}
		}
	}
	//output = append(output, input[pos+1:]...)
	fmt.Println("\nOutput")
	fmt.Println(string(output))
	return ""
}

// parser recrusively resolves all imports and tokenizes the
// input string
func parser(input, path string) ([]Item, string, error) {

	var (
		status    []Item
		importing bool
		output    []byte
		pos       int
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
			importing = true
		} else {
			if importing {
				//Load and retrieve all tokens from imported file
				file, err := ioutil.ReadFile(fmt.Sprintf(
					"%s/_%s.scss",
					path, *item))
				pos = item.Pos + len(item.Value) + 2 //Adjust for ";
				moreTokens, moreOutput, err := parser(string(file), filepath.Dir(path))
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
