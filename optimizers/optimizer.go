package optimizers

import "github.com/dbaumgarten/yodk/parser"

// Optimizer is the common interface for all optimizers
type Optimizer interface {
	// Optimize optimizes the given ast. The ast is mutated (obviously)
	Optimize(prog *parser.Programm) error
}

// CompoundOptimizer wraps all other optimizers and executes them
type CompoundOptimizer struct {
	seopt  *StaticExpressionOptimizer
	varopt *VariableNameOptimizer
}

// NewCompoundOptimizer creates a new compound optimizer
func NewCompoundOptimizer() *CompoundOptimizer {
	return &CompoundOptimizer{
		seopt:  &StaticExpressionOptimizer{},
		varopt: NewVariableNameOptimizer(),
	}
}

// Optimize is required to implement Optimizer
func (co *CompoundOptimizer) Optimize(prog *parser.Programm) error {
	err := co.seopt.Optimize(prog)
	if err != nil {
		return err
	}
	return co.varopt.Optimize(prog)
}
