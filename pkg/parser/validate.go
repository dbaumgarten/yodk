package parser

import (
	"regexp"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

var brokenVarnameRegex = regexp.MustCompile("if|then|else|end|goto")

func varnameError(at ast.Node) *Error {
	return &Error{
		Message:       "Variable-names containing if|then|else|end|goto do not work ingame",
		StartPosition: at.Start(),
		EndPosition:   at.End(),
	}
}

// Validate checks the program for certain errors. Usualy this is to emulate bug's in the game's implementation
func Validate(prog *ast.Program) []*Error {
	errors := make([]*Error, 0)

	f := func(node ast.Node, visitType int) error {
		if visitType == ast.PostVisit || visitType == ast.SingleVisit {
			switch n := node.(type) {
			case *ast.Assignment:
				if brokenVarnameRegex.MatchString(n.Variable) {
					errors = append(errors, varnameError(n))
				}
				break
			case *ast.Dereference:
				if brokenVarnameRegex.MatchString(n.Variable) {
					errors = append(errors, varnameError(n))
				}
			}
		}
		return nil
	}

	prog.Accept(ast.VisitorFunc(f))

	return errors
}
