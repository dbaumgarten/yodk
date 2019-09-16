package generators

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/ast"
)

type YololGenerator struct {
}

func (y *YololGenerator) Generate(prog *ast.Programm) string {
	out := ""
	for _, line := range prog.Lines {
		out += y.genLine(line) + "\n"
	}
	return out
}

func (y *YololGenerator) genLine(line *ast.Line) string {
	out := ""
	for _, stmt := range line.Statements {
		out += y.genStmt(stmt) + " "
	}
	return out
}

func (y *YololGenerator) genStmt(stmt ast.Statement) string {
	switch v := stmt.(type) {
	case *ast.Assignment:
		return v.Variable + v.Operator + y.genExpr(v.Value)
	case *ast.IfStatement:
		txt := "if " + y.genExpr(v.Condition) + " then "
		for _, st := range v.IfBlock {
			txt += y.genStmt(st) + " "
		}
		if v.ElseBlock != nil {
			txt += "else "
			for _, st := range v.ElseBlock {
				txt += y.genStmt(st) + " "
			}
		}
		txt += "end"
		return txt
	case *ast.GoToStatement:
		return "goto " + strconv.Itoa(v.Line)
	case *ast.Dereference:
		return strings.Trim(y.genDeref(v), "()")
	default:
		return fmt.Sprintf("UNKNWON-STATEMENT:%T", v)
	}
}

func (y *YololGenerator) genExpr(expr ast.Expression) string {
	switch v := expr.(type) {
	case *ast.StringConstant:
		return "\"" + v.Value + "\""
	case *ast.NumberConstant:
		return fmt.Sprintf(v.Value)
	case *ast.BinaryOperation:
		op := v.Operator
		if op == "and" || op == "or" {
			op = " " + op + " "
		}
		return "(" + y.genExpr(v.Exp1) + op + y.genExpr(v.Exp2) + ")"
	case *ast.UnaryOperation:
		op := v.Operator
		if op == "not" {
			op = " " + op + " "
		}
		if op == "-" {
			op = " " + op
		}
		return op + y.genExpr(v.Exp)
	case *ast.Dereference:
		return y.genDeref(v)
	case *ast.FuncCall:
		return v.Function + "(" + y.genExpr(v.Argument) + ")"
	default:
		return fmt.Sprintf("UNKNWON-EXPRESSION:%T", v)
	}
}

func (y *YololGenerator) genDeref(d *ast.Dereference) string {
	txt := ""
	if d.PrePost != "" {
		txt += "("
	}
	if d.PrePost == "Pre" {
		txt += d.Operator
	}
	txt += d.Variable
	if d.PrePost == "Post" {
		txt += d.Operator
	}
	if d.PrePost != "" {
		txt += ")"
	}
	return txt
}
