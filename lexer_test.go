package sprite_sass

import (
	"fmt"
	"io/ioutil"
	"testing"
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

	sel := items[0].String()
	if e := "$images"; e != sel {
		t.Errorf("Invalid VAR parsing expected: %s, was: %s",
			e, sel)
	}
	sel = items[10].String()
	if e := "#00FF00"; e != sel {
		t.Errorf("Invalid TEXT parsing expected: %s, was: %s",
			e, sel)
	}

	if e := "sprite-map"; e != items[1].String() {
		t.Errorf("Invalid CMD parsing expected: %s, was: %s",
			e, items[1].String())
	}
	sel = items[3].String()
	if e := "*.png"; e != sel {
		t.Errorf("Invalid FILE parsing expected: %s, was: %s",
			sel, e)
	}
	T := items[3].Type
	if e := FILE; e != T {
		t.Errorf("Invalid FILE type parsing expected: %s, was: %s",
			e, T)
	}
}

func TestLexerImport(t *testing.T) {
	fvar, _ := ioutil.ReadFile("test/import.scss")
	items, _ := parse(string(fvar))
	sel := items[0].String()
	if e := "background"; sel != e {
		t.Errorf("Invalid token expected: %s, was %s", e, sel)
	}
	sel = items[1].String()
	if e := "purple"; sel != e {
		t.Errorf("Invalid token expected: %s, was %s", e, sel)
	}
	sel = items[3].String()
	if e := "@import"; sel != e {
		t.Errorf("Invalid token expected: %s, was %s", e, sel)
	}
	sel = items[4].String()
	if e := "var"; sel != e {
		t.Errorf("Invalid token expected: %s, was %s", e, sel)
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
			status = append(status, *item)
			//fmt.Printf("Default: %d %s\n", item.Pos, item)
		}
	}
}
