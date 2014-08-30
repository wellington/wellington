package lexer_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/drewwells/sprite-sass/lexer"
)

func TestSassLexer(t *testing.T) {

	fvar, _ := ioutil.ReadFile("../test/var.scss")

	items, err := parse(string(fvar))
	if err != nil {
		t.Errorf("Error parsing string")
	}
	for _, item := range items {
		v := fmt.Sprintf("%s", item)
		switch fmt.Sprintf("%s", item.Type) {
		case "variable":
			if !strings.HasPrefix(v, "$") {
				t.Errorf("Invalid variable prefix")
			}
			if strings.Index(v, ":") > -1 {
				t.Errorf("Invalid symbol in variable")
			}
		case "command":
			if !strings.HasPrefix(v, "sprite") {
				t.Errorf("Invalid command name: %s", v)
			}
		case "file":
			//File globbing is a vast and varied field
			// TODO: crib tests from http://golang.org/src/pkg/path/filepath/match_test.go
			if !strings.HasSuffix(v, "png") {
				t.Errorf("File safety test failed expected png$, was: %s", v)
			}
		default:
			fmt.Println(item.Type)
		}

	}

}

// create a parser for the language.
func parse(input string) ([]lexer.Item, error) {
	lex := lexer.New(func(lex *lexer.Lexer) lexer.StateFn {
		return lex.Action()
	}, input)

	var status []lexer.Item
	for {
		item := lex.Next()
		err := item.Error()
		if err != nil {
			return nil, fmt.Errorf("Error: %v (pos %d)", err, item.Pos)
		}
		switch item.Type {
		case lexer.ItemEOF:
			return status, nil
		case lexer.CMD, lexer.SPRITE, lexer.TEXT, lexer.VAR, lexer.FILE:
			fallthrough
		case lexer.LPAREN, lexer.RPAREN,
			lexer.LBRACKET, lexer.RBRACKET:
			fallthrough
		case lexer.EXTRA:
			fallthrough
		case item.Type:
			status = append(status, *item)
		default:
			fmt.Printf("Default: %d %s\n", item.Pos, item)
		}
	}
}
