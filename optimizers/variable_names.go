package optimizers

import (
	"fmt"
	"strings"

	"github.com/dbaumgarten/yodk/ast"
)

type VariableNameOptimizer struct {
	variableNames map[string]string
}

func NewVariableNameOptimizer() *VariableNameOptimizer {
	return &VariableNameOptimizer{
		variableNames: make(map[string]string),
	}
}

func (o *VariableNameOptimizer) Optimize(prog *ast.Programm) error {
	return prog.Accept(o)
}

func (o *VariableNameOptimizer) Visit(node ast.Node, visitType int) error {
	if visitType == ast.SingleVisit || visitType == ast.PreVisit {
		switch n := node.(type) {
		case *ast.Assignment:
			n.Variable = o.replaceVarName(n.Variable)
			break
		case *ast.Dereference:
			n.Variable = o.replaceVarName(n.Variable)
			break

		}
	}
	return nil
}

func (o *VariableNameOptimizer) replaceVarName(in string) string {
	// do not modify external variables
	if strings.HasPrefix(in, ":") {
		return in
	}
	newName, exists := o.variableNames[in]
	if !exists {
		newName = o.getNextVarName()
		o.variableNames[in] = newName
	}
	return newName
}

func (o *VariableNameOptimizer) getNextVarName() string {
	base := 26
	varnum := len(o.variableNames) + 1
	varname := ""
	for varnum > 0 {
		rem := (varnum % base)
		if rem == 0 {
			rem = base
			varnum = (varnum / base) - 1
		} else {
			varnum /= base
		}
		varname = fmt.Sprintf("%c", rem-1+97) + varname
	}
	return varname
}
