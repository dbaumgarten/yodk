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
		yololPrinter: parser.Printer{
			Mode: parser.PrintermodeReadable,
		},
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

	p.yololPrinter.UnknownHandlerFunc = func(node ast.Node, visitType int, p *parser.Printer) error {
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
				indentLevel++
				break
			case ast.PostVisit:
				indentLevel--
				break
			default:
				p.Write(indentation(indentLevel))
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
				p.Write(indentation(indentLevel))
				p.Write("else")
				p.Newline()
				break
			case ast.PostVisit:
				p.Write(indentation(indentLevel))
				p.Write("end")
				p.Newline()
				break
			default:
				if visitType > 0 {
					p.Write(indentation(indentLevel))
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
				p.Write(indentation(indentLevel))
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
					p.Write(";")
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
		default:
			return fmt.Errorf("Unknown node-type: %T", node)
		}
		return nil
	}

	return p.yololPrinter.Print(prog)
}
