package optimizers

import (
	"fmt"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/parser"
)

// VariableNameOptimizer replaces variable names with sorter names
// Names of external variables will be left unchanged
type VariableNameOptimizer struct {
	variableNames map[string]string
}

// NewVariableNameOptimizer returns a new VariableNameOptimizer
func NewVariableNameOptimizer() *VariableNameOptimizer {
	return &VariableNameOptimizer{
		variableNames: make(map[string]string),
	}
}

// Optimize is needed to implement Optimizer
func (o *VariableNameOptimizer) Optimize(prog parser.Node) error {
	return prog.Accept(o)
}

// Visit is needed to implement Visitor
func (o *VariableNameOptimizer) Visit(node parser.Node, visitType int) error {
	if visitType == parser.SingleVisit || visitType == parser.PreVisit {
		switch n := node.(type) {
		case *parser.Assignment:
			n.Variable = o.replaceVarName(n.Variable)
			break
		case *parser.Dereference:
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
	newName, exists := o.variableNames[in]
	if !exists {
		newName = o.getNextVarName()
		o.variableNames[in] = newName
	}
	return newName
}

// generate a new variable name
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
