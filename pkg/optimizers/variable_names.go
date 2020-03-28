package optimizers

import (
	"fmt"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// VariableNameOptimizer replaces variable names with sorter names
// Names of external variables will be left unchanged
type VariableNameOptimizer struct {
	variableMappings map[string]string
	varNumber        int
}

// NewVariableNameOptimizer returns a new VariableNameOptimizer
func NewVariableNameOptimizer() *VariableNameOptimizer {
	return &VariableNameOptimizer{
		variableMappings: make(map[string]string),
		varNumber:        1,
	}
}

// Optimize is needed to implement Optimizer
func (o *VariableNameOptimizer) Optimize(prog ast.Node) error {
	return prog.Accept(o)
}

// OptimizeVarName replaces a variable name with a new one (if it does not reference an external variable)
// the same input name will always result in the same output name
func (o *VariableNameOptimizer) OptimizeVarName(in string) string {
	// do not modify external variables
	if strings.HasPrefix(in, ":") {
		return in
	}
	newName, exists := o.variableMappings[in]
	if !exists {
		newName = o.getNextVarName()
		o.variableMappings[in] = newName
	}
	return newName
}

// Visit is needed to implement Visitor
func (o *VariableNameOptimizer) Visit(node ast.Node, visitType int) error {
	if visitType == ast.SingleVisit || visitType == ast.PreVisit {
		switch n := node.(type) {
		// only change the display name of the variable
		// this way, it is shortened when generating code, but remains the same in the debugger
		case *ast.Assignment:
			n.VariableDisplayName = o.OptimizeVarName(n.Variable)
			break
		case *ast.Dereference:
			n.VariableDisplayName = o.OptimizeVarName(n.Variable)
			break
		}
	}
	return nil
}

// generate a new variable name
func (o *VariableNameOptimizer) getNextVarName() string {
	for {
		base := 26
		varnum := o.varNumber
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
		o.varNumber++
		return varname
	}
}
