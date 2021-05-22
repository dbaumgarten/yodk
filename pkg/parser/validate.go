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

// RemoveParenthesis removes all ()-nodes from the ast. These are superflous but were required duing parding because of bugs in starbase
func RemoveParenthesis(n ast.Node) {
	g := func(node ast.Node, visitType int) error {
		if unar, is := node.(*ast.UnaryOperation); is {
			if unar.Operator == "()" {
				return ast.NewNodeReplacement(unar.Exp)
			}
		}
		return nil
	}
	n.Accept(ast.VisitorFunc(g))
}

// Validate checks the program for certain errors. Usualy this is to emulate bug's in the game's implementation
func Validate(prog *ast.Program) []*Error {
	errors := make([]*Error, 0)

	checkForInnerNot := func(n ast.Node) {
		if innerun, is := n.(*ast.UnaryOperation); is {
			if innerun.Operator == "not" {
				errors = append(errors, &Error{
					Message:       "Because of an ingame-bug this expression must be wrapped in parenthesis",
					StartPosition: innerun.Start(),
					EndPosition:   innerun.End(),
				})
			}
		}
	}

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
			case *ast.UnaryOperation:
				if n.Operator == "not" {
					if innerbin, is := n.Exp.(*ast.BinaryOperation); is {
						if innerbin.Operator == "and" || innerbin.Operator == "or" {
							errors = append(errors, &Error{
								Message:       "Because of an ingame-bug this expression must be wrapped in parenthesis",
								StartPosition: innerbin.Start(),
								EndPosition:   innerbin.End(),
							})
						}
					}
				}
				if n.Operator != "()" {
					checkForInnerNot(n.Exp)
				}
			case *ast.BinaryOperation:
				if n.Operator == "and" || n.Operator == "or" {
					checkForInnerNot(n.Exp1)
				} else {
					checkForInnerNot(n.Exp2)
				}

			}
		}
		return nil
	}

	prog.Accept(ast.VisitorFunc(f))

	return errors
}
