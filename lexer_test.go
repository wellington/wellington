package sprite_sass_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestBools(t *testing.T) {
	if IsEOF('%', 0) != true {
		t.Errorf("Did not detect EOF")
	}
}

func TestSassLexer(t *testing.T) {

	fvar, _ := ioutil.ReadFile("test/_var.scss")

	items, err := parse(string(fvar))

	if err != nil {
		t.Errorf("Error parsing string")
	}

	if e := "$red-var"; e != items[0].String() {
		t.Errorf("Invalid VAR parsing expected: %s, was: %s",
			e, items[0].String())
	}

	if e := "#00FF00"; e != items[3].String() {
		t.Errorf("Invalid TEXT parsing expected: %s, was: %s",
			e, items[3].String())
	}

	if e := "sprite-map"; e != items[9].String() {
		t.Errorf("Invalid CMD parsing expected: %s, was: %s",
			e, items[9].String())
	}

	if e := "test/*.png"; e != items[11].String() {
		t.Errorf("Invalid FILE parsing expected: %s, was: %s",
			items[11].String(), e)
	}

	if e := FILE; e != items[11].Type {
		t.Errorf("Invalid FILE type parsing expected: %s, was: %s",
			e, items[11].Type)
	}
}

func TestLexerImport(t *testing.T) {
	fvar, _ := ioutil.ReadFile("test/import.scss")
	items, _ := parse(string(fvar))

	if e := "background"; items[0].String() != e {
		t.Errorf("Invalid token expected: %s, was %s", e, items[0])
	}

	if e := "purple"; items[1].String() != e {
		t.Errorf("Invalid token expected: %s, was %s", e, items[0])
	}

	if e := "@import"; items[2].String() != e {
		t.Errorf("Invalid token expected: %s, was %s", e, items[0])
	}
	return
	if e := "var"; items[3].String() != e {
		t.Errorf("Invalid token expected: %s, was %s", e, items[0])
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
		case CMD, SPRITE, TEXT, VAR, FILE, SUB:
			fallthrough
		case LPAREN, RPAREN,
			LBRACKET, RBRACKET:
			fallthrough
		case IMPORT:
			fallthrough
		case EXTRA:
			status = append(status, *item)
		default:
			//fmt.Printf("Default: %d %s\n", item.Pos, item)
		}
	}
}
