/*
Two APIs

The Lexer type has two APIs, one is used byte StateFn types.  The other is
called by the parser. These APIs are called the scanner and the parser APIs
here.

The parser API

The only function the parser calls on the lexer is Next to retreive the next
token from the input stream.  Eventually an item with type ItemEOF is returned
at which point there are no more tokens in the stream.

The scanner API

The lexer uses Emit to construct complete lexemes to return from
future/concurrent calls to Next by the parser.  The scanner uses a combination
of methods to manipulate its position and and prepare lexemes to be emitted.
Lexer errors are emitted to the parser using the Errorf method which keeps the
scanner-parser interface uniform.

Common lexer methods used in a scanner are the Accept[Run][Range]
family of methods.  Accept* methods take a set and advance the
lexer if incoming runes are in the set. The AcceptRun* subfamily
advance the lexer as far as possible.

For scanning known sequences of bytes (e.g. keywords) the
AcceptString method avoids a lot of branching that would be
incurred using methods that match character classes.

The remaining methods provide low level functionality that can
be combined to address corner cases.
*/
package sprite_sass

import (
	"container/list"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

const EOF rune = 0x04

// IsEOF returns true if n is zero.
func IsEOF(c rune, n int) bool {
	return n == 0
}

// IsInvalid returns true if c is utf8.RuneError and n is 1.
func IsInvalid(c rune, n int) bool {
	return c == utf8.RuneError && n == 1
}

// StateFn functions scan runes from the lexer's input and emit items.  A StateFn
// is responsible for emitting ItemEOF after input has been consumed.
type StateFn func(*Lexer) StateFn

// Lexer contains an input string and state associate with the lexing the
// input.
type Lexer struct {
	input string     // string being scanned
	start int        // start position for the current lexeme
	pos   int        // current position
	width int        // length of the last rune read
	last  rune       // the last rune read
	state StateFn    // the current state
	items *list.List // Buffer of lexed items
}

// Create a new lexer. Must be given a non-nil state.
func New(start StateFn, input string) *Lexer {
	if start == nil {
		panic("nil start state")
	}
	return &Lexer{
		state: start,
		input: input,
		items: list.New(),
	}
}

// Input returns the input string being lexed by the l.
func (l *Lexer) Input() string {
	return l.input
}

// Start marks the first byte of item currently being lexed.
func (l *Lexer) Start() int {
	return l.start
}

// Pos marks the next byte to be read in the input string.  The behavior of Pos
// is unspecified if an error previously occurred or if all input has been
// consumed.
func (l *Lexer) Pos() int {
	return l.pos
}

// Current returns the contents of the item currently being lexed.
func (l *Lexer) Current() string {
	return l.input[l.start:l.pos]
}

// Last return the last rune read from the input stream.
func (l *Lexer) Last() (r rune, width int) {
	return l.last, l.width
}

// Advance adds one rune of input to the current lexeme,
// increments the lexer's position, and returns the input
// rune with its size in bytes (encoded as UTF-8).  Invalid
// UTF-8 codepoints cause the current call and all subsequent
// calls to return (utf8.RuneError, 1).  If there is no input
// the returned size is zero.
func (l *Lexer) Advance() (rune, int) {
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF, l.width
	}
	l.last, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	if l.last == utf8.RuneError && l.width == 1 {
		return l.last, l.width
	}
	l.pos += l.width
	return l.last, l.width
}

// Backup removes the last rune from the current lexeme and moves l's position
// back in the input string accordingly. Backup should only be called after a
// call to Advance.
func (l *Lexer) Backup() {
	l.pos -= l.width
}

// Peek returns the next rune in the input stream without adding it to the
// current lexeme.
func (l *Lexer) Peek() (rune, int) {
	c, n := l.Advance()
	l.Backup()
	return c, n
}

// Ignore throws away the current lexeme.
func (l *Lexer) Ignore() {
	l.start = l.pos
}

// Accept advances the lexer if the next rune is in valid.
func (l *Lexer) Accept(valid string) (ok bool) {
	r, _ := l.Advance()
	ok = strings.IndexRune(valid, r) >= 0
	if !ok {
		l.Backup()
	}
	return
}

// AcceptFunc advances the lexer if fn return true for the next rune.
func (l *Lexer) AcceptFunc(fn func(rune) bool) (ok bool) {
	switch r, n := l.Advance(); {
	case IsEOF(r, n):
		return false
	case IsInvalid(r, n):
		return false
	case fn(r):
		return true
	default:
		l.Backup()
		return false
	}
}

// AcceptRange advances l's position if the current rune is in tab.
func (l *Lexer) AcceptRange(tab *unicode.RangeTable) (ok bool) {
	r, _ := l.Advance()
	ok = unicode.Is(tab, r)
	if !ok {
		l.Backup()
	}
	return
}

// AcceptRun advances l's position as long as the current rune is in valid.
func (l *Lexer) AcceptRun(valid string) (n int) {
	for l.Accept(valid) {
		n++
	}
	return
}

// AcceptRunFunc advances l's position as long as fn returns
// true for the next input rune.
func (l *Lexer) AcceptRunFunc(fn func(rune) bool) int {
	var n int
	for l.AcceptFunc(fn) {
		n++
	}
	return n
}

// AcceptRunRange advances l's possition as long as the current
// rune is in tab.
func (l *Lexer) AcceptRunRange(tab *unicode.RangeTable) (n int) {
	for l.AcceptRange(tab) {
		n++
	}
	return
}

// AcceptString advances the lexer len(s) bytes if the next
// len(s) bytes equal s. AcceptString returns true if l advanced.
func (l *Lexer) AcceptString(s string) (ok bool) {
	if strings.HasPrefix(l.input[l.pos:], s) {
		l.pos += len(s)
		return true
	}
	return false
}

