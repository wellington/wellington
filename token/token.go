// Package token defines constants representing the lexical tokens of the Go
// programming language and basic operations on tokens (printing, predicates).
//
package token

// A type for all the types of items in the language being lexed.
// These only parser SASS specific language elements and not CSS.
type ItemType int

// Special item types.
const (
	NotFound          = -1
	ItemEOF  ItemType = iota
	ItemError
	IF
	ELSE
	EACH
	IMPORT
	INCLUDE
	INTP
	FUNC
	MIXIN
	EXTRA
	CMD
	VAR
	CMDVAR
	SUB
	VALUE
	FILE
	cmd_beg
	SPRITE
	SPRITEF
	SPRITED
	SPRITEH
	SPRITEW
	cmd_end
	NUMBER
	TEXT
	DOLLAR
	math_beg
	PLUS
	MINUS
	MULT
	DIVIDE
	math_end
	special_beg
	LPAREN
	RPAREN
	LBRACKET
	RBRACKET
	SEMIC
	COLON
	CMT
	special_end
	include_mixin_beg
	BKND
	include_mixin_end
	FIN
)

var Tokens = [...]string{
	ItemEOF:   "eof",
	ItemError: "error",
	IF:        "@if",
	ELSE:      "@else",
	EACH:      "@each",
	IMPORT:    "@import",
	INCLUDE:   "@include",
	INTP:      "#{",
	FUNC:      "@function",
	MIXIN:     "@mixin",
	EXTRA:     "extra",
	CMD:       "command",
	VAR:       "variable",
	CMDVAR:    "command-variable",
	SUB:       "sub",
	VALUE:     "value",
	FILE:      "file",
	SPRITE:    "sprite",
	SPRITEF:   "sprite-file",
	SPRITED:   "sprite-dimensions",
	SPRITEH:   "sprite-height",
	SPRITEW:   "sprite-width",
	NUMBER:    "number",
	TEXT:      "text",
	DOLLAR:    "$",
	PLUS:      "+",
	MINUS:     "-",
	MULT:      "*",
	DIVIDE:    "/",
	LPAREN:    "(",
	RPAREN:    ")",
	LBRACKET:  "{",
	RBRACKET:  "}",
	SEMIC:     ";",
	COLON:     ":",
	CMT:       "comment",
	BKND:      "background",
	FIN:       "FINISHED",
}

func (i ItemType) String() string {
	if i < 0 {
		return ""
	}
	return Tokens[i]
}

var directives map[string]ItemType

func init() {
	directives = make(map[string]ItemType)
	for i := cmd_beg; i < cmd_end; i++ {
		directives[Tokens[i]] = i
	}
}

// Lookup ItemType by token string
func Lookup(ident string) ItemType {
	if tok, is_keyword := directives[ident]; is_keyword {
		return tok
	}
	return NotFound
}

// Token is the set of lexical tokens of the Go programming language.
// Token is not currently used, but should be added to ItemType as
// CSS parsing is needed
type Token int

// The list of tokens.
const (
	css_beg Token = iota
	EM
	EX
	PX
	CM
	MM
	PT
	PC
	DEG
	RAD
	GRAD
	MS
	S
	HZ
	KHZ
	css_end

	vendor_beg
	OPERA
	WEBKIT
	MOZ
	VENDORMS
	KHTML
	vendor_end

	cssfunc_beg
	CHARSET
	MEDIA
	KEYFRAMES
	ONLY
	RGB
	URL
	IMAGEURL
	IMPORTANT
	NOT
	EVEN
	ODD
	PROGID
	EXPRESSION
	CALC
	MOZCALC
	WEBKITCALC
	cssfunc_end
)
