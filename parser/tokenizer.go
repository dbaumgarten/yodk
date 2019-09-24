package parser

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

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
)

type Position struct {
	Line    int
	Coloumn int
}

func NewPosition(line int, coloumn int) Position {
	return Position{
		Line:    line,
		Coloumn: coloumn,
	}
}

func (p Position) String() string {
	return fmt.Sprintf("Line: %d, Coloumn: %d", p.Line, p.Coloumn)
}
func (p Position) Add(col int) Position {
	p.Coloumn += col
	return p
}

func (p Position) Sub(col int) Position {
	p.Coloumn -= col
	return p
}

var symbols = []string{"++", "--", ">=", "<=", "!=", "==", "==", "+=", "-=", "*=", "/=", "%=",
	"=", ">", "<", "+", "-", "*", "/", "^", "%", ",", "(", ")"}

var KeywordRegex = regexp.MustCompile("^\\b(if|else|end|then|goto|and|or|not)\\b")

var IdentifierRegex = regexp.MustCompile("^:?[a-zA-Z]+[a-zA-Z0-9_]*")

var NumberRegex = regexp.MustCompile("^[0-9]+(\\.[0-9]+)?")

var CommentRegex = regexp.MustCompile("^[ \\t]*\\/\\/([^\n]*)")

var whitespaceRegex = regexp.MustCompile("^[ \\t\r]+")

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

type Tokenizer struct {
	column          int
	line            int
	text            string
	remaining       []byte
	symbols         []string
	KeywordRegex    *regexp.Regexp
	IdentifierRegex *regexp.Regexp
	NumberRegex     *regexp.Regexp
	CommentRegex    *regexp.Regexp
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		symbols:         symbols,
		KeywordRegex:    KeywordRegex,
		IdentifierRegex: IdentifierRegex,
		NumberRegex:     NumberRegex,
		CommentRegex:    CommentRegex,
	}
}

func (t *Tokenizer) newToken(typ string, val string) *Token {
	return &Token{
		Type:  typ,
		Value: val,
		Position: Position{
			Line:    t.line,
			Coloumn: t.column,
		},
	}
}

func (t *Tokenizer) Load(input string) {
	t.column = 1
	t.text = input
	t.remaining = []byte(strings.ToLower(input))
	t.line = 1
}

func (t *Tokenizer) Next() (*Token, error) {

	t.getComment()

	// no need to tokenize an empty string
	if len(t.remaining) == 0 {
		return t.newToken(TypeEOF, ""), nil
	}

	token := t.getWhitespace()
	if token != nil {
		return token, nil
	}

	token = t.getKeyword()
	if token != nil {
		return token, nil
	}

	token = t.getNewline()
	if token != nil {
		return token, nil
	}

	token = t.getSymbol()
	if token != nil {
		return token, nil
	}

	token = t.getIdentifier()
	if token != nil {
		return token, nil
	}

	token = t.getStringConstant()
	if token != nil {
		return token, nil
	}

	token = t.getNumberConstant()
	if token != nil {
		return token, nil
	}

	err := ParserError{
		Message:       fmt.Sprintf("Unknown token '%s'", string(t.remaining[0])),
		StartPosition: NewPosition(t.line, t.column),
		EndPosition:   NewPosition(t.line, t.column),
	}

	t.advance(1)

	return nil, &err
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
	for i := range symbols {
		symbol := []byte(symbols[i])
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
		tok := t.newToken(TypeKeyword, string(kw))
		return tok
	}
	return nil
}

func (t *Tokenizer) getIdentifier() *Token {
	found := t.IdentifierRegex.Find(t.remaining)
	if found != nil {
		defer t.advance(len(found))
		return t.newToken(TypeID, string(found))
	}
	return nil
}

func (t *Tokenizer) getStringConstant() *Token {
	if len(t.remaining) < 2 || t.remaining[0] != '"' {
		return nil
	}
	escaped := false
	for i, b := range t.remaining[1:] {
		if escaped {
			escaped = false
			continue
		}
		if b == '\\' {
			escaped = true
		}
		if b == '"' && !escaped {
			value := string(t.remaining[1 : i+1])
			defer t.advance(i + 2)
			return t.newToken(TypeString, value)
		}
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