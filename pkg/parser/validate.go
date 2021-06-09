package parser

import (
	"regexp"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

var brokenVarnameRegex = regexp.MustCompile("if|then|else|end|goto")

const (
	ValidateLocalVars   = 1
	ValidateGlobalVars  = 2
	ValidateLogicalNots = 4
	ValidateFactorials  = 8
	ValidateAll         = ValidateLocalVars | ValidateGlobalVars | ValidateLogicalNots | ValidateFactorials
)

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
func Validate(prog ast.Node, flags int) []*Error {
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

	shouldValidate := func(flag int) bool {
		return (flags & flag) != 0
	}

	checkVarname := func(name string, node ast.Node) {
		isGlobal := strings.HasPrefix(name, ":")
		if (isGlobal && shouldValidate(ValidateGlobalVars)) || (!isGlobal && shouldValidate(ValidateLocalVars)) {
			if brokenVarnameRegex.MatchString(name) {
				errors = append(errors, &Error{
					Message:       "Variable-names containing if|then|else|end|goto do not work ingame",
					StartPosition: node.Start(),
					EndPosition:   node.End(),
				})
			}
		}
	}

	f := func(node ast.Node, visitType int) error {
		if visitType == ast.PostVisit || visitType == ast.SingleVisit {
			switch n := node.(type) {
			case *ast.NumberConstant:
				if n.Value == problematicNumberConstant {
					errors = append(errors, &Error{
						Message:       "Invalid number-constant. This value is too large.",
						StartPosition: n.Start(),
						EndPosition:   n.End(),
					})
				}
			case *ast.Assignment:
				checkVarname(n.Variable, n)
				break
			case *ast.Dereference:
				checkVarname(n.Variable, n)
			case *ast.UnaryOperation:
				if shouldValidate(ValidateLogicalNots) {
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
				}
				if shouldValidate(ValidateFactorials) {
					if n.Operator == "!" {
						switch n.Exp.(type) {
						case *ast.NumberConstant:
							break
						case *ast.Dereference:
							break
						default:
							errors = append(errors, &Error{
								Message:       "Yolol only allows factorials on number-constants and variables.",
								StartPosition: n.Exp.Start(),
								EndPosition:   n.Exp.End(),
							})

						}
					}

				}
			case *ast.BinaryOperation:
				if shouldValidate(ValidateLogicalNots) {
					if n.Operator == "and" || n.Operator == "or" {
						checkForInnerNot(n.Exp1)
					} else {
						checkForInnerNot(n.Exp2)
					}
				}
			}
		}
		return nil
	}

	prog.Accept(ast.VisitorFunc(f))

	return errors
}
