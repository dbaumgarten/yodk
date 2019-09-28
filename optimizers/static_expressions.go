package optimizers

import (
	"fmt"

	"github.com/dbaumgarten/yodk/parser"
	"github.com/dbaumgarten/yodk/vm"
	"github.com/shopspring/decimal"
)

// StaticExpressionOptimizer evaluates static expressions at compile-time
type StaticExpressionOptimizer struct {
}

// NewStaticExpressionOptimizer returns a new StaticExpressionOptimizer
func NewStaticExpressionOptimizer() *StaticExpressionOptimizer {
	return &StaticExpressionOptimizer{}
}

// Optimize is required to implement Optimizer
func (o *StaticExpressionOptimizer) Optimize(prog parser.Node) error {
	return prog.Accept(o)
}

// Visit is needed to implement the Visitor interface
func (o *StaticExpressionOptimizer) Visit(node parser.Node, visitType int) error {
	if visitType == parser.PostVisit {
		if exp, isexp := node.(parser.Expression); isexp {
			optimized := optimizeExpression(exp)
			if optimized != nil {
				return parser.NewNodeReplacement(optimized)
			}
		}
	}
	return nil
}

func optimizeExpression(exp parser.Expression) parser.Expression {
	switch n := exp.(type) {
	case *parser.FuncCall:
		if !isConstant(n.Argument) {
			break
		}
		res, err := vm.RunFunction(constToVar(n.Argument), n.Function)
		if err != nil {
			break
		}
		return varToConst(res)
	case *parser.UnaryOperation:
		if !isConstant(n.Exp) {
			break
		}

		res, err := vm.RunUnaryOperation(constToVar(n.Exp), n.Operator)
		if err != nil {
			fmt.Println(err)
			break
		}
		return varToConst(res)
	case *parser.BinaryOperation:
		if !isConstant(n.Exp1) || !isConstant(n.Exp2) {
			break
		}
		res, err := vm.RunBinaryOperation(constToVar(n.Exp1), constToVar(n.Exp2), n.Operator)
		if err != nil {
			break
		}
		return varToConst(res)
	}
	return nil
}

// is the given expression constant (does not depend on a variable)
func isConstant(exp parser.Expression) bool {
	switch exp.(type) {
	case *parser.StringConstant:
		return true
	case *parser.NumberConstant:
		return true
	}
	return false
}

// convert a constant AST-node to a vm-variable
func constToVar(exp parser.Expression) *vm.Variable {
	switch e := exp.(type) {
	case *parser.StringConstant:
		return &vm.Variable{Value: e.Value}
	case *parser.NumberConstant:
		num, err := decimal.NewFromString(e.Value)
		if err != nil {
			panic("This should never happen")
		}
		return &vm.Variable{Value: num}
	}
	panic("This should never happen")
}

// convert a vm-variable to a constant AST-Node
func varToConst(v *vm.Variable) parser.Expression {
	if v.IsNumber() {
		return &parser.NumberConstant{
			Value: v.Itoa(),
		}
	}
	return &parser.StringConstant{
		Value: v.String(),
	}
}
