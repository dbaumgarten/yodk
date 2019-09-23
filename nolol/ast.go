package nolol

import "github.com/dbaumgarten/yodk/parser"

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

type ExtLine interface {
	parser.Node
}

type ExecutableLine struct {
	parser.Line
	Label    string
	Position parser.Position
}

func (n *ExecutableLine) Start() parser.Position {
	return n.Position
}

type ExtProgramm struct {
	ExecutableLines []ExtLine
}

func (n *ExtProgramm) Start() parser.Position {
	return n.ExecutableLines[0].Start()
}

func (n *ExtProgramm) End() parser.Position {
	return n.ExecutableLines[len(n.ExecutableLines)-1].End()
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
