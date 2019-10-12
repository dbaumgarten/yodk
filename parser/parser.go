package parser

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// Parser parses a yolol programm into an AST
type Parser struct {
	DebugLog     bool
	Tokenizer    *Tokenizer
	Tokens       []*Token
	CurrentToken int
	// using an interface of ourself to call the parsing-methods allows them to be overridden by 'subclasses'
	This YololParserFunctions
}

// YololParserFunctions is used together with Parser.This to allow 'subclasses' to override 'virtual functions'
type YololParserFunctions interface {
	ParseStatement() (Statement, *Error)
	ParsePreOrPostOperation() (Statement, *Error)
	ParseGoto() (Statement, *Error)
	ParseAssignment() (Statement, *Error)
	ParseIf() (Statement, *Error)
	ParseExpression() (Expression, *Error)
	ParseLogicExpression() (Expression, *Error)
	ParseCompareExpression() (Expression, *Error)
	ParseSumExpression() (Expression, *Error)
	ParseProdExpression() (Expression, *Error)
	ParseUnaryExpression() (Expression, *Error)
	ParseBracketExpression() (Expression, *Error)
	ParseSingleExpression() (Expression, *Error)
	ParseFuncCall() (Expression, *Error)
	ParsePreOpExpression() (Expression, *Error)
	ParsePostOpExpression() (Expression, *Error)
}

// NewParser creates a new parser
func NewParser() *Parser {
	p := &Parser{
		Tokenizer: NewTokenizer(),
	}
	p.This = p
	return p
}

// HasNext returns true if there is a next token
func (p *Parser) HasNext() bool {
	return p.CurrentToken < len(p.Tokens)-1
}

// Advance advances the current token to the next (non whitespace) token in the list
func (p *Parser) Advance() *Token {
	if p.CurrentToken < len(p.Tokens)-1 {
		p.CurrentToken++
	}
	// skip whitespace
	for p.CurrentToken < len(p.Tokens)-1 && p.Tokens[p.CurrentToken].Type == TypeWhitespace {
		p.CurrentToken++
	}
	return p.Tokens[p.CurrentToken]
}

func (p *Parser) prevWithWhitespace() *Token {
	if p.CurrentToken-1 > 0 {
		return p.Tokens[p.CurrentToken-1]
	}
	return p.Tokens[p.CurrentToken]
}

// Current returns the current token
func (p *Parser) Current() *Token {
	offset := 0
	// skip whitespace
	for p.CurrentToken+offset < len(p.Tokens) && p.Tokens[p.CurrentToken+offset].Type == TypeWhitespace {
		offset++
	}
	return p.Tokens[p.CurrentToken+offset]
}

// Next returns the next token
func (p *Parser) Next() *Token {
	offset := 1
	// skip whitespace
	for p.CurrentToken+offset < len(p.Tokens) && p.Tokens[p.CurrentToken+offset].Type == TypeWhitespace {
		offset++
	}
	if p.CurrentToken+offset < len(p.Tokens) {
		return p.Tokens[p.CurrentToken+offset]
	}
	return p.Tokens[p.CurrentToken]
}

// Parse is the main method of the parser. Parses a yolol-program into an AST.
func (p *Parser) Parse(prog string) (*Program, error) {
	errors := make(Errors, 0)
	p.Tokenizer.Load(prog)
	p.Tokens = make([]*Token, 0, 1000)
	for {
		token, err := p.Tokenizer.Next()
		if err != nil {
			errors = append(errors, err.(*Error))
		} else {
			if p.DebugLog {
				fmt.Print(token)
			}
			p.Tokens = append(p.Tokens, token)
			if token.Type == TypeEOF {
				break
			}
		}
	}
	p.CurrentToken = 0
	parsed, err := p.ParseProgram()
	errors = append(errors, err...)
	if len(errors) > 0 {
		return nil, errors
	}
	return parsed, nil
}

// ParseProgram parses a programm-node
func (p *Parser) ParseProgram() (*Program, Errors) {
	p.Log()
	errors := make(Errors, 0)
	ret := Program{
		Lines: make([]*Line, 0),
	}
	for p.HasNext() {
		line, err := p.ParseLine()
		if err != nil {
			errors = append(errors, err)
			p.SkipLine()
		}
		ret.Lines = append(ret.Lines, line)
	}
	return &ret, errors
}

// SkipLine skips tokens up to the next newline
func (p *Parser) SkipLine() {
	for p.Current().Type != TypeNewline && p.Current().Type != TypeEOF {
		p.Advance()
	}
}

