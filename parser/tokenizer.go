package parser

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

const (
	TypeID      = "ID"
	TypeNumber  = "Number"
	TypeString  = "String"
	TypeKeyword = "Keyword"
	TypeSymbol  = "Symbol"
	TypeNewline = "Newline"
	TypeEOF     = "EOF"
	TypeComment = "Comment"
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

var keywordRegex = regexp.MustCompile("^[ \t]*(if | else | end| then |goto | and | or | not )")

var identifierRegex = regexp.MustCompile("^:?[a-zA-Z]+[a-zA-Z0-9_]*")

var numberRegex = regexp.MustCompile("^[0-9]+(\\.[0-9]+)?")

var commentRegex = regexp.MustCompile("^[ \\t]*\\/\\/([^\n]*)")

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
	keywordRegex    *regexp.Regexp
	identifierRegex *regexp.Regexp
	numberRegex     *regexp.Regexp
	commentRegex    *regexp.Regexp
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		symbols:         symbols,
		keywordRegex:    keywordRegex,
		identifierRegex: identifierRegex,
		numberRegex:     numberRegex,
		commentRegex:    commentRegex,
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
	t.column = 0
	t.text = input
	t.remaining = []byte(strings.ToLower(input))
	t.line = 1
}

func (t *Tokenizer) Next() (*Token, error) {

	// no need to tokenize an empty string
	if len(t.remaining) == 0 {
		return t.newToken(TypeEOF, ""), nil
	}

	// try to get a comment, but silently discard it if found
	t.getComment()

	//searching for keywords must happen before trimming
	token := t.getKeyword()
	if token != nil {
		return token, nil
	}

	t.trim()

	//did the trimming result in an empty string?
	if len(t.remaining) == 0 {
		return t.newToken(TypeEOF, ""), nil
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

	return nil, &err
}

func (t *Tokenizer) advance(amount int) {
	t.column += amount
	t.remaining = t.remaining[amount:]
}

func (t *Tokenizer) trim() bool {
	counter := 0
	for _, b := range t.remaining {
		if b == ' ' || b == '\t' || b == '\r' {
			counter++
			continue
		}
		break
	}
	if counter > 0 {
		t.advance(counter)
		return true
	}
	return false
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
	found := commentRegex.Find(t.remaining)
	if found != nil {
		defer t.advance(len(found))
		return t.newToken(TypeComment, string(found))
	}
	return nil
}

func countLeadingSpace(line string) int {
	i := 0
	for _, runeValue := range line {
		if runeValue == ' ' {
			i++
		} else {
			break
		}
	}
	return i
}

func (t *Tokenizer) getKeyword() *Token {
	found := keywordRegex.Find(t.remaining)
	if found != nil {
		defer t.advance(len(found))
		kw := bytes.Trim(found, " \t\n")
		tok := t.newToken(TypeKeyword, string(kw))
		// the keyword-regex matches may contain leading spaces.
		// make sure the position is still correct
		tok.Position.Coloumn += countLeadingSpace(string(found))
		return tok
	}
	return nil
}

func (t *Tokenizer) getIdentifier() *Token {
	found := identifierRegex.Find(t.remaining)
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
	found := numberRegex.Find(t.remaining)
	if found != nil {
		defer t.advance(len(found))
		return t.newToken(TypeNumber, string(found))
	}
	return nil
}
