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

// !!= is a hack, meaning "factorial followed by equal"
var symbols = []string{"!==", "++", "--", ">=", "<=", "!=", "==", "==", "+=", "-=", "*=", "/=", "%=", "^=",
	"=", ">", "<", "+", "-", "*", "/", "^", "%", ",", "(", ")", "!"}

var keywordRegex1 = regexp.MustCompile(`(?i)^(and|or|not|abs|sqrt|sin|cos|tan|asin|acos|atan)(?:[^a-zA-Z0-9_:.]|$)`)
var keywordRegex2 = regexp.MustCompile(`(?i)^(if|then|else|end|goto)`)
var keywordRegexes = []*regexp.Regexp{keywordRegex1, keywordRegex2}

var identifierRegex = regexp.MustCompile("^:[a-zA-Z0-9_:.]+|^[a-zA-Z]+[a-zA-Z0-9_.]*")

var numberRegex = regexp.MustCompile("^[0-9]+(\\.[0-9]+)?|^\\.[0-9]+")

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
	filename     string
	text         string
	remaining    []byte
	line         int
	column       int
	currentToken Token
	// List of symbols the tokenizer should recognize
	Symbols []string
	// KeywordRegexes are used to parse keywords
	KeywordRegexes []*regexp.Regexp
	// IdentifierRegex is used to parse identifiers
	IdentifierRegex *regexp.Regexp
	// NumberRegex is used to parse numbers
	NumberRegex *regexp.Regexp
	// CommentRegex is used to parse comments
	CommentRegex *regexp.Regexp
}

type TokenizerCheckpoint struct {
	Remaining []byte
	Line      int
	Column    int
}

// NewTokenizer creates a new tokenizer
func NewTokenizer() *Tokenizer {
	tk := &Tokenizer{
		Symbols:         symbols,
		KeywordRegexes:  keywordRegexes,
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

// Checkpoint returns a checkpoint that can be used to restore the Tokenizer to the current state
func (t *Tokenizer) Checkpoint() TokenizerCheckpoint {
	return TokenizerCheckpoint{
		Remaining: t.remaining,
		Line:      t.line,
		Column:    t.column,
	}
}

// Restore uses the given Checkpoint to revert the Tokenizer to a previous state
func (t *Tokenizer) Restore(cp TokenizerCheckpoint) {
	t.remaining = cp.Remaining
	t.line = cp.Line
	t.column = cp.Column
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

// Next returns the next token from the source document and advances the curent position in the input
func (t *Tokenizer) Next() *Token {
	token, size := t.getToken()
	if token.Type == TypeNewline {
		t.line++
		t.column = 0
	}
	if token.Type != TypeEOF {
		t.column += size
		t.remaining = t.remaining[size:]
	}
	return token
}

// Peek returns the next token from the source document BUT keeps the current position unchanged
func (t *Tokenizer) Peek() *Token {
	token, _ := t.getToken()
	return token
}

// getToken finds the next token in the input
// it returns a token and the length of consumed input
func (t *Tokenizer) getToken() (*Token, int) {

	token, size := t.getComment()
	if token != nil {
		return token, size
	}

	// no need to tokenize an empty string
	if len(t.remaining) == 0 {
		return t.newToken(TypeEOF, ""), 0
	}

	token, size = t.getWhitespace()
	if token != nil {
		return token, size
	}

	token, size = t.getKeyword()
	if token != nil {
		return token, size
	}

	token, size = t.getNewline()
	if token != nil {
		return token, size
	}

	token, size = t.getSymbol()
	if token != nil {
		return token, size
	}

	token, size = t.getIdentifier()
	if token != nil {
		return token, size
	}

	token, size = t.getStringConstant()
	if token != nil {
		return token, size
	}

	token, size = t.getNumberConstant()
	if token != nil {
		return token, size
	}

	return t.newToken(TypeUnknown, string(t.remaining[0])), 1
}

func (t *Tokenizer) getWhitespace() (*Token, int) {
	found := whitespaceRegex.Find(t.remaining)
	if found != nil {
		return t.newToken(TypeWhitespace, string(found)), len(found)
	}
	return nil, 0
}

func (t *Tokenizer) getNewline() (*Token, int) {
	if len(t.remaining) > 0 && t.remaining[0] == '\n' {
		token := t.newToken(TypeNewline, "")
		return token, 1
	}
	return nil, 0
}

func (t *Tokenizer) getSymbol() (*Token, int) {
	for i := range t.Symbols {
		symbol := []byte(t.Symbols[i])
		if bytes.HasPrefix(t.remaining, symbol) {
			if t.Symbols[i] == "!==" {
				// this special case is needed, as otherwise !== would be parsed as "!= =", but it shold be "! =="
				return t.newToken(TypeSymbol, "!"), 1
			}
			return t.newToken(TypeSymbol, string(symbol)), len(symbol)
		}
	}
	return nil, 0
}

func (t *Tokenizer) getComment() (*Token, int) {
	found := t.CommentRegex.Find(t.remaining)
	if found != nil {
		return t.newToken(TypeComment, string(found)), len(found)
	}
	return nil, 0
}

func (t *Tokenizer) getKeyword() (*Token, int) {
	for _, regex := range t.KeywordRegexes {
		found := regex.FindSubmatch(t.remaining)
		if found != nil {
			kw := found[1]
			// keywords are always treated as lowercase
			tok := t.newToken(TypeKeyword, strings.ToLower(string(kw)))
			return tok, len(found[1])
		}
	}
	return nil, 0
}

func (t *Tokenizer) getIdentifier() (*Token, int) {
	found := t.IdentifierRegex.Find(t.remaining)
	if found != nil {
		// do not convert value to lowercase.
		// the parser deals with casing
		return t.newToken(TypeID, string(found)), len(found)
	}
	return nil, 0
}

func (t *Tokenizer) getStringConstant() (*Token, int) {
	if len(t.remaining) < 2 || t.remaining[0] != '"' {
		return nil, 0
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
			return t.newToken(TypeString, str), i + 2
		}
		str += string(b)
	}
	return nil, 0
}

func (t *Tokenizer) getNumberConstant() (*Token, int) {
	found := t.NumberRegex.Find(t.remaining)
	if found != nil {
		return t.newToken(TypeNumber, string(found)), len(found)
	}
	return nil, 0
}
