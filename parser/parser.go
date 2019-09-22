package parser

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

type Parser struct {
	DebugLog     bool
	tokenizer    *Tokenizer
	tokens       []*Token
	currentToken int
	// using an interface of ourself to call the parsing-methods allows them to be overridden by 'subclasses'
	this yololParser
}

type yololParser interface {
	parseStatement() (Statement, *ParserError)
	parsePreOrPostOperation() (Statement, *ParserError)
	parseGoto() (Statement, *ParserError)
	parseAssignment() (Statement, *ParserError)
	parseIf() (Statement, *ParserError)
	parseExpression() (Expression, *ParserError)
	parseLogicExpression() (Expression, *ParserError)
	parseCompareExpression() (Expression, *ParserError)
	parseSumExpression() (Expression, *ParserError)
	parseProdExpression() (Expression, *ParserError)
	parseUnaryExpression() (Expression, *ParserError)
	parseBracketExpression() (Expression, *ParserError)
	parseSingleExpression() (Expression, *ParserError)
	parseFuncCall() (Expression, *ParserError)
	parsePreOpExpression() (Expression, *ParserError)
	parsePostOpExpression() (Expression, *ParserError)
}

func NewParser() *Parser {
	p := &Parser{
		tokenizer: NewTokenizer(),
	}
	p.this = p
	return p
}

func (p *Parser) hasNext() bool {
	return p.currentToken < len(p.tokens)-1
}

func (p *Parser) advance() *Token {
	if p.currentToken < len(p.tokens) {
		p.currentToken++
	}
	// skip whitespace
	for p.currentToken < len(p.tokens) && p.tokens[p.currentToken].Type == TypeWhitespace {
		p.currentToken++
	}
	return p.tokens[p.currentToken]
}

func (p *Parser) prevWithWhitespace() *Token {
	if p.currentToken-1 > 0 {
		return p.tokens[p.currentToken-1]
	}
	return p.tokens[p.currentToken]
}

func (p *Parser) current() *Token {
	offset := 0
	// skip whitespace
	for p.currentToken+offset < len(p.tokens) && p.tokens[p.currentToken+offset].Type == TypeWhitespace {
		offset++
	}
	return p.tokens[p.currentToken+offset]
}

func (p *Parser) next() *Token {
	offset := 1
	// skip whitespace
	for p.currentToken+offset < len(p.tokens) && p.tokens[p.currentToken+offset].Type == TypeWhitespace {
		offset++
	}
	if p.currentToken+offset < len(p.tokens) {
		return p.tokens[p.currentToken+offset]
	}
	return p.tokens[p.currentToken]
}

func (p *Parser) Parse(prog string) (*Programm, error) {
	errors := make(ParserErrors, 0)
	p.tokenizer.Load(prog)
	p.tokens = make([]*Token, 0, 1000)
	for {
		token, err := p.tokenizer.Next()
		if err != nil {
			errors = append(errors, err.(*ParserError))
		} else {
			if p.DebugLog {
				fmt.Print(token)
			}
			p.tokens = append(p.tokens, token)
			if token.Type == TypeEOF {
				break
			}
		}
	}
	p.currentToken = 0
	parsed, err := p.parseProgram()
	errors = append(errors, err...)
	if len(errors) > 0 {
		return nil, errors
	}
	return parsed, nil
}

func (p *Parser) parseProgram() (*Programm, ParserErrors) {
	p.log()
	errors := make(ParserErrors, 0)
	ret := Programm{
		Lines: make([]*Line, 0),
	}
	for p.hasNext() {
		line, err := p.parseLine()
		if err != nil {
			errors = append(errors, err)
			p.skipLine()
		}
		ret.Lines = append(ret.Lines, line)
	}
	return &ret, errors
}

func (p *Parser) skipLine() {
	for p.current().Type != TypeNewline && p.current().Type != TypeEOF {
		p.advance()
	}
}

