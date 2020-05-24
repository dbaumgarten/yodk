package nast

import "github.com/dbaumgarten/yodk/pkg/parser/ast"

// Program represents a complete programm
type Program struct {
	// The parts of the program
	Elements []Element
}

// Start is needed to implement ast.Node
func (n *Program) Start() ast.Position {
	return n.Elements[0].Start()
}

// End is needed to implement ast.Node
func (n *Program) End() ast.Position {
	return n.Elements[len(n.Elements)-1].End()
}

// Element is a top-level part of a nolol-program. This is everything that can appear stand-alone
// inside a nolol program
type Element interface {
	ast.Node
}

// NestableElement describes a special kind of element, that can appear inside blocks (and not only on the top-level)
type NestableElement interface {
	Element
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

// Definition declares a constant
type Definition struct {
	Position ast.Position
	Name     string
	Value    ast.Expression
}

// Start is needed to implement ast.Node
func (n *Definition) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *Definition) End() ast.Position {
	return n.Value.End()
}

// Block represents a block/group of elements, for example inside an if
type Block struct {
	Elements []NestableElement
}

// Start is needed to implement ast.Node
func (n *Block) Start() ast.Position {
	return n.Elements[0].Start()
}

// End is needed to implement ast.Node
func (n *Block) End() ast.Position {
	return n.Elements[len(n.Elements)-1].End()
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

// WaitDirective blocks execution as long as the condition is true
type WaitDirective struct {
	Position   ast.Position
	Condition  ast.Expression
	Statements []ast.Statement
}

// Start is needed to implement ast.Node
func (n *WaitDirective) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *WaitDirective) End() ast.Position {
	if n.Statements == nil || len(n.Statements) == 0 {
		return n.Condition.End()
	}
	return n.Statements[len(n.Statements)-1].End()
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

// MacroDefinition represents the definition of a macro
type MacroDefinition struct {
	Position  ast.Position
	Name      string
	Arguments []string
	Block     *Block
}

// Start is needed to implement ast.Node
func (n *MacroDefinition) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *MacroDefinition) End() ast.Position {
	return n.Block.End()
}

// MacroInsetion represents the use of a macro
// To reduce code-duplication it re-uses the FuncCall-type
type MacroInsetion struct {
	Position ast.Position
	*FuncCall
}

// Start is needed to implement ast.Node
func (n *MacroInsetion) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *MacroInsetion) End() ast.Position {
	return n.FuncCall.End()
}

// Trigger is a special kind of node, that is sometimes inserted during code-generation
// It is used to tigger certain events when reached by a visitor and is created typically by nodes that
// replace themselves, but want to perform a certain action when a specific point in the ast is visited again.
type Trigger struct {
	Kind string
}

// Start is needed to implement ast.Node
func (n *Trigger) Start() ast.Position {
	return ast.Position{}
}

// End is needed to implement ast.Node
func (n *Trigger) End() ast.Position {
	return ast.Position{}
}

// FuncCall represents a func-call
type FuncCall struct {
	Position  ast.Position
	Function  string
	Arguments []ast.Expression
}

// Start is needed to implement Node
func (n *FuncCall) Start() ast.Position {
	return n.Position
}

// End is needed to implement Node
func (n *FuncCall) End() ast.Position {
	if n.Arguments != nil && len(n.Arguments) > 0 {
		return n.Arguments[len(n.Arguments)-1].End().Add(1)
	}
	return n.Position.Add(len(n.Function) + 2)
}

// BreakStatement represents the break-keyword inside a loop
type BreakStatement struct {
	Position ast.Position
}

// Start is needed to implement ast.Node
func (n *BreakStatement) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *BreakStatement) End() ast.Position {
	return n.Position.Add(len("break"))
}

// ContinueStatement represents the continue-keyword inside a loop
type ContinueStatement struct {
	Position ast.Position
}

// Start is needed to implement ast.Node
func (n *ContinueStatement) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *ContinueStatement) End() ast.Position {
	return n.Position.Add(len("continue"))
}
