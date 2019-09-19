package optimizers

import "github.com/dbaumgarten/yodk/parser"

type Optimizer interface {
	Optimize(prog *parser.Programm) error
}

type CompoundOptimizer struct {
	seopt  *StaticExpressionOptimizer
	varopt *VariableNameOptimizer
}

func NewCompoundOptimizer() *CompoundOptimizer {
	return &CompoundOptimizer{
		seopt:  &StaticExpressionOptimizer{},
		varopt: NewVariableNameOptimizer(),
	}
}

func (co *CompoundOptimizer) Optimize(prog *parser.Programm) error {
	err := co.seopt.Optimize(prog)
	if err != nil {
		return err
	}
	return co.varopt.Optimize(prog)
}
