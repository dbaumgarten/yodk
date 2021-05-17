package parser

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Printermode describes various modes for the printer
type Printermode int

const (
	// PrintermodeCompact inserts only spaces that are reasonably necessary
	PrintermodeCompact Printermode = 0
	// PrintermodeReadable inserts spaces the improve readability
	PrintermodeReadable Printermode = 1
)

// Printer generates yolol-code from an AST
type Printer struct {
	// If this functiontion is set, it is called for every AST-node, before printing anything for that node.
	// It can be used to customize printing of certain nodes or add new kinds of nodes.
	// If it returns an errior, that errors is bubbled up.
	// The function should return true, if it can handle the given node and does not want this printer to continue processing it
	PrinterExtensionFunc func(node ast.Node, visitType int, p *Printer) (bool, error)
	// If true, only insert spaces where absolutely necessary
	Mode                      Printermode
	text                      string
	lastWasSpace              bool
	lastWasIdentifier         bool
	spaceIfNextIdentifierChar bool
	// If true, at position-information to every printed token.
	// Does not produce valid yolol, but is usefull for debugging
	DebugPositions bool
}

var binaryOperatorPriority = map[string]int{
	"and": 0,
	"or":  1,
	"+":   3,
	"-":   3,
	"==":  4,
	"!=":  4,
	">=":  4,
	"<=":  4,
	">":   4,
	"<":   4,
	"*":   5,
	"/":   5,
	"%":   5,
	"^":   6,
}

var unaryOperatorPriority = map[string]int{
	"not":  2,
	"abs":  7,
	"sqrt": 7,
	"sin":  7,
	"cos":  7,
	"tan":  7,
	"asin": 7,
	"acos": 7,
	"atan": 7,
	"-":    8,
	"!":    9,
}

func isIdentifierChar(b byte) bool {
	s := rune(b)
	if unicode.IsLetter(s) {
		return true
	}
	if unicode.IsDigit(s) {
		return true
	}
	if s == '_' || s == ':' {
		return true
	}
	return false
}

// Write adds text to the source-code that is currently build
func (p *Printer) Write(content string) {
	if p.spaceIfNextIdentifierChar && isIdentifierChar(content[0]) {
		p.Space()
	}
	p.text += content
	p.spaceIfNextIdentifierChar = false
	p.lastWasSpace = false
	p.lastWasIdentifier = false
}

// Space adds a space to the source-code that is currently build
func (p *Printer) Space() {
	if !p.lastWasSpace {
		p.text += " "
	}
	p.lastWasSpace = true
	p.lastWasIdentifier = false
	p.spaceIfNextIdentifierChar = false
}

// OptionalSpace adds a space to the source-code that is currently build, IF we are not producing compressed output
func (p *Printer) OptionalSpace() {
	if p.Mode == PrintermodeReadable {
		p.Space()
	}
}

// SpaceIfAfterIdentifier writes a space if the previously printed thing was an identifier
func (p *Printer) SpaceIfAfterIdentifier() {
	if p.Mode == PrintermodeReadable || p.lastWasIdentifier {
		p.Space()
	}
}