// ParseLine parses a line-node
func (p *Parser) ParseLine() (*Line, *Error) {
	p.Log()
	ret := Line{
		Statements: make([]Statement, 0),
	}

	for p.HasNext() {
		if p.Current().Type == TypeNewline || p.Current().Type == TypeEOF {
			p.Advance()
			return &ret, nil
		}
		stmt, err := p.This.ParseStatement()
		if err != nil {
			return nil, err
		}
		ret.Statements = append(ret.Statements, stmt)
	}

	if p.Current().Type == TypeEOF {
		return &ret, nil
	}

	return nil, p.NewError("Missing newline", true, ret.Start(), ret.End())
}

// ParseStatement parses a statement-node
func (p *Parser) ParseStatement() (Statement, *Error) {
	p.Log()

	// the only place where whitespace can not be ignored
	if p.Current().Position.Coloumn != 1 && p.prevWithWhitespace().Type != TypeWhitespace {
		return nil, p.NewError("Statements on same line must be seperated by space", true, p.Current().Position, p.Current().Position)
	}

	stmt, err := p.This.ParseAssignment()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return stmt, nil
	}

	stmt2, err2 := p.This.ParseIf()
	if err2 != nil && err2.Fatal {
		return nil, err2
	}
	if err2 == nil {
		return stmt2, nil
	}

	stmt3, err3 := p.This.ParseGoto()
	if err3 != nil && err3.Fatal {
		return nil, err3
	}
	if err3 == nil {
		return stmt3, nil
	}

	stmt4, err4 := p.This.ParsePreOrPostOperation()
	if err4 != nil && err4.Fatal {
		return nil, err4
	}
	if err4 == nil {
		return stmt4, nil
	}

	return nil, p.NewError("Expected assignment, if-statement or goto", false, p.Current().Position, p.Current().Position)
}

// ParsePreOrPostOperation parses a pre-/post operation (x++, ++x) as a statement
func (p *Parser) ParsePreOrPostOperation() (Statement, *Error) {
	p.Log()
	preOpVarDeref, err := p.This.ParsePreOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		preOpVarDeref.(*Dereference).IsStatement = true
		return preOpVarDeref, nil
	}

	postOpVarDeref, err := p.This.ParsePostOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		postOpVarDeref.(*Dereference).IsStatement = true
		return postOpVarDeref, nil
	}
	return nil, p.NewError("No Pre- or Post expression-statement found", false, p.Current().Position, p.Current().Position)
}

// ParseGoto parse parses a goto-node
func (p *Parser) ParseGoto() (Statement, *Error) {
	if p.Current().Type == TypeKeyword && p.Current().Value == "goto" {
		stmt := GoToStatement{
			Position: p.Current().Position,
		}
		p.Advance()
		if p.Current().Type != TypeNumber {
			return nil, p.NewError("Goto must be followed by a line number", true, stmt.Start(), stmt.Start())
		}
		line, err := strconv.Atoi(p.Current().Value)
		stmt.Line = line
		if err != nil {
			return nil, p.NewError("Goto must be followed by a line number", true, stmt.Start(), stmt.End())
		}
		p.Advance()
		return &stmt, nil
	}
	return nil, p.NewError("Goto statements must start with 'goto'", false, p.Current().Position, p.Current().Position)
}

// ParseAssignment parses an assignment-node
func (p *Parser) ParseAssignment() (Statement, *Error) {
	p.Log()
	assignmentOperators := []string{"=", "+=", "-=", "*=", "/=", "%="}
	ret := Assignment{
		Position: p.Current().Position,
	}
	if p.Current().Type != TypeID || !contains(assignmentOperators, p.Next().Value) {
		return nil, p.NewError("Expected identifier and assignment operator", false, p.Current().Position, p.Next().Position)
	}
	ret.Variable = p.Current().Value
	p.Advance()
	ret.Operator = p.Current().Value
	p.Advance()
	exp, err := p.This.ParseExpression()
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("Expected expression on right side of assignment"))
	}
	ret.Value = exp
	return &ret, nil
}

