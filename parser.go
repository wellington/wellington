package sprite_sass

import (
	"fmt"
	"io/ioutil"
	"log"

	//. "github.com/kr/pretty"
)

type Parser struct {
	Output  string
	Vars    map[string]string
	Sprites map[string]ImageList
}

func (p Parser) Start(f string) {
	p.Vars = make(map[string]string)
	p.Sprites = make(map[string]ImageList)
	fvar, _ := ioutil.ReadFile(f)
	tokens, err := parser(string(fvar))

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
					// Can this ever happen, do we care?
				case SUB:
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
				token = tokens[i]
				tokens[i].Value = sprite.CSS(fmt.Sprintf("%s", token))
				cmd = ""
			} else {
				tokens[i].Value = p.Vars[token.Value]
			}
		} else if token.Type == CMD {
			cmd = fmt.Sprintf("%s", token)
		}
	}
}

func parser(input string) ([]Item, error) {

	var status []Item
	lex := New(func(lex *Lexer) StateFn {
		return lex.Action()
	}, input)

	for {
		item := lex.Next()
		err := item.Error()
		if err != nil {
			return nil, fmt.Errorf("Error: %v (pos %d)", err, item.Pos)
		}
		if item.Type == ItemEOF {
			return status, nil
		} else {
			status = append(status, *item)
		}
		// switch item.Type {
		// case ItemEOF:
		// 	return status, nil
		// case SPRITE, TEXT, VAR, FILE:
		// 	fallthrough
		// case LPAREN, RPAREN,
		// 	LBRACKET, RBRACKET:
		// 	fallthrough
		// case EXTRA:
		// 	status = append(status, *item)
		// case SUB:
		// 	status = append(status, *item)
		// case CMD:
		// 	status = append(status, *item)
		// default:
		// 	fmt.Printf("Default: %d %s\n", item.Pos, item)
		// }
	}

}
