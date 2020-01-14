package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Printer generates yolol-code from an AST
type Printer struct {
	// This function is called whenever an unknown node-type is encountered.
	// It can be used to add support for additional types to the generator
	// returns the yolol-code for the giben node or an error
	UnknownHandlerFunc func(node Node, visitType int) (string, error)
}

var operatorPriority = map[string]int{
	"or":  0,
	"and": 0,
	"==":  1,
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
func (y *Printer) Print(prog Node) (string, error) {
	return y.PrintCommented(prog, nil)
}

// PrintCommented returns the yolol-code the ast-node and it's children represent
// If you are not printing a parser.Program, you can use this to enable printing of comments for ither node-types
func (y *Printer) PrintCommented(prog Node, commentList []*Token) (string, error) {
	commentIndex := 0
	output := ""
	err := prog.Accept(VisitorFunc(func(node Node, visitType int) error {
		// add the original comments to the output
		if commentList != nil && len(commentList) > commentIndex {
			if commentList[commentIndex].Position.Before(node.Start()) {
				output += commentList[commentIndex].Value
				commentIndex++
			}
		}
		switch n := node.(type) {
		case *Program:
			commentList = n.Comments
			break
		case *Line:
			if visitType == PostVisit {
				if commentList != nil && len(commentList) > commentIndex && commentList[commentIndex].Position.Line == n.End().Line {
					output += commentList[commentIndex].Value
					commentIndex++
				}
				output += "\n"
			}
			if visitType > 0 {
				output += " "
			}
			break
		case *Assignment:
			if visitType == PreVisit {
				output += n.Variable + n.Operator
			}
			break
		case *IfStatement:
			output += y.printIf(visitType)
			break
		case *GoToStatement:
			output += "goto " + strconv.Itoa(n.Line)
			break
		case *Dereference:
			output += y.printDeref(n)
			break
		case *StringConstant:
			output += "\"" + n.Value + "\""
			break
		case *NumberConstant:
			if strings.HasPrefix(n.Value, "-") {
				output += " "
			}
			output += fmt.Sprintf(n.Value)
			break
		case *BinaryOperation:
			output += y.printBinaryOperation(n, visitType)
			break
		case *UnaryOperation:
			_, childBinary := n.Exp.(*BinaryOperation)
			if visitType == PreVisit {
				op := n.Operator
				if op == "not" {
					op = " " + op + " "
				}
				if op == "-" {
					op = " " + op
				}
				output += op
				if childBinary {
					output += "("
				}
			}
			if visitType == PostVisit {
				if childBinary {
					output += ")"
				}
			}
			break
		case *FuncCall:
			if visitType == PreVisit {
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

func (y *Printer) printBinaryOperation(o *BinaryOperation, visitType int) string {
	lPrio := priorityForExpression(o.Exp1)
	rPrio := priorityForExpression(o.Exp2)
	_, rBinary := o.Exp2.(*BinaryOperation)
	myPrio := priorityForExpression(o)
	output := ""
	switch visitType {
	case PreVisit:
		if lPrio < myPrio {
			output += "("
		}
		break
	case InterVisit1:
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
	case PostVisit:
		if rBinary && rPrio <= myPrio {
			output += ")"
		}
		break
	}
	return output
}

func priorityForExpression(e Expression) int {
	switch ex := e.(type) {
	case *BinaryOperation:
		return operatorPriority[ex.Operator]
	default:
		return 10
	}
}

func (y *Printer) printIf(visitType int) string {

	switch visitType {
	case PreVisit:
		return "if "
	case InterVisit1:
		return " then "
	case InterVisit2:
		return " else "
	case PostVisit:
		return " end"
	default:
		return " "
	}
}

func (y *Printer) printDeref(d *Dereference) string {
	txt := ""
	if d.PrePost == "Pre" {
		txt += " " + d.Operator
	}
	txt += d.Variable
	if d.PrePost == "Post" {
		txt += d.Operator + " "
	}
	return txt
}
