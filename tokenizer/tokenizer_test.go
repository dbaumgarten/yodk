package tokenizer

import (
	"strings"
	"testing"
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

	expected := `Line 1, Col:0, Type: Newline
	Line 2, Col:2, Type: ID, Value: 'test'
	Line 2, Col:7, Type: Symbol, Value: '='
	Line 2, Col:9, Type: Number, Value: '123'
	Line 2, Col:13, Type: ID, Value: 'def'
	Line 2, Col:16, Type: Symbol, Value: '='
	Line 2, Col:17, Type: Number, Value: '245'
	Line 2, Col:20, Type: Newline
	Line 3, Col:2, Type: ID, Value: 'abc'
	Line 3, Col:6, Type: Symbol, Value: '='
	Line 3, Col:8, Type: ID, Value: 'test'
	Line 3, Col:13, Type: Symbol, Value: '+'
	Line 3, Col:15, Type: Number, Value: '1'
	Line 3, Col:16, Type: Newline
	Line 4, Col:2, Type: ID, Value: 'hallo'
	Line 4, Col:8, Type: Symbol, Value: '='
	Line 4, Col:10, Type: String, Value: 'welt'
	Line 4, Col:27, Type: Newline
	Line 5, Col:1, Type: Keyword, Value: 'if'
	Line 5, Col:5, Type: ID, Value: 'hallo'
	Line 5, Col:11, Type: Symbol, Value: '=='
	Line 5, Col:14, Type: ID, Value: 'welt'
	Line 5, Col:18, Type: Keyword, Value: 'then'
	Line 5, Col:24, Type: ID, Value: 'bla'
	Line 5, Col:27, Type: Symbol, Value: '='
	Line 5, Col:28, Type: ID, Value: 'blub'
	Line 5, Col:32, Type: Symbol, Value: '+'
	Line 5, Col:33, Type: Number, Value: '1'
	Line 5, Col:34, Type: Keyword, Value: 'else'
	Line 5, Col:40, Type: Keyword, Value: 'goto'
	Line 5, Col:45, Type: Number, Value: '1'
	Line 5, Col:46, Type: Keyword, Value: 'end'
	Line 5, Col:50, Type: Newline
	Line 6, Col:2, Type: ID, Value: 'i'
	Line 6, Col:3, Type: Symbol, Value: '++'
	Line 6, Col:5, Type: Newline
	Line 7, Col:2, Type: ID, Value: 'a'
	Line 7, Col:4, Type: Symbol, Value: '='
	Line 7, Col:6, Type: Number, Value: '0'
	Line 7, Col:7, Type: Keyword, Value: 'and'
	Line 7, Col:12, Type: Number, Value: '1'
	Line 7, Col:13, Type: Newline
	Line 8, Col:14, Type: Newline
	Line 9, Col:2, Type: ID, Value: ':var'
	Line 9, Col:7, Type: Symbol, Value: '='
	Line 9, Col:9, Type: String, Value: 'another\"test'
	Line 9, Col:24, Type: Newline
	Line 10, Col:2, Type: ID, Value: ':foo'
	Line 10, Col:7, Type: Symbol, Value: '='
	Line 10, Col:9, Type: Symbol, Value: '('
	Line 10, Col:10, Type: Number, Value: '1'
	Line 10, Col:11, Type: Symbol, Value: '+'
	Line 10, Col:12, Type: Number, Value: '2.75'
	Line 10, Col:16, Type: Symbol, Value: ')'
	Line 10, Col:17, Type: Symbol, Value: '*'
	Line 10, Col:18, Type: Number, Value: '3'
	Line 10, Col:19, Type: Newline
	`

	tk := Tokenizer{}
	tk.Load(input)

	output := ""

	for {
		token, err := tk.Next()
		if err != nil {
			t.Fatal("Error when tokenizing:", err)
		}
		if token.Type == TypeEOF {
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
			t.Fatalf("Unexpected token, wanted '%s' but found '%s", line, outputLines[i])
		}
	}
}