// SpaceIfFollowedByIdentifierChar adds a space, if the next character is one of the chars that are allowed inside identifiers
func (p *Printer) SpaceIfFollowedByIdentifierChar() {
	if p.Mode == PrintermodeReadable {
		p.Space()
	} else {
		p.spaceIfNextIdentifierChar = true
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
	p.lastWasIdentifier = false
	p.spaceIfNextIdentifierChar = false
}

// Print returns the yolol-code the ast-node and it's children represent
func (p *Printer) Print(prog ast.Node) (string, error) {
	p.text = ""
	p.lastWasSpace = false
	numberoflines := 0
	currentline := 0
	err := prog.Accept(ast.VisitorFunc(func(node ast.Node, visitType int) error {
		if (visitType == ast.PreVisit || visitType == ast.SingleVisit) && p.DebugPositions {
			p.Write(fmt.Sprintf("{%s(%v - %v)", reflect.TypeOf(node).String(), node.Start(), node.End()))
		}
		if p.PrinterExtensionFunc != nil {
			skip, err := p.PrinterExtensionFunc(node, visitType, p)
			if err != nil {
				return err
			}
			if skip {
				return nil
			}
		}
		switch n := node.(type) {
		case *ast.Program:
			if visitType == ast.PreVisit {
				numberoflines = len(n.Lines)
			}
			break
		case *ast.Line:
			if visitType == ast.PreVisit {
				currentline++
			}
			if visitType == ast.PostVisit {
				if n.Comment != "" {
					if len(n.Statements) != 0 {
						p.Space()
					}
					p.Write(n.Comment)
				}

				// Emit a newline after every line, except it is the last one and it is not empty
				if currentline != numberoflines || (len(n.Statements) == 0 && len(n.Comment) == 0) {
					p.Newline()
				}
			}
			if visitType > 0 {
				p.StatementSeparator()
			}
			break
		case *ast.Assignment:
			if visitType == ast.PreVisit {
				p.Write(n.Variable)
				p.lastWasIdentifier = true
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
				p.OptionalSpace()
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
			p.printUnaryOperation(n, visitType)
			break
		default:
			return fmt.Errorf("Unknown ast-node: %T%v", node, node)
		}
		if (visitType == ast.PostVisit || visitType == ast.SingleVisit) && p.DebugPositions {
			p.Write("}")
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
	rBinary, rIsBinary := o.Exp2.(*ast.BinaryOperation)
	lBinary, lIsBinary := o.Exp1.(*ast.BinaryOperation)
	myPrio := priorityForExpression(o)

	// check if we need braces because of the right-associativity of the ^-operator
	rightAssocL := o.Operator == "^" && lIsBinary && lBinary.Operator == "^"
	rightAssocR := o.Operator == "^" && rIsBinary && rBinary.Operator == "^"

	switch visitType {
	case ast.PreVisit:
		if lPrio < myPrio || rightAssocL {
			p.Write("(")
		}
		break
	case ast.InterVisit1:
		if lPrio < myPrio || rightAssocL {
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

		if rIsBinary && rPrio <= myPrio && !rightAssocR {
			p.Write("(")
		}
		break
	case ast.PostVisit:
		if rIsBinary && rPrio <= myPrio && !rightAssocR {
			p.Write(")")
		}
		break
	}
}

func (p *Printer) printUnaryOperation(o *ast.UnaryOperation, visitType int) {
	childPrio := priorityForExpression(o.Exp)
	thisPrio := priorityForExpression(o)
	if visitType == ast.PreVisit {
		if o.Operator == "-" {
			if p.text[len(p.text)-1] == byte('=') {
				p.Space()
			} else {
				p.OptionalSpace()
			}

			p.Write(o.Operator)
		} else if o.Operator == "!" {
			//do not write anything in PreVisit
		} else {
			p.Write(o.Operator)
			p.SpaceIfFollowedByIdentifierChar()
		}
		if childPrio < thisPrio {
			p.Write("(")
		}
	}
	if visitType == ast.PostVisit {
		if childPrio < thisPrio {
			p.Write(")")
		}
		if o.Operator == "!" {
			p.Write(o.Operator)
		}
	}
}

func priorityForExpression(e ast.Expression) int {
	switch ex := e.(type) {
	case *ast.BinaryOperation:
		return binaryOperatorPriority[ex.Operator]
	case *ast.UnaryOperation:
		return unaryOperatorPriority[ex.Operator]
	default:
		return 10
	}
}

func (p *Printer) printIf(visitType int) {

	switch visitType {
	case ast.PreVisit:
		p.Write("if")
		p.OptionalSpace()
		break
	case ast.InterVisit1:
		p.SpaceIfAfterIdentifier()
		p.Write("then")
		p.OptionalSpace()
		break
	case ast.InterVisit2:
		p.SpaceIfAfterIdentifier()
		p.Write("else")
		p.OptionalSpace()
		break
	case ast.PostVisit:
		p.SpaceIfAfterIdentifier()
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
	p.lastWasIdentifier = true
	if d.PrePost == "Post" {
		p.Write(d.Operator)
		p.Space()
	}
}
