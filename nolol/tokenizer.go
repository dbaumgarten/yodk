package nolol

import (
	"regexp"

	"github.com/dbaumgarten/yodk/parser"
)

func NewNololTokenizer() *parser.Tokenizer {
	tok := parser.NewTokenizer()
	tok.KeywordRegex = regexp.MustCompile("^\\b(if|else|end|then|goto|and|or|not|const)\\b")
	return tok
}
