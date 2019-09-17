package optimizers

import (
	"fmt"

	"github.com/dbaumgarten/yodk/ast"
	"github.com/dbaumgarten/yodk/vm"
	"github.com/shopspring/decimal"
)

type StaticExpressionOptimizer struct {
}

func (o *StaticExpressionOptimizer) Optimize(prog *ast.Programm) error {
	return prog.Accept(o)
}

func (o *StaticExpressionOptimizer) Visit(node ast.Node, visitType int) error {
	if visitType == ast.PostVisit {
		switch n := node.(type) {
		case *ast.Assignment:
			n.Value = optimizeExpression(n.Value)
			break
		case *ast.IfStatement:
			n.Condition = optimizeExpression(n.Condition)
			break
		case *ast.BinaryOperation:
			n.Exp1 = optimizeExpression(n.Exp1)
			n.Exp2 = optimizeExpression(n.Exp2)
			break
		case *ast.UnaryOperation:
			n.Exp = optimizeExpression(n.Exp)
			break
		case *ast.FuncCall:
			n.Argument = optimizeExpression(n.Argument)
			break
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
	return exp
}

func isConstant(exp ast.Expression) bool {
	switch exp.(type) {
	case *ast.StringConstant:
		return true
	case *ast.NumberConstant:
		return true
	}
	return false
}

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

func varToConst(v *vm.Variable) ast.Expression {
	if v.IsNumber() {
		return &ast.NumberConstant{
			Value: v.Itoa(),
		}
	} else {
		return &ast.StringConstant{
			Value: v.String(),
		}
	}
}