func (p *Parser) parseLine() (*Line, *ParserError) {
	p.log()
	ret := Line{
		Statements: make([]Statement, 0),
	}

	for p.hasNext() {
		if p.current().Type == TypeNewline || p.current().Type == TypeEOF {
			p.advance()
			return &ret, nil
		}
		stmt, err := p.this.parseStatement()
		if err != nil {
			return nil, err
		}
		ret.Statements = append(ret.Statements, stmt)
	}

	if p.current().Type == TypeEOF {
		return &ret, nil
	}

	return nil, p.newError("Missing newline", true, ret.Start(), ret.End())
}

func (p *Parser) parseStatement() (Statement, *ParserError) {
	p.log()

	// the only place where whitespace can not be ignored
	if p.current().Position.Coloumn != 1 && p.prevWithWhitespace().Type != TypeWhitespace {
		return nil, p.newError("Statements on same line must be seperated by space", true, p.current().Position, p.current().Position)
	}

	stmt, err := p.this.parseAssignment()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return stmt, nil
	}

	stmt2, err2 := p.this.parseIf()
	if err2 != nil && err2.Fatal {
		return nil, err2
	}
	if err2 == nil {
		return stmt2, nil
	}

	stmt3, err3 := p.this.parseGoto()
	if err3 != nil && err3.Fatal {
		return nil, err3
	}
	if err3 == nil {
		return stmt3, nil
	}

	stmt4, err4 := p.this.parsePreOrPostOperation()
	if err4 != nil && err4.Fatal {
		return nil, err4
	}
	if err4 == nil {
		return stmt4, nil
	}

	return nil, p.newError("Expected assignment, if-statement or goto", false, p.current().Position, p.current().Position)
}

func (p *Parser) parsePreOrPostOperation() (Statement, *ParserError) {
	p.log()
	preOpVarDeref, err := p.this.parsePreOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		preOpVarDeref.(*Dereference).IsStatement = true
		return preOpVarDeref, nil
	}

	postOpVarDeref, err := p.this.parsePostOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		postOpVarDeref.(*Dereference).IsStatement = true
		return postOpVarDeref, nil
	}
	return nil, p.newError("No Pre- or Post expression-statement found", false, p.current().Position, p.current().Position)
}

func (p *Parser) parseGoto() (Statement, *ParserError) {
	if p.current().Type == TypeKeyword && p.current().Value == "goto" {
		stmt := GoToStatement{
			Position: p.current().Position,
		}
		p.advance()
		if p.current().Type != TypeNumber {
			return nil, p.newError("Goto must be followed by a line number", true, stmt.Start(), stmt.Start())
		}
		line, err := strconv.Atoi(p.current().Value)
		stmt.Line = line
		if err != nil {
			return nil, p.newError("Goto must be followed by a line number", true, stmt.Start(), stmt.End())
		}
		p.advance()
		return &stmt, nil
	}
	return nil, p.newError("Goto statements must start with 'goto'", false, p.current().Position, p.current().Position)
}

func (p *Parser) parseAssignment() (Statement, *ParserError) {
	p.log()
	assignmentOperators := []string{"=", "+=", "-=", "*=", "/=", "%="}
	ret := Assignment{
		Position: p.current().Position,
	}
	if p.current().Type != TypeID || !contains(assignmentOperators, p.next().Value) {
		return nil, p.newError("Expected identifier and assignment operator", false, p.current().Position, p.next().Position)
	}
	ret.Variable = p.current().Value
	p.advance()
	ret.Operator = p.current().Value
	p.advance()
	exp, err := p.this.parseExpression()
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("Expected expression on right side of assignment"))
	}
	ret.Value = exp
	return &ret, nil
}

