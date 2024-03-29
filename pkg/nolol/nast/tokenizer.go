package nast

import (
	"regexp"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// NewNololTokenizer creates a Yolol-Tokenizer that is modified to also accept Nolol-specific tokens
func NewNololTokenizer() *ast.Tokenizer {
	tok := ast.NewTokenizer()
	tok.KeywordRegexes = []*regexp.Regexp{regexp.MustCompile("(?i)^\\b(if|else|end|then|goto|and|or|not|define|while|do|wait|include|macro|insert|break|continue|block|line|expr)\\b")}
	tok.Symbols = append(tok.Symbols, []string{";", "$"}...)
	return tok
}