// Errorf causes an error item to be emitted from l.Next().  The item's value
// (and its error message) are the result of evaluating format and vs with
// fmt.Sprintf.
func (l *Lexer) Errorf(format string, vs ...interface{}) StateFn {
	l.enqueue(&Item{
		ItemError,
		l.start,
		fmt.Sprintf(format, vs...),
	})
	return nil
}

// Emit the current value as an Item with the specified type.
func (l *Lexer) Emit(t ItemType) {
	l.enqueue(&Item{
		t,
		l.start,
		l.input[l.start:l.pos],
	})
	l.start = l.pos
}

// The method by which items are extracted from the input.
// Returns nil if the lexer has entered a nil state.
func (l *Lexer) Next() (i *Item) {
	for {
		if head := l.dequeue(); head != nil {
			return head
		}
		if l.state == nil {
			return &Item{ItemEOF, l.start, ""}
		}
		l.state = l.state(l)
	}
	panic("unreachable")
}

func (l *Lexer) enqueue(i *Item) {
	l.items.PushBack(i)
}

func (l *Lexer) dequeue() *Item {
	head := l.items.Front()
	if head == nil {
		return nil
	}
	return l.items.Remove(head).(*Item)
}

// A type for all the types of items in the language being lexed.
type ItemType uint16

// Special item types.
const (
	ItemEOF ItemType = math.MaxUint16 - iota
	ItemError
	IMPORT
	EXTRA
	CMD
	VAR
	SUB
	VALUE
	FILE
	SPRITE
	NUMBER
	TEXT
	DOLLAR
	special_beg
	LPAREN
	RPAREN
	LBRACKET
	RBRACKET
	special_end
)

var Tokens = [...]string{
	ItemEOF:   "eof",
	ItemError: "error",
	IMPORT:    "@import",
	EXTRA:     "extra",
	CMD:       "command",
	VAR:       "variable",
	SUB:       "sub",
	VALUE:     "value",
	FILE:      "file",
	SPRITE:    "sprite",
	NUMBER:    "number",
	TEXT:      "text",
	DOLLAR:    "$",
	LPAREN:    "(",
	RPAREN:    ")",
	LBRACKET:  "{",
	RBRACKET:  "}",
}

const (
	Symbols = `/\.*-_`
)

func IsAllowedRune(r rune) bool {
	return unicode.IsNumber(r) ||
		unicode.IsLetter(r) ||
		strings.ContainsRune(Symbols, r)
}

func (i ItemType) String() string {
	return Tokens[i]
}

// An individual scanned item (a lexeme).
type Item struct {
	Type  ItemType
	Pos   int
	Value string
}

func (i Item) Error() error {
	if i.Type == ItemError {
		return errors.New("Error reading input")
	}
	return nil
}

// String returns the raw lexeme of i.
func (i Item) String() string {
	switch i.Type {
	case ItemError:
		return i.Value
	case ItemEOF:
		return "EOF"
	}
	if len(i.Value) > 10 {
		return fmt.Sprintf("%s", i.Value)
	}
	return i.Value
}

func (l *Lexer) Action() StateFn {
	for {
		switch r, _ := l.Advance(); {
		case r == EOF: // || r == '\n':
			l.enqueue(&Item{
				ItemEOF,
				l.start,
				"",
			})
			return nil
		case IsSpace(r):
			l.Ignore()
		case IsSymbol(r):
			return l.Paren()
		case r == '@':
			return l.Directive()
		case r == '"':
			return l.File()
		case r == '$':
			return l.Var()
		case IsAllowedRune(r):
			return l.Text()
		default:
			//l.Advance()
			//l.Emit(EXTRA)
		}
	}
}

func IsSymbol(r rune) bool {
	return strings.ContainsRune("(),;{}", r)
}

func IsSpace(r rune) bool {
	return strings.ContainsRune(" \n", r)
}

func IsPrintable(r rune) bool {
	return true
}

func (l *Lexer) Directive() StateFn {

	l.AcceptRunFunc(IsAllowedRune)
	l.Emit(IMPORT)

	switch l.Current() {
	case "@import":

	}
	return l.Action()
}

func (l *Lexer) Paren() StateFn {
	switch l.Current() {
	case "(":
		l.Emit(LPAREN)
	case ")":
		l.Emit(RPAREN)
	case "{":
		l.Emit(LBRACKET)
	case "}":
		l.Emit(RBRACKET)
	}
	return l.Action()
}

// $images: sprite-map("*.png");
func (l *Lexer) Var() StateFn {
	l.Accept("$")
	l.AcceptRunFunc(IsAllowedRune)
	r, _ := l.Peek()

	if r == ':' {
		l.Emit(VAR)
	} else {
		l.Emit(SUB)
	}
	return l.Action()
}

func (l *Lexer) Text() StateFn {
	l.AcceptRunFunc(IsAllowedRune)

	switch l.Current() {
	//Primary support
	case "sprite", "sprite-file", "sprite-height",
		"sprite-map", "sprite-path", "sprite-position",
		"sprite-width", "sprite-url":
		l.Emit(CMD)
		return l.Action()
	//Tertiary support
	case "sprite-map-name", "sprite-names":
		l.Emit(CMD)
		return l.Action()
	default:
		l.Emit(TEXT)
	}
	return l.Action()
}

func (l *Lexer) File() StateFn {
	l.Ignore()
	if !l.AcceptFunc(IsAllowedRune) {
		return l.Action()
	}
	l.AcceptRunFunc(IsAllowedRune)
	l.Emit(FILE)
	return l.Action()
}
