package optimizers

import (
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// CommentOptimizer removes all comments from the code
type CommentOptimizer struct {
}

// Optimize is needed to implement Optimizer
func (o *CommentOptimizer) Optimize(prog ast.Node) error {
	err := prog.Accept(o)
	if err != nil {
		return err
	}

	// remove trailing empty lines
	if prog, is := prog.(*ast.Program); is {
		for i := len(prog.Lines) - 1; i >= 0; i-- {
			hasStatements := len(prog.Lines[i].Statements) > 0
			if !hasStatements {
				prog.Lines = prog.Lines[:i]
			} else {
				break
			}
		}
	}

	return nil
}

// Visit is needed to implement Visitor
func (o *CommentOptimizer) Visit(node ast.Node, visitType int) error {
	if visitType == ast.SingleVisit || visitType == ast.PreVisit {
		switch n := node.(type) {
		case *ast.Line:
			n.Comment = ""
			break
		}
	}
	return nil
}
