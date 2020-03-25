package nast

import "github.com/dbaumgarten/yodk/pkg/parser/ast"

// Program represents a complete programm
type Program struct {
	// The 'lines' of the program a line can also be multiple lines (if, while)
	Lines []Line
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
	// If true, do not append this line to any other line during conversion to yolol
	HasBOL bool
	// If true, no other lines may be appended to this line during conversion to yolol
	HasEOL   bool
	Label    string
	Position ast.Position
	Comment  string
}

// Start is needed to implement ast.Node
func (n *StatementLine) Start() ast.Position {
	return n.Position
}

// ConstDeclaration declares a constant
type ConstDeclaration struct {
	Position    ast.Position
	Name        string
	DisplayName string
	Value       ast.Expression
}

// Start is needed to implement ast.Node
func (n *ConstDeclaration) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *ConstDeclaration) End() ast.Position {
	return n.Value.End()
}

// Block represents a block/group of lines, for example inside an if
type Block struct {
	Lines []ExecutableLine
}

// Start is needed to implement ast.Node
func (n *Block) Start() ast.Position {
	return n.Lines[0].Start()
}

// End is needed to implement ast.Node
func (n *Block) End() ast.Position {
	return n.Lines[len(n.Lines)-1].End()
}

// MultilineIf represents a nolol-style multiline if
type MultilineIf struct {
	Position   ast.Position
	Conditions []ast.Expression
	Blocks     []*Block
	ElseBlock  *Block
}

// Start is needed to implement ast.Node
func (n *MultilineIf) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *MultilineIf) End() ast.Position {
	if n.ElseBlock == nil {
		return n.Blocks[len(n.Blocks)-1].End()
	}
	return n.ElseBlock.End()
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
	Block     *Block
}

// Start is needed to implement ast.Node
func (n *WhileLoop) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *WhileLoop) End() ast.Position {
	return n.Block.End()
}

// WaitStatement blocks execution as long as the condition is true
type WaitStatement struct {
	Position  ast.Position
	Condition ast.Expression
}

// Start is needed to implement ast.Node
func (n *WaitStatement) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *WaitStatement) End() ast.Position {
	return n.Condition.End()
}

// IncludeDirective represents the inclusion of another file in the source-file
type IncludeDirective struct {
	Position ast.Position
	File     string
}

// Start is needed to implement ast.Node
func (n *IncludeDirective) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *IncludeDirective) End() ast.Position {
	return n.Position.Add(len(n.File) + 3 + len("include"))
}
