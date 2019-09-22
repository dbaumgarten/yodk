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

type ExtLine struct {
	Line
	Label string
}

type ExtProgramm struct {
	ExecutableLines []*ExtLine
}

func (n *ExtProgramm) Start() Position {
	return n.ExecutableLines[0].Start()
}

func (n *ExtProgramm) End() Position {
	return n.ExecutableLines[len(n.ExecutableLines)-1].End()
}
