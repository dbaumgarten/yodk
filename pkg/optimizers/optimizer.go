package optimizers

import "github.com/dbaumgarten/yodk/pkg/parser/ast"

// Optimizer is the common interface for all optimizers
type Optimizer interface {
	// Optimize optimizes the given ast. The ast is mutated (obviously)
	Optimize(prog *ast.Program) error
}

// CompoundOptimizer wraps all other optimizers and executes them
type CompoundOptimizer struct {
	seopt  *StaticExpressionOptimizer
	varopt *VariableNameOptimizer
	comopt *CommentOptimizer
	expinv *ExpressionInversionOptimizer
}

// NewCompoundOptimizer creates a new compound optimizer
func NewCompoundOptimizer() *CompoundOptimizer {
	return &CompoundOptimizer{
		seopt:  &StaticExpressionOptimizer{},
		varopt: NewVariableNameOptimizer(),
		comopt: &CommentOptimizer{},
		expinv: &ExpressionInversionOptimizer{},
	}
}

// Optimize is required to implement Optimizer
func (co *CompoundOptimizer) Optimize(prog *ast.Program) error {
	err := co.seopt.Optimize(prog)
	if err != nil {
		return err
	}
	err = co.comopt.Optimize(prog)
	if err != nil {
		return err
	}
	err = co.expinv.Optimize(prog)
	if err != nil {
		return err
	}
	return co.varopt.Optimize(prog)
}
