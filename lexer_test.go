package sprite_sass_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestSassLexer(t *testing.T) {

	fvar, _ := ioutil.ReadFile("../test/var.scss")

	items, err := parse(string(fvar))
	if err != nil {
		t.Errorf("Error parsing string")
	}
	for _, item := range items {
		v := fmt.Sprintf("%s", item)
		switch item.Type {
		case VAR:
			if !strings.HasPrefix(v, "$") {
				t.Errorf("Invalid variable prefix")
			}
			if strings.Index(v, ":") > -1 {
				t.Errorf("Invalid symbol in variable")
			}
		case CMD:
			if !strings.HasPrefix(v, "sprite") {
				t.Errorf("Invalid command name: %s", v)
			}
		case FILE:
			//File globbing is a vast and varied field
			// TODO: crib tests from http://golang.org/src/pkg/path/filepath/match_test.go
			if !strings.HasSuffix(v, "png") {
				t.Errorf("File safety test failed expected png$, was: %s", v)
			}
		default:
			//fmt.Println(item.Type)
		}

	}

}

// create a parser for the language.
func parse(input string) ([]Item, error) {
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
			status = append(status, *item)
		default:
			fmt.Printf("Default: %d %s\n", item.Pos, item)
		}
	}
}