func (p *Parser) parseIf() (Statement, *ParserError) {
	p.log()
	ret := IfStatement{
		Position: p.current().Position,
	}
	if p.current().Type != TypeKeyword || p.current().Value != "if" {
		return nil, p.newError("If-statements have to start with 'if'", false, p.current().Position, p.current().Position)
	}
	p.advance()

	var err *ParserError
	ret.Condition, err = p.this.parseExpression()
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("No expression found as if-condition"))
	}

	if p.current().Type != TypeKeyword || p.current().Value != "then" {
		return nil, p.newError("Expected 'then' after condition", true, p.current().Position, p.current().Position)
	}
	p.advance()

	stmt, err := p.this.parseStatement()
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("Expected statement after 'then'"))
	}
	ret.IfBlock = make([]Statement, 0, 1)
	ret.IfBlock = append(ret.IfBlock, stmt)

	for {
		stmt2, err := p.this.parseStatement()
		if err != nil {
			break
		}
		ret.IfBlock = append(ret.IfBlock, stmt2)
	}

	if p.current().Type == TypeKeyword && p.current().Value == "else" {
		p.advance()
		stmt, err := p.this.parseStatement()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected statement after 'else'"))
		}
		ret.ElseBlock = make([]Statement, 0, 1)
		ret.ElseBlock = append(ret.IfBlock, stmt)

		for {
			stmt2, err := p.this.parseStatement()
			if err != nil {
				break
			}
			ret.ElseBlock = append(ret.IfBlock, stmt2)
		}
	}

	if p.current().Type != TypeKeyword || p.current().Value != "end" {
		return nil, p.newError("Expected 'end' after if statement", true, ret.Start(), ret.End())
	}
	p.advance()

	return &ret, nil
}

func (p *Parser) parseExpression() (Expression, *ParserError) {
	p.log()
	return p.this.parseLogicExpression()
}

func (p *Parser) parseLogicExpression() (Expression, *ParserError) {
	p.log()
	var exp Expression
	var err *ParserError

	exp, err = p.this.parseCompareExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"or", "and"}

	for p.current().Type == TypeKeyword && contains(logOps, p.current().Value) {
		binexp := &BinaryOperation{
			Operator: p.current().Value,
			Exp1:     exp,
		}
		p.advance()
		var err *ParserError
		binexp.Exp2, err = p.this.parseCompareExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp, nil
}

func (p *Parser) parseCompareExpression() (Expression, *ParserError) {
	p.log()
	exp1, err := p.this.parseSumExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"==", "!=", "<=", ">=", "<", ">"}

	if p.current().Type == TypeSymbol && contains(logOps, p.current().Value) {
		binexp := &BinaryOperation{
			Operator: p.current().Value,
			Exp1:     exp1,
		}
		p.advance()
		var err *ParserError
		binexp.Exp2, err = p.this.parseSumExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		return binexp, nil
	}
	return exp1, nil
}

func (p *Parser) parseSumExpression() (Expression, *ParserError) {
	p.log()
	var exp Expression
	var err *ParserError

	exp, err = p.this.parseProdExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"+", "-"}

	for p.current().Type == TypeSymbol && contains(logOps, p.current().Value) {
		binexp := &BinaryOperation{
			Operator: p.current().Value,
			Exp1:     exp,
		}
		p.advance()
		var err *ParserError
		binexp.Exp2, err = p.this.parseProdExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp, nil
}

func (p *Parser) parseProdExpression() (Expression, *ParserError) {
	p.log()
	var exp Expression
	var err *ParserError

	exp, err = p.this.parseUnaryExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"*", "/", "%", "^"}

	for p.current().Type == TypeSymbol && contains(logOps, p.current().Value) {
		binexp := &BinaryOperation{
			Operator: p.current().Value,
			Exp1:     exp,
		}
		p.advance()
		var err *ParserError
		binexp.Exp2, err = p.this.parseUnaryExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp, nil
}

func (p *Parser) parseUnaryExpression() (Expression, *ParserError) {
	p.log()
	preUnaryOps := []string{"not", "-"}
	if contains(preUnaryOps, p.current().Value) {
		unaryExp := &UnaryOperation{
			Operator: p.current().Value,
			Position: p.current().Position,
		}
		p.advance()
		subexp, err := p.this.parseUnaryExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression on right side of %s", unaryExp.Operator))
		}
		unaryExp.Exp = subexp
		return unaryExp, nil
	}
	return p.this.parseBracketExpression()
}

