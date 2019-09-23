package parser

import "regexp"

func NewNololTokenizer() *Tokenizer {
	tok := NewTokenizer()
	tok.keywordRegex = regexp.MustCompile("^\\b(if|else|end|then|goto|and|or|not|const)\\b")
	return tok
}
