package sprite_sass

import (
	"fmt"
	"io/ioutil"
	"log"
)

type Parser struct {
	Output  string
	Vars    map[string]string
	Sprites map[string]ImageList
}

func (p Parser) Start(f string) {

	p.Vars = make(map[string]string)
	fvar, _ := ioutil.ReadFile(f)
	tokens, err := parser(string(fvar))

	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(tokens); i = i + 1 {
		token := tokens[i]
		// Generate list of vars
		if token.Type == VAR {
			t, val := fmt.Sprintf("%s", token), ""
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
					val += fmt.Sprintf("RUN: %s", token)
				case SUB:
					fmt.Println("SUB")
				default:
					val += fmt.Sprintf("%s", token)
				}
				if !nested && token.Type != CMD {
					break
				}
			}
			p.Vars[t] = val
			//Replace subsitution tokens
		} else if token.Type == SUB {
			tokens[i].Value = p.Vars[token.Value]
		}
	}
	for _, v := range tokens {
		fmt.Printf("%s\n", v)
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

		switch item.Type {
		case ItemEOF:
			return status, nil
		case SPRITE, TEXT, VAR, FILE:
			fallthrough
		case LPAREN, RPAREN,
			LBRACKET, RBRACKET:
			fallthrough
		case EXTRA:
			status = append(status, *item)
		case SUB:
			status = append(status, *item)
		case CMD:
			if item.String() == "sprite-map" {
				status = append(status, *item)
			}
		default:
			fmt.Printf("Default: %d %s\n", item.Pos, item)
		}
	}

}
