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

// OptimizeExpression optimizes a single expression recursively
func (o *StaticExpressionOptimizer) OptimizeExpression(e ast.Expression) ast.Expression {
	e, _ = ast.AcceptChild(o, e)
	return e
}

// Visit is needed to implement the Visitor interface
func (o *StaticExpressionOptimizer) Visit(node ast.Node, visitType int) error {
	if visitType == ast.PostVisit {
		if exp, isexp := node.(ast.Expression); isexp {
			optimized := o.OptimizeExpressionNonRecursive(exp)
			if optimized != nil {
				return ast.NewNodeReplacementSkip(optimized)
			}
		}
	}
	return nil
}

// OptimizeExpressionNonRecursive optimizes a single expression
// if no optimization is possible, nil is returned
func (o *StaticExpressionOptimizer) OptimizeExpressionNonRecursive(exp ast.Expression) ast.Expression {
	switch n := exp.(type) {
	case *ast.UnaryOperation:
		if !isConstant(n.Exp) {
			break
		}

		res, err := vm.RunUnaryOperation(constToVar(n.Exp), n.Operator)
		if err != nil {
			fmt.Println(err)
			break
		}
		return varToConst(res, n.Position)
	case *ast.BinaryOperation:
		if !isConstant(n.Exp1) || !isConstant(n.Exp2) {
			break
		}
		res, err := vm.RunBinaryOperation(constToVar(n.Exp1), constToVar(n.Exp2), n.Operator)
		if err != nil {
			break
		}

		// Most of the times the results of exponentiation are far longer then the original expression
		// Only pre-evaluate if the result is relatively short
		if n.Operator == "^" && res.IsNumber() && len(res.Number().String()) > 3 {
			break
		}

		return varToConst(res, n.Exp1.Start())
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
func varToConst(v *vm.Variable, pos ast.Position) ast.Expression {
	if v.IsNumber() {
		return &ast.NumberConstant{
			Value:    v.Itoa(),
			Position: pos,
		}
	}
	return &ast.StringConstant{
		Value:    v.String(),
		Position: pos,
	}
}
