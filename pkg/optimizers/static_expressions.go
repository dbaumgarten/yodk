package optimizers

import (
	"github.com/dbaumgarten/yodk/pkg/number"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
	"github.com/dbaumgarten/yodk/pkg/vm"
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
	e, _ = ast.MustExpression(ast.AcceptChild(o, e))
	return e
}

// Visit is needed to implement the Visitor interface
func (o *StaticExpressionOptimizer) Visit(node ast.Node, visitType int) error {
	if visitType == ast.PostVisit || visitType == ast.SingleVisit {
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
	case *ast.NumberConstant:
		trimNumberConstant(n)
		return nil

	case *ast.UnaryOperation:
		if !isConstant(n.Exp) {
			break
		}

		res, err := vm.RunUnaryOperation(constToVar(n.Exp), n.Operator)
		if err != nil {
			break
		}

		// Most of the times the results of factorials are far longer then the original expression
		// Only pre-evaluate if the result is relatively short
		if n.Operator == "!" && res.IsNumber() && len(res.Number().String()) > 2 {
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

	if binop, is := exp.(*ast.BinaryOperation); is {
		return simplifyIdentities(binop)
	}

	return nil
}

func trimNumberConstant(cn *ast.NumberConstant) {
	num, err := number.FromString(cn.Value)
	if err == nil {
		cn.Value = num.String()
	}
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
		num, err := number.FromString(e.Value)
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

func simplifyIdentities(binop *ast.BinaryOperation) ast.Expression {

	if _, is := binop.Exp1.(*ast.StringConstant); is {
		return binop
	}
	if _, is := binop.Exp2.(*ast.StringConstant); is {
		return binop
	}

	switch binop.Operator {
	case "+":
		if isNumConstWithValue(binop.Exp1, "0") {
			return binop.Exp2
		}
		if isNumConstWithValue(binop.Exp2, "0") {
			return binop.Exp1
		}
	case "-":
		if isNumConstWithValue(binop.Exp1, "0") {
			return &ast.UnaryOperation{
				Position: binop.Start(),
				Operator: "-",
				Exp:      binop.Exp2,
			}
		}
		if isNumConstWithValue(binop.Exp2, "0") {
			return binop.Exp1
		}
	case "*":
		if isNumConstWithValue(binop.Exp1, "1") {
			return binop.Exp2
		}
		if isNumConstWithValue(binop.Exp2, "1") {
			return binop.Exp1
		}
		if isNumConstWithValue(binop.Exp1, "0") {
			return &ast.NumberConstant{Position: binop.Start(), Value: "0"}
		}
		if isNumConstWithValue(binop.Exp2, "0") {
			return &ast.NumberConstant{Position: binop.Start(), Value: "0"}
		}
	case "/":
		if isNumConstWithValue(binop.Exp2, "1") {
			return binop.Exp1
		}
	case "^":
		if isNumConstWithValue(binop.Exp2, "0") {
			return &ast.NumberConstant{Position: binop.Start(), Value: "1"}
		}
		if isNumConstWithValue(binop.Exp2, "1") {
			return binop.Exp1
		}
	}

	return binop
}

// checkis if check is a NumberConstant and if it's value matches expected
func isNumConstWithValue(check ast.Expression, expected string) bool {

	if num, is := check.(*ast.NumberConstant); is {
		if num.Value == expected {
			return true
		}
	}

	return false
}
