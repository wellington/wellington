package lexer

import (
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/drewwells/sprite_sass/token"
)

func printItems(items []Item) {
	for i, item := range items {
		fmt.Printf("%4d: %s %s\n", i, item.Type, item.Value)
	}
}

func TestLexerBools(t *testing.T) {
	if IsEOF('%', 0) != true {
		t.Errorf("Did not detect EOF")
	}
}

func TestLexer(t *testing.T) {

	lex := New(nil, "")
	if lex != nil {
		t.Errorf("non-nil Lexer on nil state")
	}

	fvar, _ := ioutil.ReadFile("../test/sass/_var.scss")

	items, err := testParse(string(fvar))

	if err != nil {
		t.Errorf("Error parsing string")
	}

	sel := items[0].String()
	if e := "@import"; e != sel {
		t.Errorf("Invalid VAR parsing expected: %s, was: %s",
			e, sel)
	}

	if e := "sprite-map"; e != items[5].String() {
		t.Errorf("Invalid CMD parsing expected: %s, was: %s",
			e, items[1].String())
	}

	sel = items[7].String()
	if e := "*.png"; e != sel {
		t.Errorf("Invalid FILE parsing expected: %s, was: %s",
			e, sel)
	}

	T := items[7].Type
	if e := FILE; e != T {
		t.Errorf("Invalid FILE parsing expected: %s, was: %s",
			e, T)
	}

	sel = items[16].String()
	if e := "#00FF00"; e != sel {
		t.Errorf("Invalid TEXT parsing expected: %s, was: %s",
			e, sel)
	}
}

func TestLexerComment(t *testing.T) {
	in := `/* some;
multiline comments +*-0
with symbols in them*/
//*Just a specially crafted single line comment
div {}
/* Invalid multiline comment`
	items, err := testParse(in)
	if err != nil {
		panic(err)
	}

	if e := `/* some;
multiline comments +*-0
with symbols in them*/`; items[0].Value != e {
		t.Errorf("Multiline comment mismatch expected:%s\nwas:%s",
			e, items[0].Value)

	}
	if e := CMT; e != items[0].Type {
		t.Errorf("Multiline CMT mismatch expected:%s, was:%s",
			e, items[0].Type)
	}
	if e := CMT; e != items[1].Type {
		t.Errorf("CMT with special chars mismatch expected:%s, was:%s",
			e, items[1].Type)
	}

	if e := CMT; e != items[5].Type {
		t.Errorf("CMT with invalid ending expected: %s, was: %s",
			e, items[5].Type)
	}
	if e := 6; len(items) != e {
		t.Errorf("Invalid number of comments expected: %d, was: %d",
			len(items), e)
	}
}

func TestLexerSub(t *testing.T) {
	in := `$name: foo;
$attr: border;
p.#{$name} {
  #{$attr}-color: blue;
}`
	items, err := testParse(in)

	if err != nil {
		panic(err)
	}
	vals := map[int]string{
		4:  "$attr",
		13: "#{",
		0:  "$name",
	}
	errors := false
	for i, v := range vals {
		if v != items[i].Value {
			errors = true
			t.Errorf("at %d expected: %s, was: %s", i, v, items[i].Value)
		}
	}
	if errors {
		printItems(items)
	}
}

func TestLexerCmds(t *testing.T) {
	in := `$s: sprite-map("test/*.png");
$file: sprite-file($s, 140);
div {
  width: image-width($file, 140);
  height: image-height(sprite-file($s, 140));
  url: sprite-file($s, 140);
}`
	items, err := testParse(in)
	if err != nil {
		panic(err)
	}

	types := map[int]ItemType{
		0:  VAR,
		2:  CMDVAR,
		4:  FILE,
		7:  VAR,
		9:  CMD,
		11: SUB,
		12: FILE,
		17: TEXT,
		19: CMD,
		21: SUB,
		22: FILE,
		27: CMD,
		29: CMD,
		32: FILE,
		40: SUB,
		41: FILE,
	}
	errors := false
	for i, tp := range types {
		if tp != items[i].Type {
			errors = true
			t.Errorf("at %d expected: %s, was: %s", i, tp, items[i].Type)
		}
	}
	if errors {
		printItems(items)
	}
}

