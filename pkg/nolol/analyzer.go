package nolol

import (
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// AnalysisReport contains collected information about a nolol-programm
type AnalysisReport struct {
	FileDocstring string
	Definitions   map[string]*nast.Definition
	Macros        map[string]*nast.MacroDefinition
	Variables     []string
	Labels        []string
	Docstrings    map[string]string
}

// Analyse returns an AnalysisReport for the given program
// All includes in the input-program must have been already resolved (use the converter for this)
// The input-programm is mutated. Do NOT use it after analysis
func Analyse(prog *nast.Program) (*AnalysisReport, error) {
	res := &AnalysisReport{
		Definitions: make(map[string]*nast.Definition),
		Macros:      make(map[string]*nast.MacroDefinition),
		Docstrings:  make(map[string]string),
		Labels:      make([]string, 0),
	}

	prevDocstrings := ""
	isStartOfFile := true

	vars := make(map[string]bool)

	f := func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *nast.Definition:
			isStartOfFile = false
			res.Definitions[n.Name] = n
			if prevDocstrings != "" {
				res.Docstrings[n.Name] = prevDocstrings
			}
			return ast.NewNodeReplacementSkip()
		case *nast.MacroDefinition:
			isStartOfFile = false
			res.Macros[n.Name] = n
			if prevDocstrings != "" {
				res.Docstrings[n.Name] = prevDocstrings
			}
			return ast.NewNodeReplacementSkip()
		case *ast.Assignment:
			isStartOfFile = false
			if _, isDef := res.Definitions[n.Variable]; !isDef {
				vars[n.Variable] = true
			}
			return nil
		case *ast.Dereference:
			isStartOfFile = false
			if _, isDef := res.Definitions[n.Variable]; !isDef {
				vars[n.Variable] = true
			}
			return nil
		case *nast.StatementLine:
			if visitType == ast.PostVisit {
				if isDocsLine(n) {
					prevDocstrings += strings.TrimLeft(n.Comment, "/ \t") + "\n"
				} else {
					if isStartOfFile {
						res.FileDocstring = prevDocstrings
						isStartOfFile = false
					}
					prevDocstrings = ""
				}
				if n.Label != "" {
					res.Labels = append(res.Labels, n.Label)
				}
			}
		default:
			//prevDocstrings = ""
		}
		return nil
	}

	err := prog.Accept(ast.VisitorFunc(f))
	if err != nil {
		return nil, err
	}

	res.Variables = make([]string, 0, len(vars))
	for k := range vars {
		res.Variables = append(res.Variables, k)
	}

	return res, err
}

// GetMacroLocalVars returns the local variables for the given macro
func (a AnalysisReport) GetMacroLocalVars(mac *nast.MacroDefinition) []string {
	variables := make(map[string]bool)
	f := func(node ast.Node, visitType int) error {
		if assign, is := node.(*ast.Assignment); visitType == ast.PreVisit && is {
			variables[assign.Variable] = true
		}
		if deref, is := node.(*ast.Dereference); visitType == ast.SingleVisit && is {
			variables[deref.Variable] = true
		}
		return nil
	}
	err := mac.Accept(ast.VisitorFunc(f))
	if err != nil {
		panic(err)
	}

	for _, arg := range mac.Arguments {
		variables[arg] = true
	}

	for _, arg := range mac.Externals {
		variables[arg] = true
	}

	vars := make([]string, 0, len(variables))
	for k := range variables {
		vars = append(vars, k)
	}

	return vars
}

// GetVarsAtLine returns all variables that are in scope at the given line
func (a AnalysisReport) GetVarsAtLine(line int) []string {
	vars := make([]string, 0, len(a.Variables)+10)
	for _, def := range a.Definitions {
		if def.Start().Line == line {
			vars = append(vars, a.Variables...)
			return vars
		}
	}
	for _, mac := range a.Macros {
		if mac.Start().Line < line && mac.End().Line >= line {
			vars = append(vars, a.GetMacroLocalVars(mac)...)
			return vars
		}
	}
	vars = append(vars, a.Variables...)
	return vars
}

// checks if the given line consist purely of a comment
func isDocsLine(thisStmtLine *nast.StatementLine) bool {
	if thisStmtLine.Comment != "" && len(thisStmtLine.Statements) == 0 && !thisStmtLine.HasBOL && !thisStmtLine.HasEOL {
		return true
	}
	return false
}
