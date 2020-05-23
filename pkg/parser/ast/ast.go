package ast

import (
	"strconv"
)

// Node is the base interface
type Node interface {
	Acceptor
	// Start returns the start-position of the node in the source-code
	Start() Position
	// End returns the end-position of the node in the source-code
	End() Position
}

// Program represents the whole yolol-programm
type Program struct {
	Lines []*Line
}

// Start is needed to implement Node
func (n *Program) Start() Position {
	return n.Lines[0].Start()
}

// End is needed to implement Node
func (n *Program) End() Position {
	return n.Lines[len(n.Lines)-1].End()
}

// Line represents a line in the yolol programm
type Line struct {
	Position   Position
	Statements []Statement
	// A Line can have a comment at the end or be just a comment without statements
	Comment string
}

// Start is needed to implement Node
func (n *Line) Start() Position {
	return n.Position
}

// End is needed to implement Node
func (n *Line) End() Position {
	if len(n.Statements) == 0 {
		return n.Position.Add(len(n.Comment))
	}
	return n.Statements[len(n.Statements)-1].End().Add(len(n.Comment) + 1)
}

// Expression is the interface for all expressions
type Expression interface {
	Node
}

// StringConstant represents a constant of type string
type StringConstant struct {
	Position Position
	// Value of the constant
	Value string
}

// Start is needed to implement Node
func (n *StringConstant) Start() Position {
	return n.Position
}

// End is needed to implement Node
func (n *StringConstant) End() Position {
	pos := n.Start()
	pos.Coloumn += len(n.Value) + 2
	return pos
}

// NumberConstant represents a constant of type number
type NumberConstant struct {
	Position Position
	// Value of the constant
	Value string
}

// Start is needed to implement Node
func (n *NumberConstant) Start() Position {
	return n.Position
}

// End is needed to implement Node
func (n *NumberConstant) End() Position {
	return n.Start().Add(len(n.Value))
}

// Dereference represents the dereferencing of a variable
type Dereference struct {
	Position Position
	// The name of the dereferenced variable
	Variable string
	// Additional operator (++ or --)
	Operator string
	// Wheter to use the Operator as Pre- or Postoperator
	PrePost string
	// True if this is used as a statement instead of expression
	IsStatement bool
}

// Start is needed to implement Node
func (n *Dereference) Start() Position {
	return n.Position
}

// End is needed to implement Node
func (n *Dereference) End() Position {
	return n.Start().Add(len(n.Variable) + len(n.Operator))
}

// UnaryOperation represents a unary operation (-, not)
type UnaryOperation struct {
	Position Position
	Operator string
	Exp      Expression
}

// Start is needed to implement Node
func (n *UnaryOperation) Start() Position {
	return n.Position
}

// End is needed to implement Node
func (n *UnaryOperation) End() Position {
	return n.Exp.End()
}

// BinaryOperation is a binary operation
type BinaryOperation struct {
	Operator string
	Exp1     Expression
	Exp2     Expression
}

// Start is needed to implement Node
func (n *BinaryOperation) Start() Position {
	return n.Exp1.Start()
}

// End is needed to implement Node
func (n *BinaryOperation) End() Position {
	return n.Exp2.End()
}

// Statement is the interface for all statements
type Statement interface {
	Node
}

// Assignment represents the assignment to a variable
type Assignment struct {
	Position Position
	// The name of the variable that is assigned to
	Variable string
	// The value to be assigned
	Value Expression
	// Operator to use (=,+=,-=, etc.)
	Operator string
}

// Start is needed to implement Node
func (n *Assignment) Start() Position {
	return n.Position
}

// End is needed to implement Node
func (n *Assignment) End() Position {
	if n.Value.End().Before(n.Position) {
		return n.Position.Add(70)
	}
	return n.Value.End()
}

// IfStatement represents an if-statement
type IfStatement struct {
	Position Position
	// Condition for the if
	Condition Expression
	// Statements to execute if true
	IfBlock []Statement
	// Statements to execute if false
	ElseBlock []Statement
}

// Start is needed to implement Node
func (n *IfStatement) Start() Position {
	return n.Position
}

// End is needed to implement Node
func (n *IfStatement) End() Position {
	if n.ElseBlock == nil {
		return n.IfBlock[len(n.IfBlock)-1].End().Add(3)
	}
	return n.ElseBlock[len(n.ElseBlock)-1].End().Add(3)
}

// GoToStatement represents a goto
type GoToStatement struct {
	Position Position
	// Number of the line to go to
	Line int
}

// Start is needed to implement Node
func (n *GoToStatement) Start() Position {
	return n.Position
}

// End is needed to implement Node
func (n *GoToStatement) End() Position {
	return n.Position.Add(len(strconv.Itoa(n.Line)) + 1)
}
