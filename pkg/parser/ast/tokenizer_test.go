package ast_test

import (
	"strings"
	"testing"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

func TestTokenizer(t *testing.T) {
	input := `
	test = 123 def=245
	abc = test + 1
	hallo = "welt" //comment1
	if hallo == welt then bla=blub+1 else goto 1 end
	i++
	a = 0 and 1
	// comment 2
	:var = "another\"test"
	:foo = (1+2.75)*3
	`

	expected := `Line: 1, Coloumn: 1, Type: Newline
	Line: 2, Coloumn: 1, Type: Whitespace, Value: '	'
	Line: 2, Coloumn: 2, Type: ID, Value: 'test'
	Line: 2, Coloumn: 6, Type: Whitespace, Value: ' '
	Line: 2, Coloumn: 7, Type: Symbol, Value: '='
	Line: 2, Coloumn: 8, Type: Whitespace, Value: ' '
	Line: 2, Coloumn: 9, Type: Number, Value: '123'
	Line: 2, Coloumn: 12, Type: Whitespace, Value: ' '
	Line: 2, Coloumn: 13, Type: ID, Value: 'def'
	Line: 2, Coloumn: 16, Type: Symbol, Value: '='
	Line: 2, Coloumn: 17, Type: Number, Value: '245'
	Line: 2, Coloumn: 20, Type: Newline
	Line: 3, Coloumn: 1, Type: Whitespace, Value: '	'
	Line: 3, Coloumn: 2, Type: ID, Value: 'abc'
	Line: 3, Coloumn: 5, Type: Whitespace, Value: ' '
	Line: 3, Coloumn: 6, Type: Symbol, Value: '='
	Line: 3, Coloumn: 7, Type: Whitespace, Value: ' '
	Line: 3, Coloumn: 8, Type: ID, Value: 'test'
	Line: 3, Coloumn: 12, Type: Whitespace, Value: ' '
	Line: 3, Coloumn: 13, Type: Symbol, Value: '+'
	Line: 3, Coloumn: 14, Type: Whitespace, Value: ' '
	Line: 3, Coloumn: 15, Type: Number, Value: '1'
	Line: 3, Coloumn: 16, Type: Newline
	Line: 4, Coloumn: 1, Type: Whitespace, Value: '	'
	Line: 4, Coloumn: 2, Type: ID, Value: 'hallo'
	Line: 4, Coloumn: 7, Type: Whitespace, Value: ' '
	Line: 4, Coloumn: 8, Type: Symbol, Value: '='
	Line: 4, Coloumn: 9, Type: Whitespace, Value: ' '
	Line: 4, Coloumn: 10, Type: String, Value: 'welt'
	Line: 4, Coloumn: 16, Type: Comment, Value: ' //comment1'
	Line: 4, Coloumn: 27, Type: Newline
	Line: 5, Coloumn: 1, Type: Whitespace, Value: '	'
	Line: 5, Coloumn: 2, Type: Keyword, Value: 'if'
	Line: 5, Coloumn: 4, Type: Whitespace, Value: ' '
	Line: 5, Coloumn: 5, Type: ID, Value: 'hallo'
	Line: 5, Coloumn: 10, Type: Whitespace, Value: ' '
	Line: 5, Coloumn: 11, Type: Symbol, Value: '=='
	Line: 5, Coloumn: 13, Type: Whitespace, Value: ' '
	Line: 5, Coloumn: 14, Type: ID, Value: 'welt'
	Line: 5, Coloumn: 18, Type: Whitespace, Value: ' '
	Line: 5, Coloumn: 19, Type: Keyword, Value: 'then'
	Line: 5, Coloumn: 23, Type: Whitespace, Value: ' '
	Line: 5, Coloumn: 24, Type: ID, Value: 'bla'
	Line: 5, Coloumn: 27, Type: Symbol, Value: '='
	Line: 5, Coloumn: 28, Type: ID, Value: 'blub'
	Line: 5, Coloumn: 32, Type: Symbol, Value: '+'
	Line: 5, Coloumn: 33, Type: Number, Value: '1'
	Line: 5, Coloumn: 34, Type: Whitespace, Value: ' '
	Line: 5, Coloumn: 35, Type: Keyword, Value: 'else'
	Line: 5, Coloumn: 39, Type: Whitespace, Value: ' '
	Line: 5, Coloumn: 40, Type: Keyword, Value: 'goto'
	Line: 5, Coloumn: 44, Type: Whitespace, Value: ' '
	Line: 5, Coloumn: 45, Type: Number, Value: '1'
	Line: 5, Coloumn: 46, Type: Whitespace, Value: ' '
	Line: 5, Coloumn: 47, Type: Keyword, Value: 'end'
	Line: 5, Coloumn: 50, Type: Newline
	Line: 6, Coloumn: 1, Type: Whitespace, Value: '	'
	Line: 6, Coloumn: 2, Type: ID, Value: 'i'
	Line: 6, Coloumn: 3, Type: Symbol, Value: '++'
	Line: 6, Coloumn: 5, Type: Newline
	Line: 7, Coloumn: 1, Type: Whitespace, Value: '	'
	Line: 7, Coloumn: 2, Type: ID, Value: 'a'
	Line: 7, Coloumn: 3, Type: Whitespace, Value: ' '
	Line: 7, Coloumn: 4, Type: Symbol, Value: '='
	Line: 7, Coloumn: 5, Type: Whitespace, Value: ' '
	Line: 7, Coloumn: 6, Type: Number, Value: '0'
	Line: 7, Coloumn: 7, Type: Whitespace, Value: ' '
	Line: 7, Coloumn: 8, Type: Keyword, Value: 'and'
	Line: 7, Coloumn: 11, Type: Whitespace, Value: ' '
	Line: 7, Coloumn: 12, Type: Number, Value: '1'
	Line: 7, Coloumn: 13, Type: Newline
	Line: 8, Coloumn: 1, Type: Comment, Value: '	// comment 2'
	Line: 8, Coloumn: 14, Type: Newline
	Line: 9, Coloumn: 1, Type: Whitespace, Value: '	'
	Line: 9, Coloumn: 2, Type: ID, Value: ':var'
	Line: 9, Coloumn: 6, Type: Whitespace, Value: ' '
	Line: 9, Coloumn: 7, Type: Symbol, Value: '='
	Line: 9, Coloumn: 8, Type: Whitespace, Value: ' '
	Line: 9, Coloumn: 9, Type: String, Value: 'another\"test'
	Line: 9, Coloumn: 24, Type: Newline
	Line: 10, Coloumn: 1, Type: Whitespace, Value: '	'
	Line: 10, Coloumn: 2, Type: ID, Value: ':foo'
	Line: 10, Coloumn: 6, Type: Whitespace, Value: ' '
	Line: 10, Coloumn: 7, Type: Symbol, Value: '='
	Line: 10, Coloumn: 8, Type: Whitespace, Value: ' '
	Line: 10, Coloumn: 9, Type: Symbol, Value: '('
	Line: 10, Coloumn: 10, Type: Number, Value: '1'
	Line: 10, Coloumn: 11, Type: Symbol, Value: '+'
	Line: 10, Coloumn: 12, Type: Number, Value: '2.75'
	Line: 10, Coloumn: 16, Type: Symbol, Value: ')'
	Line: 10, Coloumn: 17, Type: Symbol, Value: '*'
	Line: 10, Coloumn: 18, Type: Number, Value: '3'
	Line: 10, Coloumn: 19, Type: Newline
	Line: 11, Coloumn: 1, Type: Whitespace, Value: '	'
	`

	tk := ast.NewTokenizer()
	tk.Load(input)

	output := ""

	for {
		token := tk.Next()
		if token.Type == ast.TypeEOF {
			break
		}
		output += token.String()
	}

	expectedLines := strings.Split(expected, "\n")
	outputLines := strings.Split(output, "\n")

	if len(expectedLines) < len(outputLines) {
		t.Log(output)
		t.Fatalf("To many tokens found (%d)", len(outputLines)-len(expectedLines))
	}

	if len(expectedLines) > len(outputLines) {
		t.Log(output)
		t.Fatalf("Not enough tokens found (%d)", len(expectedLines)-len(outputLines))
	}

	for i, line := range expectedLines {
		if strings.Trim(line, " \t\n") != strings.Trim(outputLines[i], " \t\n") {
			t.Fatalf("Unexpected token, wanted:\n'%s'\nbut found: \n'%s'\n", line, outputLines[i])
		}
	}
}
