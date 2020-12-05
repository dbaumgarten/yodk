package nolol

import (
	"path/filepath"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// AnalysisReport contains collected information about a nolol-programm
type AnalysisReport struct {
	Definitions map[string]*nast.Definition
	Macros      map[string]*nast.MacroDefinition
	Variables   []string
	Labels      []string
	Docstrings  map[string]string
}

// AnalyseFile returns an AnalysisReport for the given file
func AnalyseFile(mainfile string) (*AnalysisReport, error) {
	files := DiskFileSystem{
		Dir: filepath.Dir(mainfile),
	}
	return AnalyseFileEx(filepath.Base(mainfile), files)
}

// AnalyseFileEx returns an AnalysisReport for the given file
func AnalyseFileEx(mainfile string, files FileSystem) (*AnalysisReport, error) {
	file, err := files.Get(mainfile)
	if err != nil {
		return nil, err
	}
	p := NewParser()
	parsed, err := p.Parse(file)
	if err != nil {
		return nil, err
	}
	return Analyse(parsed, files)
}

// Analyse returns an AnalysisReport for the given program
func Analyse(prog *nast.Program, files FileSystem) (*AnalysisReport, error) {
	res := &AnalysisReport{
		Definitions: make(map[string]*nast.Definition),
		Macros:      make(map[string]*nast.MacroDefinition),
		Docstrings:  make(map[string]string),
		Labels:      make([]string, 0),
	}

	includecount := 0
	prevDocstrings := ""

	vars := make(map[string]bool)

	f := func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *nast.Definition:
			res.Definitions[n.Name] = n
			if prevDocstrings != "" {
				res.Docstrings[n.Name] = prevDocstrings
			}
			return ast.NewNodeReplacementSkip()
		case *nast.MacroDefinition:
			res.Macros[n.Name] = n
			if prevDocstrings != "" {
				res.Docstrings[n.Name] = prevDocstrings
			}
			return ast.NewNodeReplacementSkip()
		case *ast.Assignment:
			if _, isDef := res.Definitions[n.Variable]; !isDef {
				vars[n.Variable] = true
			}
			return nil
		case *ast.Dereference:
			if _, isDef := res.Definitions[n.Variable]; !isDef {
				vars[n.Variable] = true
			}
			return nil
		case *nast.IncludeDirective:
			return convertInclude(n, files, &includecount)
		case *nast.StatementLine:
			if visitType == ast.PostVisit {
				if isDocsLine(n) {
					prevDocstrings += strings.TrimLeft(n.Comment, "/ \t") + "\n"
				} else {
					prevDocstrings = ""
				}
				if n.Label != "" {
					res.Labels = append(res.Labels, n.Label)
				}
			}
		case *nast.FuncCall:
			if visitType == ast.PreVisit && n.Function == "line" {
				n.Arguments = []ast.Expression{}
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

// GetDefinitionLocalVars returns the placeholders for the given Definition
func (a AnalysisReport) GetDefinitionLocalVars(def *nast.Definition) []string {
	return def.Placeholders
}

// GetVarsAtLine returns all variables that are in scope at the given line
func (a AnalysisReport) GetVarsAtLine(line int) []string {
	vars := make([]string, 0, len(a.Variables)+10)
	for _, def := range a.Definitions {
		if def.Start().Line == line {
			vars = append(vars, a.Variables...)
			vars = append(vars, a.GetDefinitionLocalVars(def)...)
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

// TODO: reduce code-duplication between here and the converter
func convertInclude(include *nast.IncludeDirective, files FileSystem, count *int) error {
	p := NewParser()

	*count++
	if *count > 20 {
		return nil
	}

	file, err := files.Get(include.File)
	if err != nil {
		return nil
	}
	p.SetFilename(include.File)
	parsed, err := p.Parse(file)
	if err != nil {
		return nil
	}

	replacements := make([]ast.Node, len(parsed.Elements))
	for i := range parsed.Elements {
		replacements[i] = parsed.Elements[i]
	}
	return ast.NewNodeReplacement(replacements...)
}
