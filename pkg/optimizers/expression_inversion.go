package optimizers

import (
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// ExpressionInversionOptimizer inverts negated expressions to shorten them
type ExpressionInversionOptimizer struct {
}

var inversions = map[string]string{
	">=": "<",
	"<=": ">",
	"<":  ">=",
	">":  "<=",
	"==": "!=",
	"!=": "==",
}

var andor = map[string]string{
	"and": "or",
	"or":  "and",
}

// Optimize is needed to implement Optimizer
func (o ExpressionInversionOptimizer) Optimize(prog ast.Node) error {
	return prog.Accept(o)
}

func pushDownNots(node ast.Expression) ast.Expression {
	if op, is := node.(*ast.UnaryOperation); is {
		if op.Operator == "not" {
			switch inner := op.Exp.(type) {
			case *ast.BinaryOperation:
				if opposite, invertable := inversions[inner.Operator]; invertable {
					inner.Operator = opposite
					return inner
				}
				if opposite, is := andor[inner.Operator]; is {
					inner.Operator = opposite
					inner.Exp1 = pushDownNots(&ast.UnaryOperation{
						Operator: "not",
						Exp:      inner.Exp1,
					})
					inner.Exp2 = pushDownNots(&ast.UnaryOperation{
						Operator: "not",
						Exp:      inner.Exp2,
					})
					return inner
				}
			case *ast.UnaryOperation:
				if inner.Operator == "not" {
					return inner.Exp
				}
			}
		}
	}
	return node
}

func bubbleUpNots(node ast.Expression) ast.Expression {
	if bin, isbinary := node.(*ast.BinaryOperation); isbinary {
		bin.Exp1 = bubbleUpNots(bin.Exp1)
		bin.Exp2 = bubbleUpNots(bin.Exp2)
		l, lisunary := bin.Exp1.(*ast.UnaryOperation)
		r, risunary := bin.Exp2.(*ast.UnaryOperation)
		if lisunary && risunary && l.Operator == "not" && r.Operator == "not" {
			if opposite, invertable := andor[bin.Operator]; invertable {
				bin.Operator = opposite
				bin.Exp1 = l.Exp
				bin.Exp2 = r.Exp
				return &ast.UnaryOperation{
					Operator: "not",
					Exp:      bin,
				}
			}
		}
	}
	return node
}

// OptimizeExpression optimizes a single expression
// Optimize() in contrast can only optimize whole programms
func (o ExpressionInversionOptimizer) OptimizeExpression(e ast.Expression) ast.Expression {
	return bubbleUpNots(pushDownNots(e))
}

// Visit is needed to implement Visitor
func (o ExpressionInversionOptimizer) Visit(node ast.Node, visitType int) error {
	if visitType == ast.PreVisit || visitType == ast.SingleVisit {
		replace := o.OptimizeExpression(node)
		if replace != node {
			return ast.NewNodeReplacement(replace)
		}
	}
	return nil
}