// ParseIf parses an if-node
func (p *Parser) ParseIf() (Statement, *Error) {
	p.Log()
	ret := IfStatement{
		Position: p.Current().Position,
	}
	if p.Current().Type != TypeKeyword || p.Current().Value != "if" {
		return nil, p.NewError("If-statements have to start with 'if'", false, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	var err *Error
	ret.Condition, err = p.This.ParseExpression()
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("No expression found as if-condition"))
	}

	if p.Current().Type != TypeKeyword || p.Current().Value != "then" {
		return nil, p.NewError("Expected 'then' after condition", true, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	stmt, err := p.This.ParseStatement()
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("Expected statement after 'then'"))
	}
	ret.IfBlock = make([]Statement, 0, 1)
	ret.IfBlock = append(ret.IfBlock, stmt)

	for {
		stmt2, err := p.This.ParseStatement()
		if err != nil {
			break
		}
		ret.IfBlock = append(ret.IfBlock, stmt2)
	}

	if p.Current().Type == TypeKeyword && p.Current().Value == "else" {
		p.Advance()
		stmt, err := p.This.ParseStatement()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected statement after 'else'"))
		}
		ret.ElseBlock = make([]Statement, 0, 1)
		ret.ElseBlock = append(ret.IfBlock, stmt)

		for {
			stmt2, err := p.This.ParseStatement()
			if err != nil {
				break
			}
			ret.ElseBlock = append(ret.IfBlock, stmt2)
		}
	}

	if p.Current().Type != TypeKeyword || p.Current().Value != "end" {
		return nil, p.NewError("Expected 'end' after if statement", true, ret.Start(), ret.End())
	}
	p.Advance()

	return &ret, nil
}

// ParseExpression parses an expression
func (p *Parser) ParseExpression() (Expression, *Error) {
	p.Log()
	return p.This.ParseLogicExpression()
}

// ParseLogicExpression parses a logical expression
func (p *Parser) ParseLogicExpression() (Expression, *Error) {
	p.Log()
	var exp Expression
	var err *Error

	exp, err = p.This.ParseCompareExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"or", "and"}

	for p.Current().Type == TypeKeyword && contains(logOps, p.Current().Value) {
		binexp := &BinaryOperation{
			Operator: p.Current().Value,
			Exp1:     exp,
		}
		p.Advance()
		var err *Error
		binexp.Exp2, err = p.This.ParseCompareExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp, nil
}

// ParseCompareExpression parses a compare expression
func (p *Parser) ParseCompareExpression() (Expression, *Error) {
	p.Log()
	exp1, err := p.This.ParseSumExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"==", "!=", "<=", ">=", "<", ">"}

	if p.Current().Type == TypeSymbol && contains(logOps, p.Current().Value) {
		binexp := &BinaryOperation{
			Operator: p.Current().Value,
			Exp1:     exp1,
		}
		p.Advance()
		var err *Error
		binexp.Exp2, err = p.This.ParseSumExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		return binexp, nil
	}
	return exp1, nil
}

// ParseSumExpression parses a sum-expression
func (p *Parser) ParseSumExpression() (Expression, *Error) {
	p.Log()
	var exp Expression
	var err *Error

	exp, err = p.This.ParseProdExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"+", "-"}

	for p.Current().Type == TypeSymbol && contains(logOps, p.Current().Value) {
		binexp := &BinaryOperation{
			Operator: p.Current().Value,
			Exp1:     exp,
		}
		p.Advance()
		var err *Error
		binexp.Exp2, err = p.This.ParseProdExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp, nil
}

// ParseProdExpression parses a product expression
func (p *Parser) ParseProdExpression() (Expression, *Error) {
	p.Log()
	var exp Expression
	var err *Error

	exp, err = p.This.ParseUnaryExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"*", "/", "%", "^"}

	for p.Current().Type == TypeSymbol && contains(logOps, p.Current().Value) {
		binexp := &BinaryOperation{
			Operator: p.Current().Value,
			Exp1:     exp,
		}
		p.Advance()
		var err *Error
		binexp.Exp2, err = p.This.ParseUnaryExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp, nil
}

// ParseUnaryExpression parses an unary expression
func (p *Parser) ParseUnaryExpression() (Expression, *Error) {
	p.Log()
	preUnaryOps := []string{"not", "-"}
	if contains(preUnaryOps, p.Current().Value) {
		unaryExp := &UnaryOperation{
			Operator: p.Current().Value,
			Position: p.Current().Position,
		}
		p.Advance()
		subexp, err := p.This.ParseUnaryExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", unaryExp.Operator))
		}
		unaryExp.Exp = subexp
		return unaryExp, nil
	}
	return p.This.ParseBracketExpression()
}

// ParseBracketExpression parses a racketed expression
func (p *Parser) ParseBracketExpression() (Expression, *Error) {
	p.Log()
	if p.Current().Type == TypeSymbol && p.Current().Value == "(" {
		p.Advance()
		innerExp, err := p.This.ParseExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression after '('"))
		}
		if p.Current().Type == TypeSymbol && p.Current().Value == ")" {
			p.Advance()
			return innerExp, nil
		}
		return nil, p.NewError("Missing ')'", true, innerExp.End(), innerExp.End())
	}
	return p.This.ParseSingleExpression()
}

