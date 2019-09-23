package parser

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

type Parser struct {
	DebugLog     bool
	Tokenizer    *Tokenizer
	Tokens       []*Token
	CurrentToken int
	// using an interface of ourself to call the parsing-methods allows them to be overridden by 'subclasses'
	This YololParserFunctions
}

type YololParserFunctions interface {
	ParseStatement() (Statement, *ParserError)
	ParsePreOrPostOperation() (Statement, *ParserError)
	ParseGoto() (Statement, *ParserError)
	ParseAssignment() (Statement, *ParserError)
	ParseIf() (Statement, *ParserError)
	ParseExpression() (Expression, *ParserError)
	ParseLogicExpression() (Expression, *ParserError)
	ParseCompareExpression() (Expression, *ParserError)
	ParseSumExpression() (Expression, *ParserError)
	ParseProdExpression() (Expression, *ParserError)
	ParseUnaryExpression() (Expression, *ParserError)
	ParseBracketExpression() (Expression, *ParserError)
	ParseSingleExpression() (Expression, *ParserError)
	ParseFuncCall() (Expression, *ParserError)
	ParsePreOpExpression() (Expression, *ParserError)
	ParsePostOpExpression() (Expression, *ParserError)
}

func NewParser() *Parser {
	p := &Parser{
		Tokenizer: NewTokenizer(),
	}
	p.This = p
	return p
}

func (p *Parser) HasNext() bool {
	return p.CurrentToken < len(p.Tokens)-1
}

func (p *Parser) Advance() *Token {
	if p.CurrentToken < len(p.Tokens) {
		p.CurrentToken++
	}
	// skip whitespace
	for p.CurrentToken < len(p.Tokens) && p.Tokens[p.CurrentToken].Type == TypeWhitespace {
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

func (p *Parser) Current() *Token {
	offset := 0
	// skip whitespace
	for p.CurrentToken+offset < len(p.Tokens) && p.Tokens[p.CurrentToken+offset].Type == TypeWhitespace {
		offset++
	}
	return p.Tokens[p.CurrentToken+offset]
}

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

func (p *Parser) Parse(prog string) (*Programm, error) {
	errors := make(ParserErrors, 0)
	p.Tokenizer.Load(prog)
	p.Tokens = make([]*Token, 0, 1000)
	for {
		token, err := p.Tokenizer.Next()
		if err != nil {
			errors = append(errors, err.(*ParserError))
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

func (p *Parser) ParseProgram() (*Programm, ParserErrors) {
	p.Log()
	errors := make(ParserErrors, 0)
	ret := Programm{
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

func (p *Parser) SkipLine() {
	for p.Current().Type != TypeNewline && p.Current().Type != TypeEOF {
		p.Advance()
	}
}

func (p *Parser) ParseLine() (*Line, *ParserError) {
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

func (p *Parser) ParseStatement() (Statement, *ParserError) {
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

func (p *Parser) ParsePreOrPostOperation() (Statement, *ParserError) {
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

func (p *Parser) ParseGoto() (Statement, *ParserError) {
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

func (p *Parser) ParseAssignment() (Statement, *ParserError) {
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

func (p *Parser) ParseIf() (Statement, *ParserError) {
	p.Log()
	ret := IfStatement{
		Position: p.Current().Position,
	}
	if p.Current().Type != TypeKeyword || p.Current().Value != "if" {
		return nil, p.NewError("If-statements have to start with 'if'", false, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	var err *ParserError
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

func (p *Parser) ParseExpression() (Expression, *ParserError) {
	p.Log()
	return p.This.ParseLogicExpression()
}

func (p *Parser) ParseLogicExpression() (Expression, *ParserError) {
	p.Log()
	var exp Expression
	var err *ParserError

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
		var err *ParserError
		binexp.Exp2, err = p.This.ParseCompareExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp, nil
}

func (p *Parser) ParseCompareExpression() (Expression, *ParserError) {
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
		var err *ParserError
		binexp.Exp2, err = p.This.ParseSumExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		return binexp, nil
	}
	return exp1, nil
}

func (p *Parser) ParseSumExpression() (Expression, *ParserError) {
	p.Log()
	var exp Expression
	var err *ParserError

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
		var err *ParserError
		binexp.Exp2, err = p.This.ParseProdExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp, nil
}

func (p *Parser) ParseProdExpression() (Expression, *ParserError) {
	p.Log()
	var exp Expression
	var err *ParserError

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
		var err *ParserError
		binexp.Exp2, err = p.This.ParseUnaryExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp, nil
}

func (p *Parser) ParseUnaryExpression() (Expression, *ParserError) {
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

func (p *Parser) ParseBracketExpression() (Expression, *ParserError) {
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

func (p *Parser) ParseSingleExpression() (Expression, *ParserError) {
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

func (p *Parser) ParseFuncCall() (Expression, *ParserError) {
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

func (p *Parser) ParsePreOpExpression() (Expression, *ParserError) {
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

func (p *Parser) ParsePostOpExpression() (Expression, *ParserError) {
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

func (p *Parser) NewError(msg string, terminal bool, start Position, end Position) *ParserError {
	err := &ParserError{
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

func (p *Parser) Log() {
	if p.DebugLog {
		// Print the name of the function
		fmt.Println("Called:", callingFunctionName(), "from line", callingLineNumber(), "with", p.Current().String())
	}
}
