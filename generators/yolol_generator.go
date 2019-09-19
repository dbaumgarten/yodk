package generators

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/parser"
)

type YololGenerator struct {
	programm string
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

func (y *YololGenerator) Visit(node parser.Node, visitType int) error {
	switch n := node.(type) {
	case *parser.Line:
		if visitType == parser.PostVisit {
			y.programm += "\n"
		}
		if visitType > 0 {
			y.programm += " "
		}
		break
	case *parser.Assignment:
		if visitType == parser.PreVisit {
			y.programm += n.Variable + n.Operator
		}
		break
	case *parser.IfStatement:
		y.generateIf(visitType)
		break
	case *parser.GoToStatement:
		y.programm += "goto " + strconv.Itoa(n.Line)
		break
	case *parser.Dereference:
		y.genDeref(n)
		break
	case *parser.StringConstant:
		y.programm += "\"" + n.Value + "\""
		break
	case *parser.NumberConstant:
		if strings.HasPrefix(n.Value, "-") {
			y.programm += " "
		}
		y.programm += fmt.Sprintf(n.Value)
		break
	case *parser.BinaryOperation:
		y.generateBinaryOperation(n, visitType)
		break
	case *parser.UnaryOperation:
		_, childBinary := n.Exp.(*parser.BinaryOperation)
		if visitType == parser.PreVisit {
			op := n.Operator
			if op == "not" {
				op = " " + op + " "
			}
			if op == "-" {
				op = " " + op
			}
			y.programm += op
			if childBinary {
				y.programm += "("
			}
		}
		if visitType == parser.PostVisit {
			if childBinary {
				y.programm += ")"
			}
		}
		break
	case *parser.FuncCall:
		if visitType == parser.PreVisit {
			y.programm += n.Function + "("
		} else {
			y.programm += ")"
		}
		break
	case *parser.Programm:
		//do noting
		break
	default:
		return fmt.Errorf("Unknown ast-node type: %t", node)
	}
	return nil
}

func (y *YololGenerator) generateBinaryOperation(o *parser.BinaryOperation, visitType int) {
	lPrio := priorityForExpression(o.Exp1)
	rPrio := priorityForExpression(o.Exp2)
	_, rBinary := o.Exp2.(*parser.BinaryOperation)
	myPrio := priorityForExpression(o)
	switch visitType {
	case parser.PreVisit:
		if lPrio < myPrio {
			y.programm += "("
		}
		break
	case parser.InterVisit1:
		if lPrio < myPrio {
			y.programm += ")"
		}
		op := o.Operator
		if op == "and" || op == "or" {
			op = " " + op + " "
		}
		y.programm += op
		if rBinary && rPrio <= myPrio {
			y.programm += "("
		}
		break
	case parser.PostVisit:
		if rBinary && rPrio <= myPrio {
			y.programm += ")"
		}
		break
	}
}

func priorityForExpression(e parser.Expression) int {
	switch ex := e.(type) {
	case *parser.BinaryOperation:
		return operatorPriority[ex.Operator]
	default:
		return 10
	}
}

func (y *YololGenerator) generateIf(visitType int) {
	switch visitType {
	case parser.PreVisit:
		y.programm += "if "
	case parser.InterVisit1:
		y.programm += " then "
	case parser.InterVisit2:
		y.programm += " else "
	case parser.PostVisit:
		y.programm += " end"
	default:
		y.programm += " "
	}
}

func (y *YololGenerator) genDeref(d *parser.Dereference) {
	txt := ""
	if d.PrePost == "Pre" {
		txt += " " + d.Operator
	}
	txt += d.Variable
	if d.PrePost == "Post" {
		txt += d.Operator + " "
	}
	y.programm += txt
}

func (y *YololGenerator) Generate(prog *parser.Programm) string {
	y.programm = ""
	prog.Accept(y)
	// during the generation duplicate spaces might appear. Remove them
	return strings.Replace(y.programm, "  ", " ", -1)
}
