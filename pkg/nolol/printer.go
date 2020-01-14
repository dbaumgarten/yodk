package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/parser"
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
func (p *Printer) Print(prog *Program) (string, error) {
	indentLevel := 0

	indentation := func(amount int) string {
		ind := ""
		for i := 0; i < amount; i++ {
			ind += p.Indentation
		}
		return ind
	}

	p.yololPrinter.UnknownHandlerFunc = func(node parser.Node, visitType int) (string, error) {
		switch n := node.(type) {
		case *GoToLabelStatement:
			if visitType == parser.SingleVisit {
				return "goto " + n.Label, nil
			}
			return "", nil

		case *MultilineIf:
			switch visitType {
			case parser.PreVisit:
				return indentation(indentLevel) + "if ", nil
			case parser.InterVisit1:
				indentLevel++
				return " then\n", nil
			case parser.InterVisit2:
				return indentation(indentLevel-1) + "else\n", nil
			case parser.PostVisit:
				indentLevel--
				return indentation(indentLevel) + "end\n", nil
			default:
				return "", nil
			}
		case *WhileLoop:
			switch visitType {
			case parser.PreVisit:
				return indentation(indentLevel) + "while ", nil
			case parser.InterVisit1:
				indentLevel++
				return " do\n", nil
			case parser.PostVisit:
				indentLevel--
				return indentation(indentLevel) + "end\n", nil
			default:
				return "", nil
			}
		case *StatementLine:
			switch visitType {
			case parser.PreVisit:
				out := indentation(indentLevel)
				if n.Label != "" {
					out += n.Label + "> "
				}
				return out, nil
			default:
				return "", nil
			}
		case *ConstDeclaration:
			switch visitType {
			case parser.PreVisit:
				return "const " + n.Name + " = ", nil
			case parser.PostVisit:
				return "", nil
			}
		case *Program:
			return "", nil
		}
		return "", fmt.Errorf("Unknown node-type: %T", node)
	}

	return p.yololPrinter.PrintCommented(prog, prog.Comments)
}
