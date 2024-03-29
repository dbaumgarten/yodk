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

// The different types of macros
const (
	MacroTypeExpr  = "expr"
	MacroTypeLine  = "line"
	MacroTypeBlock = "block"
)

// MacroDefinition represents the definition of a macro
type MacroDefinition struct {
	Position  ast.Position
	Name      string
	Arguments []string
	Externals []string
	Type      string
	// Depending on Type, Code can either be ast.Expression, *nast.StatementLine or *nast.Block
	Code ast.Node
	// Macros of type expr and line do not have StatementLines that would hold comments
	// So we need to store the comments before and after the real content of the macro manually
	PreComments  []string
	PostComments []string
}

// Start is needed to implement ast.Node
func (n *MacroDefinition) Start() ast.Position {
	return n.Position
}

// End is needed to implement ast.Node
func (n *MacroDefinition) End() ast.Position {
	if n.Code == nil {
		return n.Position
	}
	return n.Code.End()
}

// El implements the type-marker method
func (n *MacroDefinition) El() {}

// FuncCall represents a func-call
type FuncCall struct {
	Position  ast.Position
	Function  string
	Arguments []ast.Expression
	// A funccall can be of type block, line or statement
	Type string
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

// El implements type-checking dummy-func
func (n *FuncCall) El() {}

// NestEl implements type-checking dummy-func
func (n *FuncCall) NestEl() {}

// Expr implements type-checking dummy-func
func (n *FuncCall) Expr() {}

// Stmt implements type-checking dummy-func
func (n *FuncCall) Stmt() {}

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
