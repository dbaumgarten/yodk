package nolol

import (
	"fmt"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Printer can generate the nolol-code corresponding to a nolol ast
type Printer struct {
	yololPrinter parser.Printer
	Indentation  string
}

// NewPrinter creates a new Printer
func NewPrinter() *Printer {
	return &Printer{
		Indentation: "\t",
	}
}

// Print returns the nolol-code for the given ast
func (p *Printer) Print(prog ast.Node) (string, error) {
	indentLevel := 0

	indentation := func(amount int) string {
		ind := ""
		for i := 0; i < amount; i++ {
			ind += p.Indentation
		}
		return ind
	}

	p.yololPrinter.UnknownHandlerFunc = func(node ast.Node, visitType int) (string, error) {
		switch n := node.(type) {
		case *nast.GoToLabelStatement:
			if visitType == ast.SingleVisit {
				return "goto " + n.Label, nil
			}
			return "", nil

		case *nast.Block:
			switch visitType {
			case ast.PreVisit:
				indentLevel++
				return "", nil
			case ast.PostVisit:
				indentLevel--
				return "", nil
			default:
				return indentation(indentLevel), nil
			}

		case *nast.MacroDefinition:
			switch visitType {
			case ast.PreVisit:
				arglist := strings.Join(n.Arguments, ",")
				return "macro " + n.Name + "(" + arglist + ")\n", nil
			case ast.PostVisit:
				return "end", nil
			}

		case *nast.MacroInsetion:
			switch visitType {
			case ast.PreVisit:
				return "insert " + n.Name + "(", nil
			case ast.PostVisit:
				return ")\n", nil
			default:
				if visitType > 0 {
					return ",", nil
				}
				return "", nil
			}

		case *nast.MultilineIf:
			switch visitType {
			case ast.PreVisit:
				return "if ", nil
			case ast.InterVisit1:
				return " then\n", nil
			case ast.InterVisit2:
				return indentation(indentLevel) + "else\n", nil
			case ast.PostVisit:
				return indentation(indentLevel) + "end\n", nil
			default:
				if visitType > 0 {
					return indentation(indentLevel) + "else if ", nil
				}
				return "", nil
			}
		case *nast.WhileLoop:
			switch visitType {
			case ast.PreVisit:
				return "while ", nil
			case ast.InterVisit1:
				return " do\n", nil
			case ast.PostVisit:
				return indentation(indentLevel) + "end\n", nil
			default:
				return "", nil
			}
		case *nast.StatementLine:
			switch visitType {
			case ast.PreVisit:
				out := ""
				if n.Label != "" {
					out += n.Label + "> "
				}
				if n.HasBOL {
					out += "$"
					if len(n.Statements) > 0 {
						out += " "
					}
				}
				return out, nil
			case ast.PostVisit:
				out := ""
				if n.HasEOL && len(n.Statements) > 0 {
					out += " $"
				}
				if n.Comment != "" && (n.HasBOL || n.Label != "" || len(n.Statements) > 0) {
					// the line has a comment and somethin else. Seperate it by a space
					out += " "
				}
				out += n.Comment
				out += "\n"
				return out, nil
			default:
				if visitType > 0 {
					return "; ", nil
				}
			}
			return "", nil
		case *nast.IncludeDirective:
			return "include \"" + n.File + "\"\n", nil
		case *nast.Definition:
			switch visitType {
			case ast.PreVisit:
				return "define " + n.Name + " = ", nil
			case ast.PostVisit:
				return "\n", nil
			}
		case *nast.Program:
			return "", nil
		case *nast.WaitDirective:
			if visitType == ast.PreVisit {
				return "wait ", nil
			}
			return "", nil
		}
		return "", fmt.Errorf("Unknown node-type: %T", node)
	}

	return p.yololPrinter.Print(prog)
}
