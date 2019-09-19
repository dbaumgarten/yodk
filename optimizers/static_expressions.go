package optimizers

import (
	"fmt"

	"github.com/dbaumgarten/yodk/parser"
	"github.com/dbaumgarten/yodk/vm"
	"github.com/shopspring/decimal"
)

type StaticExpressionOptimizer struct {
}

func (o *StaticExpressionOptimizer) Optimize(prog *parser.Programm) error {
	return prog.Accept(o)
}

func (o *StaticExpressionOptimizer) Visit(node parser.Node, visitType int) error {
	if visitType == parser.PostVisit {
		switch n := node.(type) {
		case *parser.Assignment:
			n.Value = optimizeExpression(n.Value)
			break
		case *parser.IfStatement:
			n.Condition = optimizeExpression(n.Condition)
			break
		case *parser.BinaryOperation:
			n.Exp1 = optimizeExpression(n.Exp1)
			n.Exp2 = optimizeExpression(n.Exp2)
			break
		case *parser.UnaryOperation:
			n.Exp = optimizeExpression(n.Exp)
			break
		case *parser.FuncCall:
			n.Argument = optimizeExpression(n.Argument)
			break
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
	return exp
}

func isConstant(exp parser.Expression) bool {
	switch exp.(type) {
	case *parser.StringConstant:
		return true
	case *parser.NumberConstant:
		return true
	}
	return false
}

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

func varToConst(v *vm.Variable) parser.Expression {
	if v.IsNumber() {
		return &parser.NumberConstant{
			Value: v.Itoa(),
		}
	} else {
		return &parser.StringConstant{
			Value: v.String(),
		}
	}
}
