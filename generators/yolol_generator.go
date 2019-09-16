package generators

import (
	"fmt"
	"strconv"

	"github.com/dbaumgarten/yodk/ast"
)

type YololGenerator struct {
	programm string
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
		y.programm += fmt.Sprintf(n.Value)
		break
	case *ast.BinaryOperation:
		if visitType == ast.PreVisit {
			y.programm += "("
		}
		if visitType == ast.PostVisit {
			y.programm += ")"
		}
		if visitType == ast.InterVisit1 {
			op := n.Operator
			if op == "and" || op == "or" {
				op = " " + op + " "
			}
			y.programm += op
		}
		break
	case *ast.UnaryOperation:
		if visitType == ast.PreVisit {
			op := n.Operator
			if op == "not" {
				op = " " + op + " "
			}
			if op == "-" {
				op = " " + op
			}
			y.programm += op
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
	if d.PrePost != "" && !d.IsStatement {
		txt += "("
	}
	if d.PrePost == "Pre" {
		txt += d.Operator
	}
	txt += d.Variable
	if d.PrePost == "Post" {
		txt += d.Operator
	}
	if d.PrePost != "" && !d.IsStatement {
		txt += ")"
	}
	y.programm += txt
}

func (y *YololGenerator) Generate(prog *ast.Programm) string {
	y.programm = ""
	prog.Accept(y)
	return y.programm
}
