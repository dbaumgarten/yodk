package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Printer generates yolol-code from an AST
type Printer struct {
	// This function is called whenever an unknown node-type is encountered.
	// It can be used to add support for additional types to the generator
	// returns the yolol-code for the giben node or an error
	UnknownHandlerFunc func(node ast.Node, visitType int) (string, error)
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

// Print returns the yolol-code the ast-node and it's children represent
func (y *Printer) Print(prog ast.Node) (string, error) {
	output := ""
	err := prog.Accept(ast.VisitorFunc(func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *ast.Program:
			break
		case *ast.Line:
			if visitType == ast.PostVisit {
				if n.Comment != "" {
					if len(n.Statements) == 0 {
						output += n.Comment
					} else {
						output += " " + n.Comment
					}
				}
				output += "\n"
			}
			if visitType > 0 {
				output += " "
			}
			break
		case *ast.Assignment:
			if visitType == ast.PreVisit {
				output += n.VariableDisplayName + n.Operator
			}
			break
		case *ast.IfStatement:
			output += y.printIf(visitType)
			break
		case *ast.GoToStatement:
			output += "goto " + strconv.Itoa(n.Line)
			break
		case *ast.Dereference:
			output += y.printDeref(n)
			break
		case *ast.StringConstant:
			output += "\"" + n.Value + "\""
			break
		case *ast.NumberConstant:
			if strings.HasPrefix(n.Value, "-") {
				output += " "
			}
			output += fmt.Sprintf(n.Value)
			break
		case *ast.BinaryOperation:
			output += y.printBinaryOperation(n, visitType)
			break
		case *ast.UnaryOperation:
			_, childBinary := n.Exp.(*ast.BinaryOperation)
			if visitType == ast.PreVisit {
				op := n.Operator
				if op == "not" {
					op = " " + op + " "
				} else if op == "-" {
					op = " " + op
				} else {
					panic("Unknown operator: " + op)
				}
				output += op
				if childBinary {
					output += "("
				}
			}
			if visitType == ast.PostVisit {
				if childBinary {
					output += ")"
				}
			}
			break
		case *ast.FuncCall:
			if visitType == ast.PreVisit {
				output += n.Function + "("
			} else {
				output += ")"
			}
			break
		default:
			if y.UnknownHandlerFunc == nil {
				return fmt.Errorf("Unknown ast-node: %T%v", node, node)
			}
			str, err := y.UnknownHandlerFunc(node, visitType)
			if err != nil {
				return err
			}
			output += str
		}
		return nil
	}))

	if err != nil {
		return "", err
	}
	// during the generation duplicate spaces might appear. Remove them
	return strings.Replace(output, "  ", " ", -1), nil
}

func (y *Printer) printBinaryOperation(o *ast.BinaryOperation, visitType int) string {
	lPrio := priorityForExpression(o.Exp1)
	rPrio := priorityForExpression(o.Exp2)
	_, rBinary := o.Exp2.(*ast.BinaryOperation)
	myPrio := priorityForExpression(o)
	output := ""
	switch visitType {
	case ast.PreVisit:
		if lPrio < myPrio {
			output += "("
		}
		break
	case ast.InterVisit1:
		if lPrio < myPrio {
			output += ")"
		}
		op := o.Operator
		if op == "and" || op == "or" {
			op = " " + op + " "
		}
		output += op
		if rBinary && rPrio <= myPrio {
			output += "("
		}
		break
	case ast.PostVisit:
		if rBinary && rPrio <= myPrio {
			output += ")"
		}
		break
	}
	return output
}

func priorityForExpression(e ast.Expression) int {
	switch ex := e.(type) {
	case *ast.BinaryOperation:
		return operatorPriority[ex.Operator]
	default:
		return 10
	}
}

func (y *Printer) printIf(visitType int) string {

	switch visitType {
	case ast.PreVisit:
		return "if "
	case ast.InterVisit1:
		return " then "
	case ast.InterVisit2:
		return " else "
	case ast.PostVisit:
		return " end"
	default:
		return " "
	}
}

func (y *Printer) printDeref(d *ast.Dereference) string {
	txt := ""
	if d.PrePost == "Pre" {
		txt += " " + d.Operator
	}
	txt += d.VariableDisplayName
	if d.PrePost == "Post" {
		txt += d.Operator + " "
	}
	return txt
}
