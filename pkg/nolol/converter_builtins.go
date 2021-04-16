package nolol

import (
	"fmt"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// reservedTimeVariable is the variable used to track passed time
var reservedTimeVariable = "_time"

// convert a built-in function to yolol
func (c *Converter) convertFuncCall(function *nast.FuncCall, visitType int) error {

	if visitType != ast.PreVisit {
		return nil
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
		Message:       fmt.Sprintf("Unknown function or macro: %s(%d arguments)", function.Function, len(function.Arguments)),
		StartPosition: function.Start(),
		EndPosition:   function.End(),
	}
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
