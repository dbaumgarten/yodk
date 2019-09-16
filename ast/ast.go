package ast

type Node interface {
}

type Programm struct {
	Lines []*Line
}

type Line struct {
	Statements []Statement
}

// Expressions

type Expression interface {
}

type StringConstant struct {
	Value string
}

type NumberConstant struct {
	Value string
}

type Dereference struct {
	Variable string
	Operator string
	PrePost  string
}

type UnaryOperation struct {
	Operator string
	Exp      Expression
}

type BinaryOperation struct {
	Operator string
	Exp1     Expression
	Exp2     Expression
}

type FuncCall struct {
	Function string
	Argument Expression
}

// Statements

type Statement interface {
}

type Assignment struct {
	Variable string
	Value    Expression
	Operator string
}

type IfStatement struct {
	Condition Expression
	IfBlock   []Statement
	ElseBlock []Statement
}

type GoToStatement struct {
	Line int
}
