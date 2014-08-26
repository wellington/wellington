// Package token defines constants representing the lexical tokens of the Go
// programming language and basic operations on tokens (printing, predicates).
//
package token

import "strconv"

// Token is the set of lexical tokens of the Go programming language.
type Token int

// The list of tokens.
const (
	// Special tokens
	directive_beg iota
	IMPORT
	MIXIN
	FUNCTION
	RETURN
	INCLUDE
	CONTENT
	EXTEND
	ATIF
	ELSE
	IF
	FOR
	FROM
	TO
	THROUGH
	EACH
	IN
	WHILE
	WARN
	DEFAULT
	GLOBAL
	NULL
	OPTIONAL
	directive_end

	css_beg
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
	MS
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

var tokens = [...]string{
	IMPORT:   "@import",
	MIXIN:    "@mixin",
	FUNCTION: "@function",
	RETURN:   "@return",
	INCLUDE:  "@include",
	CONTENT:  "@content",
	EXTEND:   "@extend",
	ATIF:     "@if",
	ELSE:     "@else",
	IF:       "if",
	FOR:      "@for",
	FROM:     "from",
	TO:       "to",
	THROUGH:  "through",
	EACH:     "@each",
	IN:       "in",
	WHILE:    "@while",
	WARN:     "@warn",
	DEFAULT:  "default",
	GLOBAL:   "global",
	NULL:     "null",
	OPTIONAL: "optional",

	EM:   "em",
	EX:   "ex",
	PX:   "px",
	CM:   "cm",
	MM:   "mm",
	PT:   "pt",
	PC:   "pc",
	DEG:  "deg",
	RAD:  "rad",
	GRAD: "grad",
	MS:   "ms",
	S:    "s",
	HZ:   "Hz",
	KHZ:  "kHz",

	OPERA:  "-o-",
	WEBKIT: "-webkit-",
	MOZ:    "-moz-",
	MS:     "-ms-",
	KHTML:  "-khtml-",

	CHARSET:    "@charset",
	MEDIA:      "@media",
	KEYFRAMES:  "keyframes",
	ONLY:       "only",
	RGB:        "rgb(",
	URL:        "url(",
	IMAGEURL:   "image-url(",
	IMPORTANT:  "important",
	NOT:        ":not(",
	EVEN:       "even",
	ODD:        "odd",
	PROGID:     "progid",
	EXPRESSION: "expression",
	CALC:       "calc(",
	MOZCALC:    "-moz-calc(",
	WEBKITCALC: "-webkit-calc(",
}

// String returns the string corresponding to the token tok.
// For operators, delimiters, and keywords the string is the actual
// token character sequence (e.g., for the token ADD, the string is
// "+"). For all other tokens the string corresponds to the token
// constant name (e.g. for the token IDENT, the string is "IDENT").
//
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// A set of constants for precedence-based expression parsing.
// Non-operators have lowest precedence, followed by operators
// starting with precedence 1 up to unary operators. The highest
// precedence corresponds serves as "catch-all" precedence for
// selector, indexing, and other operator and delimiter tokens.
//
const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 7
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
//
func (op Token) Precedence() int {
	switch op {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB, OR, XOR:
		return 4
	case MUL, QUO, REM, SHL, SHR, AND, AND_NOT:
		return 5
	}
	return LowestPrec
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

// Lookup maps an identifier to its keyword token or IDENT (if not a keyword).
//
func Lookup(ident string) Token {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	return IDENT
}

// Predicates

// IsDirective returns true for tokens corresponding to sass directives
func (tok Token) IsDirective() bool { return directive_beg < tok && tok < directive_end }

// IsCss returns true for tokens corresponding to css units
func (tok Token) IsCss() bool { return css_beg < tok && tok < css_end }

// IsVendor returns true for tokens corresponding for css vendor prefixes
func (tok Token) IsVendor() bool { return vendor_beg < tok && tok < vendor_end }

// IsCssFunc returns true for tokens corresponding to css functions
func (tok Token) IsCssFunc() bool { return cssfunc_beg < tok && tok < cssfunc_beg }
