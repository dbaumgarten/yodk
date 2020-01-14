package nast

import "github.com/dbaumgarten/yodk/pkg/parser/ast"

// Program represents a complete programm
type Program struct {
	// The 'lines' of the program a line can also be multiple lines (if, while)
	Lines []Line
	// Ordered list of comments found while parsing this program
	Comments []*ast.Token
}

// Start is needed to implement ast.Node
func (n *Program) Start() ast.Position {
	return n.Lines[0].Start()
}

// End is needed to implement ast.Node
func (n *Program) End() ast.Position {
	return n.Lines[len(n.Lines)-1].End()
}

// Line is the interface for everything that can be a line in nolol
type Line interface {
	ast.Node
}

// ExecutableLine a line that is not compile-time only (not a constant declaration etc.)
type ExecutableLine interface {
	Line
}

// StatementLine is a line consisting of yolol-statements
type StatementLine struct {
	ast.Line
	Label    string
	Position ast.Position
}

// Start is needed to implement ast.Node
func (n *StatementLine) Start() ast.Position {
	return n.Position
}

// ConstDeclaration declares a constant
type ConstDeclaration struct {
	Position ast.Position
	Name     string
	Value    ast.Expression
}

// Start is needed to implement ast.Node
func (n *ConstDeclaration) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *ConstDeclaration) End() ast.Position {
	return n.Value.End()
}

// MultilineIf represents a nolol-style multiline if
type MultilineIf struct {
	Position  ast.Position
	Condition ast.Expression
	IfBlock   []ExecutableLine
	ElseBlock []ExecutableLine
}

// Start is needed to implement ast.Node
func (n *MultilineIf) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *MultilineIf) End() ast.Position {
	if n.ElseBlock == nil {
		return n.IfBlock[len(n.IfBlock)-1].End()
	}
	return n.ElseBlock[len(n.ElseBlock)-1].End()
}

// GoToLabelStatement represents a goto to a line-label
type GoToLabelStatement struct {
	Position ast.Position
	Label    string
}

// Start is needed to implement ast.Node
func (n *GoToLabelStatement) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *GoToLabelStatement) End() ast.Position {
	return n.Position.Add(len(n.Label) + 1)
}

// WhileLoop represents a nolol-style while loop
type WhileLoop struct {
	Position  ast.Position
	Condition ast.Expression
	Block     []ExecutableLine
}

// Start is needed to implement ast.Node
func (n *WhileLoop) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *WhileLoop) End() ast.Position {
	return n.Block[len(n.Block)-1].End()
}