// ParseSingleExpression parses a single expression
func (p *Parser) ParseSingleExpression() (Expression, *Error) {
	p.Log()

	preOpVarDeref, err := p.This.ParsePreOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return preOpVarDeref, nil
	}

	postOpVarDeref, err := p.This.ParsePostOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return postOpVarDeref, nil
	}

	funccall, err := p.This.ParseFuncCall()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return funccall, nil
	}

	if p.Current().Type == TypeID {
		defer p.Advance()
		return &Dereference{
			Variable: p.Current().Value,
			Position: p.Current().Position,
		}, nil
	}
	if p.Current().Type == TypeString {
		defer p.Advance()
		return &StringConstant{
			Value:    p.Current().Value,
			Position: p.Current().Position,
		}, nil
	}
	if p.Current().Type == TypeNumber {
		defer p.Advance()
		return &NumberConstant{
			Value:    p.Current().Value,
			Position: p.Current().Position,
		}, nil
	}
	return nil, p.NewError("Expected constant, variable, func-call or pre/post operation", true, p.Current().Position, p.Current().Position)
}

// ParseFuncCall parse a function call
func (p *Parser) ParseFuncCall() (Expression, *Error) {
	p.Log()
	if p.Current().Type != TypeID || p.Next().Type != TypeSymbol || p.Next().Value != "(" {
		return nil, p.NewError("No function call detected", false, p.Current().Position, p.Current().Position)
	}
	fc := &FuncCall{
		Function: p.Current().Value,
	}
	p.Advance()
	p.Advance()
	arg, err := p.This.ParseExpression()
	fc.Argument = arg
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("Expected expression as function argument"))
	}

	if p.Current().Type != TypeSymbol || p.Current().Value != ")" {
		return nil, p.NewError("Missing ')' on function call", true, fc.Start(), fc.End())
	}
	p.Advance()

	return fc, nil
}

// ParsePreOpExpression parse pre-expression
func (p *Parser) ParsePreOpExpression() (Expression, *Error) {
	p.Log()
	if p.Current().Type == TypeSymbol && (p.Current().Value == "++" || p.Current().Value == "--") {
		exp := Dereference{
			Operator: p.Current().Value,
			PrePost:  "Pre",
			Position: p.Current().Position,
		}
		p.Advance()
		if p.Current().Type != TypeID {
			return nil, p.NewError("Pre- Increment/Decrement must be followed by a variable", true, exp.Start(), exp.Start())
		}
		exp.Variable = p.Current().Value
		p.Advance()
		return &exp, nil
	}
	return nil, p.NewError("No Pre-Operator found", false, p.Current().Position, p.Current().Position)
}

// ParsePostOpExpression parse post-expression
func (p *Parser) ParsePostOpExpression() (Expression, *Error) {
	p.Log()
	if p.Next().Type == TypeSymbol && (p.Next().Value == "++" || p.Next().Value == "--") && p.Current().Type == TypeID {
		exp := Dereference{
			Variable: p.Current().Value,
			Operator: p.Next().Value,
			PrePost:  "Post",
			Position: p.Current().Position,
		}
		p.Advance()
		p.Advance()
		return &exp, nil
	}
	return nil, p.NewError("No Post-Operator found", false, p.Current().Position, p.Current().Position)
}

// NewError creates a new parser error
func (p *Parser) NewError(msg string, terminal bool, start Position, end Position) *Error {
	err := &Error{
		Message:       msg + ". Found Token: '" + p.Current().Value + "'(" + p.Current().Type + ")",
		Fatal:         terminal,
		StartPosition: start,
		EndPosition:   end,
	}
	if p.DebugLog {
		fmt.Println("Created error for", callingFunctionName(), ":", err, "Fatal:", err.Fatal)
	}
	return err
}

func contains(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// get the name of the funtion that called the function that called this function
func callingFunctionName() string {
	fpcs := make([]uintptr, 1)
	// Skip 3 levels to get the caller
	n := runtime.Callers(3, fpcs)
	if n == 0 {
		return ""
	}
	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller == nil {
		return ""
	}
	return strings.Replace(caller.Name(), "github.com/dbaumgarten/yodk/parser.(*Parser).", "", -1)
}

func callingLineNumber() int {
	_, _, line, ok := runtime.Caller(3)
	if ok {
		return line
	}
	return 0
}

// Log logs the visiting of a parsing function
func (p *Parser) Log() {
	if p.DebugLog {
		// Print the name of the function
		fmt.Println("Called:", callingFunctionName(), "from line", callingLineNumber(), "with", p.Current().String())
	}
}
