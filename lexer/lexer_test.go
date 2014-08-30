package lexer_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/drewwells/sprite-sass/lexer"
)

func TestSassLexer(t *testing.T) {

	// create a StateFn to parse the language.
	var start lexer.StateFn
	start = func(lex *lexer.Lexer) lexer.StateFn {
		return lex.Action()
	}

	// create a parser for the language.
	parse := func(input string) ([]lexer.Item, error) {
		lex := lexer.New(start, input)

		var status []lexer.Item
		for {
			item := lex.Next()
			err := item.Err()
			//fmt.Printf("Item: %s\n", item)
			if err != nil {
				return nil, fmt.Errorf("Error: %v (pos %d)", err, item.Pos)
			}
			switch item.Type {
			case lexer.ItemEOF:
				return status, nil
			case lexer.CMD, lexer.SPRITE, lexer.TEXT, lexer.VAR, lexer.FILE:
				status = append(status, *item)
			case lexer.EXTRA:
				status = append(status, *item)
			default:
				fmt.Printf("Default: %d %s\n", item.Pos, item)
			}
		}
	}

	sheet1, _ := ioutil.ReadFile("../test/sheet1.scss")

	// parse a valid string and print the status
	status, err := parse(string(sheet1)) //`sprite($images,"one.png");`)

	for _, item := range status {
		fmt.Printf("%8s: %s\n", item.Type, item.String())
	}
	fmt.Printf("  Status: %s", err)

}
