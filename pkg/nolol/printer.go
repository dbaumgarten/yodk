package nolol

import (
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Printer can generate the nolol-code corresponding to a nolol ast
type Printer struct {
	Indentation  string
	yololPrinter parser.Printer
	indentLevel  int
}

// NewPrinter creates a new Printer
func NewPrinter() *Printer {
	p := &Printer{
		yololPrinter: parser.Printer{},
		Indentation:  "\t",
	}
	p.yololPrinter.PrinterExtensionFunc = p.handleNololNodes
	return p
}

func (np *Printer) indentation() string {
	ind := ""
	for i := 0; i < np.indentLevel; i++ {
		ind += np.Indentation
	}
	return ind
}

func (np *Printer) handleNololNodes(node ast.Node, visitType int, p *parser.Printer) (bool, error) {
	switch n := node.(type) {
	case *nast.GoToLabelStatement:
		if visitType == ast.SingleVisit {
			p.Write("goto")
			p.Space()
			p.Write(n.Label)
		}
		break

	case *nast.Block:
		switch visitType {
		case ast.PreVisit:
			np.indentLevel++
			break
		case ast.PostVisit:
			np.indentLevel--
			break
		default:
			p.Write(np.indentation())
		}
		break

	case *nast.MacroDefinition:
		switch visitType {
		case ast.PreVisit:
			arglist := strings.Join(n.Arguments, ", ")
			p.Write("macro")
			p.Space()
			p.Write(n.Name)
			p.Write("(")
			p.Write(arglist)
			p.Write(")")
			if len(n.Externals) != 0 {
				extlist := strings.Join(n.Externals, ", ")
				p.Write("(")
				p.Write(extlist)
				p.Write(")")
			}
			p.Newline()
			break
		case ast.PostVisit:
			p.Write("end")
			break
		}
		break

	case *nast.MacroInsetion:
		switch visitType {
		case ast.PreVisit:
			p.Write("insert")
			p.Space()
			break
		case ast.PostVisit:
			p.Newline()
			break
		default:
		}
		break

	case *nast.FuncCall:
		switch visitType {
		case ast.PreVisit:
			p.Write(n.Function)
			p.Write("(")
			break
		case ast.PostVisit:
			p.Write(")")
			break
		default:
			if visitType > 0 {
				p.Write(",")
				p.OptionalSpace()
			}
		}
		break

	case *nast.MultilineIf:
		switch visitType {
		case ast.PreVisit:
			p.Write("if")
			p.Space()
			break
		case ast.InterVisit1:
			p.Space()
			p.Write("then")
			p.Newline()
			break
		case ast.InterVisit2:
			p.Write(np.indentation())
			p.Write("else")
			p.Newline()
			break
		case ast.PostVisit:
			p.Write(np.indentation())
			p.Write("end")
			p.Newline()
			break
		default:
			if visitType > 0 {
				p.Write(np.indentation())
				p.Write("else if")
				p.Space()
			}
		}
		break
	case *nast.WhileLoop:
		switch visitType {
		case ast.PreVisit:
			p.Write("while")
			p.Space()
			break
		case ast.InterVisit1:
			p.Space()
			p.Write("do")
			p.Newline()
		case ast.PostVisit:
			p.Write(np.indentation())
			p.Write("end")
			p.Newline()
			break
		default:
		}
		break
	case *nast.StatementLine:
		switch visitType {
		case ast.PreVisit:
			if n.Label != "" {
				p.Write(n.Label)
				p.Write(">")
				p.Space()
			}
			if n.HasBOL {
				p.Write("$")
				if len(n.Statements) > 0 {
					p.Space()
				}
			}
			break
		case ast.PostVisit:
			if n.HasEOL && len(n.Statements) > 0 {
				p.Space()
				p.Write("$")
			}
			if n.Comment != "" && (n.HasBOL || n.Label != "" || len(n.Statements) > 0) {
				// the line has a comment and somethin else. Seperate it by a space
				p.Space()
			}
			p.Write(n.Comment)
			p.Newline()
			break
		default:
			if visitType > 0 {
				p.Write("; ")
				p.OptionalSpace()
			}
		}
		break
	case *nast.IncludeDirective:
		p.Write("include")
		p.Space()
		p.Write("\"" + n.File + "\"")
		p.Newline()
		break
	case *nast.Definition:
		switch visitType {
		case ast.PreVisit:
			p.Write("define")
			p.Space()
			p.Write(n.Name)
			if len(n.Placeholders) != 0 {
				p.Write("(")
				phlist := strings.Join(n.Placeholders, ", ")
				p.Write(phlist)
				p.Write(")")
			}
			p.OptionalSpace()
			p.Write("=")
			p.OptionalSpace()
			break
		case ast.PostVisit:
			p.Newline()
			break
		}
	case *nast.Program:
		break
	case *nast.WaitDirective:
		if visitType == ast.PreVisit {
			p.Write("wait")
			p.Space()
			break
		}
		if n.Statements != nil {
			if visitType == ast.InterVisit1 {
				p.Space()
				p.Write("then")
				p.Space()
				break
			}
			if visitType > 0 {
				p.OptionalSpace()
				p.Write(";")
				p.OptionalSpace()
				break
			}
			if visitType == ast.PostVisit {
				p.Space()
				p.Write("end")
				break
			}
		}
		break
	case *nast.BreakStatement:
		p.Write("break")
		p.Space()
		break
	case *nast.ContinueStatement:
		p.Write("continue")
		p.Space()
		break
	case *ast.GoToStatement:
		if visitType == ast.PreVisit {
			p.Write("_goto")
			p.Space()
		}
		break
	case *ast.IfStatement:
		switch visitType {
		case ast.PreVisit:
			p.Write("_if")
			p.Space()
			break
		case ast.InterVisit1:
			p.Space()
			p.Write("then")
			p.Space()
			break
		case ast.InterVisit2:
			p.Space()
			p.Write("else")
			p.Space()
			break
		case ast.PostVisit:
			p.Space()
			p.Write("end")
			break
		default:
			if visitType > 0 {
				p.StatementSeparator()
			}
		}
	default:
		// This is not a type this function can print.
		// Return false, so the yolol-printer handles that node
		return false, nil
	}
	return true, nil
}

// Print returns the nolol-code for the given ast
func (np *Printer) Print(prog ast.Node) (string, error) {
	np.indentLevel = 0
	return np.yololPrinter.Print(prog)
}
