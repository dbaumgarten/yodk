package nolol

import "github.com/dbaumgarten/yodk/parser"

type NololProgramm struct {
	Lines []NololLine
}

func (n *NololProgramm) Start() parser.Position {
	return n.Lines[0].Start()
}

func (n *NololProgramm) End() parser.Position {
	return n.Lines[len(n.Lines)-1].End()
}

type NololLine interface {
	parser.Node
}

type ExecutableLine interface {
	NololLine
}

type StatementLine struct {
	parser.Line
	Label    string
	Position parser.Position
}

func (n *StatementLine) Start() parser.Position {
	return n.Position
}

type ConstDeclaration struct {
	Position parser.Position
	Name     string
	Value    parser.Expression
}

func (n *ConstDeclaration) Start() parser.Position {
	return n.Position
}

func (n *ConstDeclaration) End() parser.Position {
	return n.Value.End()
}

type MultilineIf struct {
	Position  parser.Position
	Condition parser.Expression
	IfBlock   []ExecutableLine
	ElseBlock []ExecutableLine
}

func (n *MultilineIf) Start() parser.Position {
	return n.Position
}

func (n *MultilineIf) End() parser.Position {
	if n.ElseBlock == nil {
		return n.IfBlock[len(n.IfBlock)-1].End()
	}
	return n.ElseBlock[len(n.ElseBlock)-1].End()
}

type GoToLabelStatement struct {
	Position parser.Position
	Label    string
}

func (n *GoToLabelStatement) Start() parser.Position {
	return n.Position
}

func (n *GoToLabelStatement) End() parser.Position {
	return n.Position.Add(len(n.Label) + 1)
}

type WhileLoop struct {
	Position  parser.Position
	Condition parser.Expression
	Block     []ExecutableLine
}

func (n *WhileLoop) Start() parser.Position {
	return n.Position
}

func (n *WhileLoop) End() parser.Position {
	return n.Block[len(n.Block)-1].End()
}
