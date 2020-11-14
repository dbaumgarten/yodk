package nast

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
	"github.com/jinzhu/copier"
)

// CopyAst deep-copies the provided (part)-AST
func CopyAst(inp ast.Node) ast.Node {
	found := make(map[ast.Node]bool)
	vfunc := ast.VisitorFunc(func(node ast.Node, visitType int) error {
		if visitType == ast.SingleVisit || visitType == ast.PreVisit {

			if _, copied := found[node]; copied {
				return nil
			}
			var newnode ast.Node

			switch n := node.(type) {
			case *ast.Program:
				m := &ast.Program{}
				copier.Copy(m, n)
				m.Lines = make([]*ast.Line, len(n.Lines))
				copy(m.Lines, n.Lines)
				newnode = m
			case *ast.Line:
				m := &ast.Line{}
				copier.Copy(m, n)
				m.Statements = make([]ast.Statement, len(n.Statements))
				copy(m.Statements, n.Statements)
				newnode = m
			case *ast.Assignment:
				m := &ast.Assignment{}
				copier.Copy(m, n)
				newnode = m
			case *ast.IfStatement:
				m := &ast.IfStatement{}
				copier.Copy(m, n)
				m.IfBlock = make([]ast.Statement, len(n.IfBlock))
				copy(m.IfBlock, n.IfBlock)
				if n.ElseBlock != nil {
					m.ElseBlock = make([]ast.Statement, len(n.ElseBlock))
					copy(m.ElseBlock, n.ElseBlock)
				}
				newnode = m
			case *ast.GoToStatement:
				m := &ast.GoToStatement{}
				copier.Copy(m, n)
				newnode = m
			case *ast.Dereference:
				m := &ast.Dereference{}
				copier.Copy(m, n)
				newnode = m
			case *ast.StringConstant:
				m := &ast.StringConstant{}
				copier.Copy(m, n)
				newnode = m
			case *ast.NumberConstant:
				m := &ast.NumberConstant{}
				copier.Copy(m, n)
				newnode = m
			case *ast.BinaryOperation:
				m := &ast.BinaryOperation{}
				copier.Copy(m, n)
				newnode = m
			case *ast.UnaryOperation:
				m := &ast.UnaryOperation{}
				copier.Copy(m, n)
				newnode = m
			// begin nolol nodes
			case *GoToLabelStatement:
				m := &GoToLabelStatement{}
				copier.Copy(m, n)
				newnode = m
			case *Block:
				m := &Block{}
				copier.Copy(m, n)
				m.Elements = make([]NestableElement, len(n.Elements))
				copy(m.Elements, n.Elements)
				newnode = m
			case *MacroDefinition:
				m := &MacroDefinition{}
				copier.Copy(m, n)
				m.Arguments = make([]string, len(n.Arguments))
				m.Externals = make([]string, len(n.Externals))
				copy(m.Arguments, n.Arguments)
				newnode = m
			case *MacroInsetion:
				m := &MacroInsetion{
					FuncCall: &FuncCall{},
				}
				copier.Copy(m, n)
				newnode = m
			case *FuncCall:
				m := &FuncCall{}
				copier.Copy(m, n)
				m.Arguments = make([]ast.Expression, len(n.Arguments))
				copy(m.Arguments, n.Arguments)
				newnode = m
			case *MultilineIf:
				m := &MultilineIf{}
				copier.Copy(m, n)
				m.Positions = make([]ast.Position, len(n.Positions))
				copy(m.Positions, n.Positions)
				m.Conditions = make([]ast.Expression, len(n.Conditions))
				copy(m.Conditions, n.Conditions)
				m.Blocks = make([]*Block, len(n.Blocks))
				copy(m.Blocks, n.Blocks)
				newnode = m
			case *WhileLoop:
				m := &WhileLoop{}
				copier.Copy(m, n)
				newnode = m
			case *StatementLine:
				m := &StatementLine{
					Line: n.Line,
				}
				copier.Copy(m, n)
				newnode = m
			case *IncludeDirective:
				m := &IncludeDirective{}
				copier.Copy(m, n)
				newnode = m
			case *Definition:
				m := &Definition{}
				copier.Copy(m, n)
				newnode = m
			case *Program:
				m := &Program{}
				copier.Copy(m, n)
				m.Elements = make([]Element, len(n.Elements))
				copy(m.Elements, n.Elements)
				newnode = m
			case *WaitDirective:
				m := &WaitDirective{}
				copier.Copy(m, n)
				newnode = m
			case *BreakStatement:
				m := &BreakStatement{}
				copier.Copy(m, n)
				newnode = m
			case *ContinueStatement:
				m := &ContinueStatement{}
				copier.Copy(m, n)
				newnode = m
			default:
				panic(fmt.Sprintf("Cannot copy unkown type %T", node))
			}

			found[newnode] = true
			return ast.NewNodeReplacement(newnode)

		}
		return nil
	})

	newnode, _ := ast.AcceptChild(vfunc, inp)
	return newnode
}