func TestLexerImport(t *testing.T) {
	fvar, _ := ioutil.ReadFile("../test/sass/import.scss")
	items, _ := testParse(string(fvar))
	vals := map[int]string{
		0: "@import",
		1: "var",
		2: ";",
	}
	errors := false
	for i, v := range vals {
		if v != items[i].Value {
			errors = true
			t.Errorf("at %d expected: %s, was: %s", i, v, items[i].Value)
		}
	}
	if errors {
		printItems(items)
	}
}

// Test disabled due to not working
func TestLexerSubModifiers(t *testing.T) {
	in := `$s: sprite-map("*.png");
div {
  height: -1 * sprite-height($s,"140");
  width: -sprite-width($s,"140");
  margin: - sprite-height($s, "140")px;
  height: image-height(test/140.png);
  width: image-width(sprite-file($s, 140));
}`

	items, err := testParse(in)
	if err != nil {
		panic(err)
	}
	if e := ":"; items[1].Value != e {
		t.Errorf("Failed to parse symbol expected: %s, was: %s",
			e, items[1].Value)
	}
	if e := "*.png"; items[4].Value != e {
		t.Errorf("Failed to parse file expected: %s, was: %s",
			e, items[4].Value)
	}

	if e := "*"; items[13].Value != e {
		t.Errorf("Failed to parse text expected: %s, was: %s",
			e, items[13].Value)
	}

	if e := MINUS; items[22].Type != e {
		t.Errorf("Failed to parse CMD expected: %s, was: %s",
			e, items[22].Type)
	}

	if e := CMD; items[23].Type != e {
		t.Errorf("Failed to parse CMD expected: %s, was: %s",
			e, items[23].Type)
	}

	if e := TEXT; items[37].Type != e {
		t.Errorf("Failed to parse TEXT expected: %s, was: %s",
			e, items[37].Type)
	}

	if e := FILE; items[43].Type != e {
		t.Errorf("Type mismatch expected: %s, was: %s", e, items[43].Type)
	}
	types := map[int]ItemType{
		48: CMD,
		50: CMD,
		52: SUB,
		53: FILE,
	}
	for i, ty := range types {
		if types[i] != ty {
			t.Errorf("Type mismatch at %d expected: %s, was: %s", i, types[i], ty)
		}
	}
}

func TestLexerVars(t *testing.T) {
	in := `$a: 1;
$b: $1;
$c: ();
$d: $c`

	items, err := testParse(in)
	if err != nil {
		panic(err)
	}
	_ = items
}

func TestLexerWhitespace(t *testing.T) {
	in := `$s: sprite-map("*.png");
div {
  background:sprite($s,"140");
}`
	items, err := testParse(in)
	if err != nil {
		panic(err)
	}

	if e := TEXT; items[9].Type != e {
		t.Errorf("Type parsed improperly expected: %s, was: %s",
			e, items[9].Type)
	}

	if e := CMD; items[11].Type != e {
		t.Errorf("Type parsed improperly expected: %s, was: %s",
			e, items[11].Type)
	}

	if e := "sprite"; items[11].Value != e {
		t.Errorf("Command parsed improperly expected: %s, was: %s",
			e, items[11].Value)
	}
}

// create a parser for the language.
func testParse(input string) ([]Item, error) {
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

func TestLexerLookup(t *testing.T) {
	it := Lookup("sprite-file")
	if e := "sprite-file"; it.String() != e {
		t.Errorf("Directive should be found was: %s, expected: %s",
			it.String(), e)
	}
	it = Lookup("NOT GONNA FIND")
	if e := ""; it.String() != e {
		t.Errorf("Not a token was: %s, expected: %s", it.String(), e)
	}
	it = Lookup("/")
	if e := ""; it.String() != e {
		t.Errorf("Non-directive was: %s, expected: %s", it.String(), e)
	}
}
