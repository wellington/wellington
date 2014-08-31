package lexer

import (
	"fmt"
	"io/ioutil"
	"log"
)

func Parser(f string) {

	vars := make(map[string]string)
	fvar, _ := ioutil.ReadFile(f)
	tokens, err := parser(string(fvar))

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(tokens); i = i + 1 {
		token := tokens[i]
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
				default:
					val += fmt.Sprintf("%s", token)
				}
				if !nested && token.Type != CMD {
					break
				}
			}
			vars[t] = val
		}
	}
	for k, v := range vars {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func parser(input string) ([]Item, error) {
	lex := New(func(lex *Lexer) StateFn {
		return lex.Action()
	}, input)

	var status []Item
	for {
		item := lex.Next()
		err := item.Error()
		if err != nil {
			return nil, fmt.Errorf("Error: %v (pos %d)", err, item.Pos)
		}
		switch item.Type {
		case ItemEOF:
			return status, nil
		case CMD, SPRITE, TEXT, VAR, FILE:
			fallthrough
		case LPAREN, RPAREN,
			LBRACKET, RBRACKET:
			fallthrough
		case EXTRA:
			fallthrough
		case item.Type:
			status = append(status, *item)
		default:
			fmt.Printf("Default: %d %s\n", item.Pos, item)
		}
	}

}
