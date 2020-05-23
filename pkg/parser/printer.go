package parser

import (
	"fmt"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Printermode describes various modes for the printer
type Printermode int

const (
	// PrintermodeReadable inserts spaces the improve readability
	PrintermodeReadable Printermode = 0
	// PrintermodeShort inserts only spaces that are reasonably necessary
	PrintermodeShort Printermode = 1
)

// Printer generates yolol-code from an AST
type Printer struct {
	// This function is called whenever an unknown node-type is encountered.
	// It can be used to add support for additional types to the generator
	// returns the yolol-code for the giben node or an error
	UnknownHandlerFunc func(node ast.Node, visitType int, p *Printer) error
	// If true, only insert spaces where absolutely necessary
	Mode         Printermode
	text         string
	lastWasSpace bool
}

var operatorPriority = map[string]int{
	"or":  0,
	"and": 0,
	"==":  1,
	"!=":  1,
	">=":  1,
	"<=":  1,
	">":   1,
	"<":   1,
	"+":   2,
	"-":   2,
	"*":   3,
	"/":   3,
	"^":   3,
	"%":   3,
	"not": 4,
}

// Write adds text to the source-code that is currently build
func (p *Printer) Write(content string) {
	p.text += content
	p.lastWasSpace = false
}

// Space adds a space to the source-code that is currently build
func (p *Printer) Space() {
	if !p.lastWasSpace {
		p.text += " "
	}
	p.lastWasSpace = true
}

// OptionalSpace adds a space to the source-code that is currently build, IF we are not producing compressed output
func (p *Printer) OptionalSpace() {
	if p.Mode == PrintermodeReadable {
		p.Space()
	}
}

// StatementSeparator writes spaces to seperate statements on one line. Amount of spaces depends on settings
func (p *Printer) StatementSeparator() {
	if p.Mode == PrintermodeReadable {
		p.Write(" ")
		p.Space()
	} else {
		p.Space()
	}
}

// Newline adds a newline to the source-code that is currently build
func (p *Printer) Newline() {
	p.text += "\n"
	p.lastWasSpace = false
}

// Print returns the yolol-code the ast-node and it's children represent
func (p *Printer) Print(prog ast.Node) (string, error) {
	p.text = ""
	p.lastWasSpace = false
	err := prog.Accept(ast.VisitorFunc(func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *ast.Program:
			break
		case *ast.Line:
			if visitType == ast.PostVisit {
				if n.Comment != "" {
					if len(n.Statements) != 0 {
						p.Space()
					}
					p.Write(n.Comment)
				}
				p.Newline()
			}
			if visitType > 0 {
				p.StatementSeparator()
			}
			break
		case *ast.Assignment:
			if visitType == ast.PreVisit {
				p.Write(n.Variable)
				p.OptionalSpace()
				p.Write(n.Operator)
				p.OptionalSpace()
			}
			break
		case *ast.IfStatement:
			p.printIf(visitType)
			break
		case *ast.GoToStatement:
			if visitType == ast.PreVisit {
				p.Write("goto")
				p.Space()
			}
			break
		case *ast.Dereference:
			p.printDeref(n)
			break
		case *ast.StringConstant:
			p.Write("\"" + insertEscapesIntoString(n.Value) + "\"")
			break
		case *ast.NumberConstant:
			if strings.HasPrefix(n.Value, "-") {
				p.Space()
			}
			p.Write(fmt.Sprintf(n.Value))
			break
		case *ast.BinaryOperation:
			p.printBinaryOperation(n, visitType)
			break
		case *ast.UnaryOperation:
			_, childBinary := n.Exp.(*ast.BinaryOperation)
			if visitType == ast.PreVisit {
				op := n.Operator
				if op == "-" {
					p.Space()
					p.Write(op)
				} else {
					p.Write(op)
					p.Space()
				}
				if childBinary {
					p.Write("(")
				}
			}
			if visitType == ast.PostVisit {
				if childBinary {
					p.Write(")")
				}
			}
			break
		default:
			if p.UnknownHandlerFunc == nil {
				return fmt.Errorf("Unknown ast-node: %T%v", node, node)
			}
			err := p.UnknownHandlerFunc(node, visitType, p)
			if err != nil {
				return err
			}
		}
		return nil
	}))

	if err != nil {
		return "", err
	}

	return p.text, nil
}

func insertEscapesIntoString(in string) string {
	in = strings.Replace(in, "\n", "\\n", -1)
	in = strings.Replace(in, "\t", "\\t", -1)
	in = strings.Replace(in, "\"", "\\\"", -1)
	return in
}

func (p *Printer) printBinaryOperation(o *ast.BinaryOperation, visitType int) {
	lPrio := priorityForExpression(o.Exp1)
	rPrio := priorityForExpression(o.Exp2)
	_, rBinary := o.Exp2.(*ast.BinaryOperation)
	myPrio := priorityForExpression(o)
	switch visitType {
	case ast.PreVisit:
		if lPrio < myPrio {
			p.Write("(")
		}
		break
	case ast.InterVisit1:
		if lPrio < myPrio {
			p.Write(")")
		}
		op := o.Operator
		if op == "and" || op == "or" {
			p.Space()
			p.Write(op)
			p.Space()
		} else {
			p.OptionalSpace()
			p.Write(op)
			p.OptionalSpace()
		}

		if rBinary && rPrio <= myPrio {
			p.Write("(")
		}
		break
	case ast.PostVisit:
		if rBinary && rPrio <= myPrio {
			p.Write(")")
		}
		break
	}

}

func priorityForExpression(e ast.Expression) int {
	switch ex := e.(type) {
	case *ast.BinaryOperation:
		return operatorPriority[ex.Operator]
	default:
		return 10
	}
}

func (p *Printer) printIf(visitType int) {

	switch visitType {
	case ast.PreVisit:
		p.Write("if")
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
}

func (p *Printer) printDeref(d *ast.Dereference) {
	if d.PrePost == "Pre" {
		p.Space()
		p.Write(d.Operator)
	}
	p.Write(d.Variable)
	if d.PrePost == "Post" {
		p.Write(d.Operator)
		p.Space()
	}
}
