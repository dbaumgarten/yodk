package nolol

import (
	"regexp"

	"github.com/dbaumgarten/yodk/pkg/parser"
)

// NewNololTokenizer creates a Yolol-Tokenizer that is modified to also accept Nolol-specific tokens
func NewNololTokenizer() *parser.Tokenizer {
	tok := parser.NewTokenizer()
	tok.KeywordRegex = regexp.MustCompile("^\\b(if|else|end|then|goto|and|or|not|const|while|do)\\b")
	return tok
}
