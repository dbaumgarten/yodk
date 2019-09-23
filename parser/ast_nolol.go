package parser

type GoToLabelStatement struct {
	Position Position
	Label    string
}

func (n *GoToLabelStatement) Start() Position {
	return n.Position
}

func (n *GoToLabelStatement) End() Position {
	return n.Position.Add(len(n.Label) + 1)
}

type ExtLine interface {
	Node
}

type ExecutableLine struct {
	Line
	Label    string
	Position Position
}

func (n *ExecutableLine) Start() Position {
	return n.Position
}

type ExtProgramm struct {
	ExecutableLines []ExtLine
}

func (n *ExtProgramm) Start() Position {
	return n.ExecutableLines[0].Start()
}

func (n *ExtProgramm) End() Position {
	return n.ExecutableLines[len(n.ExecutableLines)-1].End()
}

type ConstDeclaration struct {
	Position Position
	Name     string
	Value    Expression
}

func (n *ConstDeclaration) Start() Position {
	return n.Position
}

func (n *ConstDeclaration) End() Position {
	return n.Value.End()
}
