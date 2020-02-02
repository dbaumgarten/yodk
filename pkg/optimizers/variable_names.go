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
	variableNames    map[string]struct{}
	varNumber        int
}

// NewVariableNameOptimizer returns a new VariableNameOptimizer
func NewVariableNameOptimizer() *VariableNameOptimizer {
	return &VariableNameOptimizer{
		variableMappings: make(map[string]string),
		variableNames:    make(map[string]struct{}),
		varNumber:        1,
	}
}

// Optimize is needed to implement Optimizer
func (o *VariableNameOptimizer) Optimize(prog ast.Node) error {
	return prog.Accept(o)
}

// SpecialReplacement registers aspecial replacement. Variables named in, will be renamed to out.
// No other variables will be renamed to out (no collions will occur)
func (o *VariableNameOptimizer) SpecialReplacement(in string, out string) {
	o.variableMappings[in] = out
	o.variableNames[out] = struct{}{}
}

// Visit is needed to implement Visitor
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

// replaces a variable name with a new one (if it does not reference an external variable)
func (o *VariableNameOptimizer) replaceVarName(in string) string {
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
		if _, exists := o.variableNames[varname]; exists {
			// we generated an already existing name. This can happen because of special replacements.
			// try again
			continue
		}
		o.variableNames[varname] = struct{}{}
		return varname
	}
}
