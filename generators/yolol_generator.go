package generators

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/ast"
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

func (y *YololGenerator) Visit(node ast.Node, visitType int) error {
	switch n := node.(type) {
	case *ast.Line:
		if visitType == ast.PostVisit {
			y.programm += "\n"
		}
		if visitType > 0 {
			y.programm += " "
		}
		break
	case *ast.Assignment:
		if visitType == ast.PreVisit {
			y.programm += n.Variable + n.Operator
		}
		break
	case *ast.IfStatement:
		y.generateIf(visitType)
		break
	case *ast.GoToStatement:
		y.programm += "goto " + strconv.Itoa(n.Line)
		break
	case *ast.Dereference:
		y.genDeref(n)
		break
	case *ast.StringConstant:
		y.programm += "\"" + n.Value + "\""
		break
	case *ast.NumberConstant:
		if strings.HasPrefix(n.Value, "-") {
			y.programm += " "
		}
		y.programm += fmt.Sprintf(n.Value)
		break
	case *ast.BinaryOperation:
		y.generateBinaryOperation(n, visitType)
		break
	case *ast.UnaryOperation:
		_, childBinary := n.Exp.(*ast.BinaryOperation)
		if visitType == ast.PreVisit {
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
		if visitType == ast.PostVisit {
			if childBinary {
				y.programm += ")"
			}
		}
		break
	case *ast.FuncCall:
		if visitType == ast.PreVisit {
			y.programm += n.Function + "("
		} else {
			y.programm += ")"
		}
		break
	case *ast.Programm:
		//do noting
		break
	default:
		return fmt.Errorf("Unknown ast-node type: %t", node)
	}
	return nil
}

func (y *YololGenerator) generateBinaryOperation(o *ast.BinaryOperation, visitType int) {
	lPrio := priorityForExpression(o.Exp1)
	rPrio := priorityForExpression(o.Exp2)
	_, rBinary := o.Exp2.(*ast.BinaryOperation)
	myPrio := priorityForExpression(o)
	switch visitType {
	case ast.PreVisit:
		if lPrio < myPrio {
			y.programm += "("
		}
		break
	case ast.InterVisit1:
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
	case ast.PostVisit:
		if rBinary && rPrio <= myPrio {
			y.programm += ")"
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

func (y *YololGenerator) generateIf(visitType int) {
	switch visitType {
	case ast.PreVisit:
		y.programm += "if "
	case ast.InterVisit1:
		y.programm += " then "
	case ast.InterVisit2:
		y.programm += " else "
	case ast.PostVisit:
		y.programm += " end"
	default:
		y.programm += " "
	}
}

func (y *YololGenerator) genDeref(d *ast.Dereference) {
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

func (y *YololGenerator) Generate(prog *ast.Programm) string {
	y.programm = ""
	prog.Accept(y)
	// during the generation duplicate spaces might appear. Remove them
	return strings.Replace(y.programm, "  ", " ", -1)
}
