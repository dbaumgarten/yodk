package nolol

import "github.com/dbaumgarten/yodk/pkg/parser"

// Program represents a complete programm
type Program struct {
	// The 'lines' of the program a line can also be multiple lines (if, while)
	Lines []Line
	// Ordered list of comments found while parsing this program
	Comments []*parser.Token
}

// Start is needed to implement parser.Node
func (n *Program) Start() parser.Position {
	return n.Lines[0].Start()
}

// End is needed to implement parser.Node
func (n *Program) End() parser.Position {
	return n.Lines[len(n.Lines)-1].End()
}

// Line is the interface for everything that can be a line in nolol
type Line interface {
	parser.Node
}

// ExecutableLine a line that is not compile-time only (not a constant declaration etc.)
type ExecutableLine interface {
	Line
}

// StatementLine is a line consisting of yolol-statements
type StatementLine struct {
	parser.Line
	Label    string
	Position parser.Position
}

// Start is needed to implement parser.Node
func (n *StatementLine) Start() parser.Position {
	return n.Position
}

// ConstDeclaration declares a constant
type ConstDeclaration struct {
	Position parser.Position
	Name     string
	Value    parser.Expression
}

// Start is needed to implement parser.Node
func (n *ConstDeclaration) Start() parser.Position {
	return n.Position
}

// End is needed to implement parser.Node
func (n *ConstDeclaration) End() parser.Position {
	return n.Value.End()
}

// MultilineIf represents a nolol-style multiline if
type MultilineIf struct {
	Position  parser.Position
	Condition parser.Expression
	IfBlock   []ExecutableLine
	ElseBlock []ExecutableLine
}

// Start is needed to implement parser.Node
func (n *MultilineIf) Start() parser.Position {
	return n.Position
}

// End is needed to implement parser.Node
func (n *MultilineIf) End() parser.Position {
	if n.ElseBlock == nil {
		return n.IfBlock[len(n.IfBlock)-1].End()
	}
	return n.ElseBlock[len(n.ElseBlock)-1].End()
}

// GoToLabelStatement represents a goto to a line-label
type GoToLabelStatement struct {
	Position parser.Position
	Label    string
}

// Start is needed to implement parser.Node
func (n *GoToLabelStatement) Start() parser.Position {
	return n.Position
}

// End is needed to implement parser.Node
func (n *GoToLabelStatement) End() parser.Position {
	return n.Position.Add(len(n.Label) + 1)
}

// WhileLoop represents a nolol-style while loop
type WhileLoop struct {
	Position  parser.Position
	Condition parser.Expression
	Block     []ExecutableLine
}

// Start is needed to implement parser.Node
func (n *WhileLoop) Start() parser.Position {
	return n.Position
}

// End is needed to implement parser.Node
func (n *WhileLoop) End() parser.Position {
	return n.Block[len(n.Block)-1].End()
}
