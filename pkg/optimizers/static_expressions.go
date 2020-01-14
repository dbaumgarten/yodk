package optimizers

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
	"github.com/dbaumgarten/yodk/pkg/vm"
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
func (o *StaticExpressionOptimizer) Optimize(prog ast.Node) error {
	return prog.Accept(o)
}

// Visit is needed to implement the Visitor interface
func (o *StaticExpressionOptimizer) Visit(node ast.Node, visitType int) error {
	if visitType == ast.PostVisit {
		if exp, isexp := node.(ast.Expression); isexp {
			optimized := optimizeExpression(exp)
			if optimized != nil {
				return ast.NewNodeReplacement(optimized)
			}
		}
	}
	return nil
}

func optimizeExpression(exp ast.Expression) ast.Expression {
	switch n := exp.(type) {
	case *ast.FuncCall:
		if !isConstant(n.Argument) {
			break
		}
		res, err := vm.RunFunction(constToVar(n.Argument), n.Function)
		if err != nil {
			break
		}
		return varToConst(res)
	case *ast.UnaryOperation:
		if !isConstant(n.Exp) {
			break
		}

		res, err := vm.RunUnaryOperation(constToVar(n.Exp), n.Operator)
		if err != nil {
			fmt.Println(err)
			break
		}
		return varToConst(res)
	case *ast.BinaryOperation:
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
func isConstant(exp ast.Expression) bool {
	switch exp.(type) {
	case *ast.StringConstant:
		return true
	case *ast.NumberConstant:
		return true
	}
	return false
}

// convert a constant AST-node to a vm-variable
func constToVar(exp ast.Expression) *vm.Variable {
	switch e := exp.(type) {
	case *ast.StringConstant:
		return &vm.Variable{Value: e.Value}
	case *ast.NumberConstant:
		num, err := decimal.NewFromString(e.Value)
		if err != nil {
			panic("This should never happen")
		}
		return &vm.Variable{Value: num}
	}
	panic("This should never happen")
}

// convert a vm-variable to a constant AST-Node
func varToConst(v *vm.Variable) ast.Expression {
	if v.IsNumber() {
		return &ast.NumberConstant{
			Value: v.Itoa(),
		}
	}
	return &ast.StringConstant{
		Value: v.String(),
	}
}
