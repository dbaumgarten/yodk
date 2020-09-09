package ast

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// Defines the different types a token can be
const (
	TypeID         = "ID"
	TypeNumber     = "Number"
	TypeString     = "String"
	TypeKeyword    = "Keyword"
	TypeSymbol     = "Symbol"
	TypeNewline    = "Newline"
	TypeEOF        = "EOF"
	TypeComment    = "Comment"
	TypeWhitespace = "Whitespace"
	TypeUnknown    = "Unknown"
)

// Position represents the starting-position of a token in the source-code
type Position struct {
	File    string
	Line    int
	Coloumn int
}

// NewPosition creates a new position from a given line and coloumn
func NewPosition(file string, line int, coloumn int) Position {
	return Position{
		File:    file,
		Line:    line,
		Coloumn: coloumn,
	}
}

// UnknownPosition is used, when a position is expected, but a real one can not be provided
// Usually used by ast-elements that rely on their children to determin their position, if the children are nil
var UnknownPosition = Position{}

func (p Position) String() string {
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Coloumn)
	}
	return fmt.Sprintf("Line: %d, Coloumn: %d", p.Line, p.Coloumn)
}

// Add creates a new position from the old one and adds the given amount of coloumns
func (p Position) Add(col int) Position {
	p.Coloumn += col
	return p
}

// Before returns true if p represents a position in the file before the position of other
func (p Position) Before(other Position) bool {
	if p.Line < other.Line {
		return true
	}
	if p.Line == other.Line && p.Coloumn < other.Coloumn {
		return true
	}
	return false
}

var symbols = []string{"++", "--", ">=", "<=", "!=", "==", "==", "+=", "-=", "*=", "/=", "%=",
	"=", ">", "<", "+", "-", "*", "/", "^", "%", ",", "(", ")"}

var keywordRegex = regexp.MustCompile(`(?i)^(if|else\b|end\b|then|goto|and|or|not|abs|sqrt|sin|cos|tan|asin|acos|atan)`)

var identifierRegex = regexp.MustCompile("^:?[a-zA-Z]+[a-zA-Z0-9_]*")

var numberRegex = regexp.MustCompile("^[0-9]+(\\.[0-9]+)?")

var commentRegex = regexp.MustCompile("^\\/\\/([^\n]*)")

var whitespaceRegex = regexp.MustCompile("^[ \\t\r]+")

// Token represents a token fount in the source-code
type Token struct {
	Type     string
	Value    string
	Position Position
}

func (t Token) String() string {
	str := fmt.Sprintf("%s, Type: %s", t.Position.String(), t.Type)
	if t.Value != "" {
		str += ", Value: '" + t.Value + "'"
	}
	str += "\n"
	return str
}

// Tokenizer splits the input source-code into tokens
type Tokenizer struct {
	filename  string
	column    int
	line      int
	text      string
	remaining []byte
	Symbols   []string
	// KeywordRegex is used to parse keywords
	KeywordRegex *regexp.Regexp
	// IdentifierRegex is used to parse identifiers
	IdentifierRegex *regexp.Regexp
	// NumberRegex is used to parse numbers
	NumberRegex *regexp.Regexp
	// CommentRegex is used to parse comments
	CommentRegex *regexp.Regexp
}

// NewTokenizer creates a new tokenizer
func NewTokenizer() *Tokenizer {
	tk := &Tokenizer{
		Symbols:         symbols,
		KeywordRegex:    keywordRegex,
		IdentifierRegex: identifierRegex,
		NumberRegex:     numberRegex,
		CommentRegex:    commentRegex,
	}
	return tk
}

// SetFilename sets the filename that is set in the position if all returned tokens
func (t *Tokenizer) SetFilename(name string) {
	t.filename = name
}

func (t *Tokenizer) newToken(typ string, val string) *Token {
	return &Token{
		Type:  typ,
		Value: val,
		Position: Position{
			File:    t.filename,
			Line:    t.line,
			Coloumn: t.column,
		},
	}
}

