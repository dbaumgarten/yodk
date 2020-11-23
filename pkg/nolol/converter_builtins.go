package nolol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// reservedTimeVariable is the variable used to track passed time
var reservedTimeVariable = "_time"

// convert a wait directive to yolol
func (c *Converter) convertWait(wait *nast.WaitDirective, visitType int) error {
	if visitType == ast.PostVisit {
		label := fmt.Sprintf("wait%d", c.waitlabelcounter)
		c.waitlabelcounter++
		line := &nast.StatementLine{
			Label:  label,
			HasEOL: true,
			Line: ast.Line{
				Position: wait.Start(),
				Statements: []ast.Statement{
					&ast.IfStatement{
						Position:  wait.Start(),
						Condition: wait.Condition,
						IfBlock: []ast.Statement{
							&nast.GoToLabelStatement{
								Label: label,
							},
						},
					},
				},
			},
		}
		if wait.Statements != nil {
			line.Statements = append(line.Statements, wait.Statements...)
		}
		if c.getLengthOfLine(&line.Line) > c.maxLineLength() {
			return &parser.Error{
				Message:       "The line is too long to be converted to yolol",
				StartPosition: wait.Start(),
				EndPosition:   wait.End(),
			}
		}
		return ast.NewNodeReplacementSkip(line)
	}
	return nil
}

// convert a built-in function to yolol
func (c *Converter) convertFuncCall(function *nast.FuncCall, visitType int) error {

	if visitType != ast.PreVisit {
		return nil
	}

	// first try if the funccall refers to a definition with replacements
	// if it returns nil, it doesnt
	err := c.convertDefinedFunction(function)
	if err != nil {
		return err
	}

	nfunc := strings.ToLower(function.Function)
	switch nfunc {
	case "time":
		// time is a nolol-built-in function
		if len(function.Arguments) != 0 {
			return &parser.Error{
				Message:       "The time() function takes no arguments",
				StartPosition: function.Start(),
				EndPosition:   function.End(),
			}
		}
		c.usesTimeTracking = true
		return ast.NewNodeReplacementSkip(&ast.Dereference{
			Variable: c.varnameOptimizer.OptimizeVarName(reservedTimeVariable),
		})
	case "line":
		if len(function.Arguments) > 1 {
			return &parser.Error{
				Message:       "The line() function takes zero or one arguments",
				StartPosition: function.Start(),
				EndPosition:   function.End(),
			}
		}
		if len(function.Arguments) == 1 {
			arg, is := function.Arguments[0].(*ast.Dereference)
			if !is || arg.Operator != "" {
				return &parser.Error{
					Message:       "The line() function takes a jump-label name (or nothing) as an argument",
					StartPosition: function.Arguments[0].Start(),
					EndPosition:   function.Arguments[0].End(),
				}
			}
		}
		// Replace the function with itself. This way we keep it for later processing, but skip processing the arguments
		return ast.NewNodeReplacementSkip(function)
	}
	unaryops := []string{"abs", "sqrt", "sin", "cos", "tan", "asin", "acos", "atan"}
	for _, unaryop := range unaryops {
		if unaryop == nfunc {
			if len(function.Arguments) != 1 {
				return &parser.Error{
					Message:       "The yolol-functions all take exactly one argument",
					StartPosition: function.Start(),
					EndPosition:   function.End(),
				}
			}
			return ast.NewNodeReplacement(&ast.UnaryOperation{
				Position: function.Position,
				Operator: nfunc,
				Exp:      function.Arguments[0],
			})
		}
	}
	return &parser.Error{
		Message:       fmt.Sprintf("Unknown function: %s(%d arguments)", function.Function, len(function.Arguments)),
		StartPosition: function.Start(),
		EndPosition:   function.End(),
	}
}

// convertLineFuncCalls converts ALL remaining line() calls in the programm to the acual line-number at the position
func (c *Converter) convertLineFuncCalls(prog *nast.Program) error {
	linecounter := 0
	f := func(node ast.Node, visitType int) error {
		if _, is := node.(*nast.StatementLine); is && visitType == ast.PreVisit {
			linecounter++
		}
		if function, is := node.(*nast.FuncCall); is {
			nfunc := strings.ToLower(function.Function)
			if nfunc != "line" {
				return &parser.Error{
					Message:       "This function can only convert the line() function",
					StartPosition: function.Start(),
					EndPosition:   function.End(),
				}
			}
			result := linecounter
			// when the line() function has an argument, that argument is a label-name
			// line(label) should then return the position of label
			if len(function.Arguments) == 1 {
				// at this point we are already sure, that it is a Dereference
				deref := function.Arguments[0].(*ast.Dereference)
				label := deref.Variable
				line, exists := c.jumpLabels[label]
				if !exists {
					return &parser.Error{
						Message:       "Unknown jump-label: " + label,
						StartPosition: deref.Start(),
						EndPosition:   deref.End(),
					}
				}
				result = line
			}
			return ast.NewNodeReplacementSkip(&ast.NumberConstant{
				Position: function.Position,
				Value:    strconv.Itoa(result),
			})
		}
		return nil
	}
	return prog.Accept(ast.VisitorFunc(f))
}

// checkes, if the program uses nolols time-tracking feature
func usesTimeTracking(n ast.Node) bool {
	uses := false
	f := func(node ast.Node, visitType int) error {
		if function, is := node.(*nast.FuncCall); is {
			if function.Function == "time" {
				uses = true
			}
		}
		return nil
	}
	n.Accept(ast.VisitorFunc(f))
	return uses
}

// inserts the line-counting statement into the beginning of each line
func (c *Converter) insertLineCounter(p *nast.Program) {
	for _, line := range p.Elements {
		if stmtline, is := line.(*nast.StatementLine); is {
			stmts := make([]ast.Statement, 1, len(stmtline.Statements)+1)
			stmts[0] = &ast.Dereference{
				Variable:    c.varnameOptimizer.OptimizeVarName(reservedTimeVariable),
				Operator:    "++",
				PrePost:     "Post",
				IsStatement: true,
			}
			stmts = append(stmts, stmtline.Statements...)
			stmtline.Statements = stmts
		}
	}
}
