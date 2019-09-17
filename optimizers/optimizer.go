package optimizers

import "github.com/dbaumgarten/yodk/ast"

type Optimizer interface {
	Optimize(prog *ast.Programm) error
}

type CompoundOptimizer struct {
	seopt StaticExpressionOptimizer
}

func (co *CompoundOptimizer) Optimize(prog *ast.Programm) error {
	return co.seopt.Optimize(prog)
}
