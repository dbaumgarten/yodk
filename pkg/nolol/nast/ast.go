package nast

import "github.com/dbaumgarten/yodk/pkg/parser/ast"

// Program represents a complete programm
type Program struct {
	// The parts of the program
	Elements []Element
}

// Start is needed to implement ast.Node
func (n *Program) Start() ast.Position {
	if len(n.Elements) == 0 {
		return ast.UnknownPosition
	}
	return n.Elements[0].Start()
}

// End is needed to implement ast.Node
func (n *Program) End() ast.Position {
	if len(n.Elements) == 0 {
		return ast.UnknownPosition
	}
	return n.Elements[len(n.Elements)-1].End()
}

// Element is a top-level part of a nolol-program. This is everything that can appear stand-alone
// inside a nolol program
type Element interface {
	ast.Node
	// Dummy marker-method
	El()
}

// NestableElement describes a special kind of element, that can appear inside blocks (and not only on the top-level)
type NestableElement interface {
	Element
	// Dummy marker-method
	NestEl()
}

// StatementLine is a line consisting of yolol-statements
type StatementLine struct {
	ast.Line
	// If true, do not append this line to any other line during conversion to yolol
	HasBOL bool
	// If true, no other lines may be appended to this line during conversion to yolol
	HasEOL  bool
	Label   string
	Comment string
}

// El implements the type-marker method
func (n *StatementLine) El() {}

// NestEl implements the type-marker method
func (n *StatementLine) NestEl() {}

// Definition declares a constant
type Definition struct {
	Position     ast.Position
	Name         string
	Placeholders []string
	Value        ast.Expression
}

// Start is needed to implement ast.Node
func (n *Definition) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *Definition) End() ast.Position {
	if n.Value == nil {
		return n.Position
	}
	return n.Value.End()
}

// El implements the type-marker method
func (n *Definition) El() {}

// Block represents a block/group of elements, for example inside an if
type Block struct {
	Elements []NestableElement
}

// Start is needed to implement ast.Node
func (n *Block) Start() ast.Position {
	if len(n.Elements) == 0 {
		return ast.UnknownPosition
	}
	return n.Elements[0].Start()
}

// End is needed to implement ast.Node
func (n *Block) End() ast.Position {
	if len(n.Elements) == 0 {
		return ast.UnknownPosition
	}
	return n.Elements[len(n.Elements)-1].End()
}

// El implements the type-marker method
func (n *Block) El() {}

// NestEl implements the type-marker method
func (n *Block) NestEl() {}

// MultilineIf represents a nolol-style multiline if
type MultilineIf struct {
	Positions  []ast.Position
	Conditions []ast.Expression
	Blocks     []*Block
	ElseBlock  *Block
}

// Start is needed to implement ast.Node
func (n *MultilineIf) Start() ast.Position {
	if len(n.Positions) == 0 {
		return ast.UnknownPosition
	}
	return n.Positions[0]
}

// End is needed to implement ast.Node
func (n *MultilineIf) End() ast.Position {
	if n.ElseBlock == nil {
		if len(n.Blocks) == 0 {
			return ast.UnknownPosition
		}
		return n.Blocks[len(n.Blocks)-1].End()
	}
	return n.ElseBlock.End()
}

// El implements the type-marker method
func (n *MultilineIf) El() {}

// NestEl implements the type-marker method
func (n *MultilineIf) NestEl() {}

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
	if n.Block == nil {
		return n.Position
	}
	return n.Block.End()
}

// El implements the type-marker method
func (n *WhileLoop) El() {}

// NestEl implements the type-marker method
func (n *WhileLoop) NestEl() {}

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
	if len(n.Statements) == 0 {
		if n.Condition == nil {
			return n.Position
		}
		return n.Condition.End()
	}
	return n.Statements[len(n.Statements)-1].End()
}

// El implements the type-marker method
func (n *WaitDirective) El() {}

// NestEl implements the type-marker method
func (n *WaitDirective) NestEl() {}

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

// El implements the type-marker method
func (n *IncludeDirective) El() {}

// MacroDefinition represents the definition of a macro
type MacroDefinition struct {
	Position  ast.Position
	Name      string
	Arguments []string
	Externals []string
	Block     *Block
}

// Start is needed to implement ast.Node
func (n *MacroDefinition) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *MacroDefinition) End() ast.Position {
	if n.Block == nil {
		return n.Position
	}
	return n.Block.End()
}

// El implements the type-marker method
func (n *MacroDefinition) El() {}

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
	if n.FuncCall == nil {
		return n.Position
	}
	return n.FuncCall.End()
}

// El implements the type-marker method
func (n *MacroInsetion) El() {}

// NestEl implements the type-marker method
func (n *MacroInsetion) NestEl() {}

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

// El implements the type-marker method
func (n *Trigger) El() {}

// NestEl implements the type-marker method
func (n *Trigger) NestEl() {}

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

// Expr implements type-checking dummy-func
func (n *FuncCall) Expr() {}

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

// Stmt implements type-checking dummy-func
func (n *BreakStatement) Stmt() {}

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

// Stmt implements type-checking dummy-func
func (n *ContinueStatement) Stmt() {}
