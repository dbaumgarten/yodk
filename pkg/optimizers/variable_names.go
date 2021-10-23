package optimizers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// VariableNameOptimizer replaces variable names with sorter names
// Names of external variables will be left unchanged
// Replacements are case-insensitive. "var" and "VaR" will be replaced with the same replacement
type VariableNameOptimizer struct {
	variableMappings map[string]string
	invertedMappings map[string]string
	blacklist        map[string]bool
	varNumber        int
}

// NewVariableNameOptimizer returns a new VariableNameOptimizer
func NewVariableNameOptimizer() *VariableNameOptimizer {
	return &VariableNameOptimizer{
		variableMappings: make(map[string]string),
		invertedMappings: make(map[string]string),
		blacklist:        make(map[string]bool),
		varNumber:        1,
	}
}

// SetBlacklist sets a list of output variable-names that shall never be produced by the optimizer
func (o *VariableNameOptimizer) SetBlacklist(bl []string) {
	o.blacklist = make(map[string]bool)
	for _, el := range bl {
		o.blacklist[el] = true
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
	lin := strings.ToLower(in)
	newName, exists := o.variableMappings[lin]
	if !exists {
		newName = o.getNextVarName()
		_, isBlacklisted := o.blacklist[newName]
		for isBlacklisted {
			newName = o.getNextVarName()
			_, isBlacklisted = o.blacklist[newName]
		}
		o.variableMappings[lin] = newName
		o.invertedMappings[newName] = in
	}
	return newName
}

// GetReversalTable returns a map that can be used to translated the shortened names back to the original names
func (o *VariableNameOptimizer) GetReversalTable() map[string]string {
	return o.invertedMappings
}

// Visit is needed to implement Visitor
func (o *VariableNameOptimizer) Visit(node ast.Node, visitType int) error {
	if visitType == ast.SingleVisit || visitType == ast.PreVisit {
		switch n := node.(type) {
		// only change the display name of the variable
		// this way, it is shortened when generating code, but remains the same in the debugger
		case *ast.Assignment:
			n.Variable = o.OptimizeVarName(n.Variable)
			break
		case *ast.Dereference:
			n.Variable = o.OptimizeVarName(n.Variable)
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

// InitializeByFrequency will initialize the replacement-table based on the frequency of each variable in the input ast
// The most frequent variable will get the replacement a, the second most frequent gets b and so on.
// If two variables are equally frequent, the variable defined earlier wins.
// This way the most frequent variables will get the shortest replacements
// variables listed in ignore (can be nil) are ignored
func (o *VariableNameOptimizer) InitializeByFrequency(node ast.Node, ignore []string) {
	type Entry struct {
		Variable string
		Count    int
		Index    int
	}
	occurences := make(map[string]*Entry)

	ignoredict := make(map[string]bool)
	for _, varname := range ignore {
		ignoredict[varname] = true
	}

	f := func(node ast.Node, visitType int) error {
		if visitType == ast.SingleVisit || visitType == ast.PreVisit {
			var name string
			switch n := node.(type) {
			case *ast.Assignment:
				name = n.Variable
			case *ast.Dereference:
				name = n.Variable
			default:
				return nil
			}

			if _, shouldBeIgnored := ignoredict[name]; shouldBeIgnored {
				return nil
			}

			prev, exist := occurences[name]
			if exist {
				prev.Count++
				occurences[name] = prev
			} else {
				occurences[name] = &Entry{
					Variable: name,
					Count:    0,
					Index:    len(occurences),
				}
			}
		}
		return nil
	}

	node.Accept(ast.VisitorFunc(f))

	li := make([]*Entry, len(occurences))
	i := 0
	for _, entry := range occurences {
		li[i] = entry
		i++
	}

	sort.Slice(li, func(i, j int) bool {
		first := li[i]
		second := li[j]
		return first.Count > second.Count || (first.Count == second.Count && first.Index < second.Index)
	})

	for _, entry := range li {
		o.OptimizeVarName(entry.Variable)
	}

}