func (p *Parser) parseBracketExpression() (Expression, *ParserError) {
	p.log()
	if p.current().Type == TypeSymbol && p.current().Value == "(" {
		p.advance()
		innerExp, err := p.this.parseExpression()
		if err != nil {
			err.Fatal = true
			return nil, err.Append(fmt.Errorf("Expected expression after '('"))
		}
		if p.current().Type == TypeSymbol && p.current().Value == ")" {
			p.advance()
			return innerExp, nil
		}
		return nil, p.newError("Missing ')'", true, innerExp.End(), innerExp.End())
	}
	return p.this.parseSingleExpression()
}

func (p *Parser) parseSingleExpression() (Expression, *ParserError) {
	p.log()

	preOpVarDeref, err := p.this.parsePreOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return preOpVarDeref, nil
	}

	postOpVarDeref, err := p.this.parsePostOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return postOpVarDeref, nil
	}

	funccall, err := p.this.parseFuncCall()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return funccall, nil
	}

	if p.current().Type == TypeID {
		defer p.advance()
		return &Dereference{
			Variable: p.current().Value,
			Position: p.current().Position,
		}, nil
	}
	if p.current().Type == TypeString {
		defer p.advance()
		return &StringConstant{
			Value:    p.current().Value,
			Position: p.current().Position,
		}, nil
	}
	if p.current().Type == TypeNumber {
		defer p.advance()
		return &NumberConstant{
			Value:    p.current().Value,
			Position: p.current().Position,
		}, nil
	}
	return nil, p.newError("Expected constant, variable, func-call or pre/post operation", true, p.current().Position, p.current().Position)
}

func (p *Parser) parseFuncCall() (Expression, *ParserError) {
	p.log()
	if p.current().Type != TypeID || p.next().Type != TypeSymbol || p.next().Value != "(" {
		return nil, p.newError("No function call detected", false, p.current().Position, p.current().Position)
	}
	fc := &FuncCall{
		Function: p.current().Value,
	}
	p.advance()
	p.advance()
	arg, err := p.this.parseExpression()
	fc.Argument = arg
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("Expected expression as function argument"))
	}

	if p.current().Type != TypeSymbol || p.current().Value != ")" {
		return nil, p.newError("Missing ')' on function call", true, fc.Start(), fc.End())
	}
	p.advance()

	return fc, nil
}

func (p *Parser) parsePreOpExpression() (Expression, *ParserError) {
	p.log()
	if p.current().Type == TypeSymbol && (p.current().Value == "++" || p.current().Value == "--") {
		exp := Dereference{
			Operator: p.current().Value,
			PrePost:  "Pre",
			Position: p.current().Position,
		}
		p.advance()
		if p.current().Type != TypeID {
			return nil, p.newError("Pre- Increment/Decrement must be followed by a variable", true, exp.Start(), exp.Start())
		}
		exp.Variable = p.current().Value
		p.advance()
		return &exp, nil
	}
	return nil, p.newError("No Pre-Operator found", false, p.current().Position, p.current().Position)
}

func (p *Parser) parsePostOpExpression() (Expression, *ParserError) {
	p.log()
	if p.next().Type == TypeSymbol && (p.next().Value == "++" || p.next().Value == "--") && p.current().Type == TypeID {
		exp := Dereference{
			Variable: p.current().Value,
			Operator: p.next().Value,
			PrePost:  "Post",
			Position: p.current().Position,
		}
		p.advance()
		p.advance()
		return &exp, nil
	}
	return nil, p.newError("No Post-Operator found", false, p.current().Position, p.current().Position)
}

func (p *Parser) newError(msg string, terminal bool, start Position, end Position) *ParserError {
	err := &ParserError{
		Message:       msg + ". Found Token: '" + p.current().Value + "'(" + p.current().Type + ")",
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

func (p *Parser) log() {
	if p.DebugLog {
		// Print the name of the function
		fmt.Println("Called:", callingFunctionName(), "from line", callingLineNumber(), "with", p.current().String())
	}
}