// Load loads programm code as input
func (t *Tokenizer) Load(input string) {
	t.column = 1
	t.text = input
	t.remaining = []byte(input)
	t.line = 1
}

// Next returns the next token from the source document
func (t *Tokenizer) Next() *Token {

	token := t.getComment()
	if token != nil {
		return token
	}

	// no need to tokenize an empty string
	if len(t.remaining) == 0 {
		return t.newToken(TypeEOF, "")
	}

	token = t.getWhitespace()
	if token != nil {
		return token
	}

	token = t.getKeyword()
	if token != nil {
		return token
	}

	token = t.getNewline()
	if token != nil {
		return token
	}

	token = t.getSymbol()
	if token != nil {
		return token
	}

	token = t.getIdentifier()
	if token != nil {
		return token
	}

	token = t.getStringConstant()
	if token != nil {
		return token
	}

	token = t.getNumberConstant()
	if token != nil {
		return token
	}

	token = t.newToken(TypeUnknown, string(t.remaining[0]))
	t.advance(1)

	return token
}

func (t *Tokenizer) advance(amount int) {
	t.column += amount
	t.remaining = t.remaining[amount:]
}

func (t *Tokenizer) getWhitespace() *Token {
	found := whitespaceRegex.Find(t.remaining)
	if found != nil {
		defer t.advance(len(found))
		return t.newToken(TypeWhitespace, string(found))
	}
	return nil
}

func (t *Tokenizer) getNewline() *Token {
	if len(t.remaining) > 0 && t.remaining[0] == '\n' {
		defer func() {
			t.line++
			t.column = 0
			t.advance(1)
		}()
		return t.newToken(TypeNewline, "")
	}
	return nil
}

func (t *Tokenizer) getSymbol() *Token {
	for i := range t.Symbols {
		symbol := []byte(t.Symbols[i])
		if bytes.HasPrefix(t.remaining, symbol) {
			defer t.advance(len(symbol))
			return t.newToken(TypeSymbol, string(symbol))
		}
	}
	return nil
}

func (t *Tokenizer) getComment() *Token {
	found := t.CommentRegex.Find(t.remaining)
	if found != nil {
		defer t.advance(len(found))
		return t.newToken(TypeComment, string(found))
	}
	return nil
}

func (t *Tokenizer) getKeyword() *Token {
	found := t.KeywordRegex.FindSubmatch(t.remaining)
	if found != nil {
		defer t.advance(len(found[0]))
		kw := found[1]
		// keywords are always treated as lowercase
		tok := t.newToken(TypeKeyword, strings.ToLower(string(kw)))
		return tok
	}
	return nil
}

func (t *Tokenizer) getIdentifier() *Token {
	found := t.IdentifierRegex.Find(t.remaining)
	if found != nil {
		defer t.advance(len(found))
		// do not convert value to lowercase.
		// the parser deals with casing
		return t.newToken(TypeID, string(found))
	}
	return nil
}

func (t *Tokenizer) getStringConstant() *Token {
	if len(t.remaining) < 2 || t.remaining[0] != '"' {
		return nil
	}
	escaped := false
	str := ""
	for i, b := range t.remaining[1:] {
		if b == '\\' {
			escaped = true
			continue
		}
		if escaped {
			switch b {
			case 'n':
				str += "\n"
				escaped = false
				continue
			case 't':
				str += "\t"
				escaped = false
				continue
			case '"':
				str += "\""
				escaped = false
				continue
			}
		}

		if b == '"' {
			defer t.advance(i + 2)
			return t.newToken(TypeString, str)
		}
		str += string(b)
	}
	return nil
}

func (t *Tokenizer) getNumberConstant() *Token {
	found := t.NumberRegex.Find(t.remaining)
	if found != nil {
		defer t.advance(len(found))
		return t.newToken(TypeNumber, string(found))
	}
	return nil
}
