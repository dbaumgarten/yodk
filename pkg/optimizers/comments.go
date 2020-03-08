package optimizers

import (
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// CommentOptimizer removes all comments from the code
type CommentOptimizer struct {
}

// Optimize is needed to implement Optimizer
func (o *CommentOptimizer) Optimize(prog ast.Node) error {
	return prog.Accept(o)
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
