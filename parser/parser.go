package parser

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/ast"
	"github.com/dbaumgarten/yodk/tokenizer"
)

type Parser struct {
	DebugLog     bool
	tokenizer    *tokenizer.Tokenizer
	tokens       []*tokenizer.Token
	currentToken int
}

func NewParser() *Parser {
	return &Parser{
		tokenizer: tokenizer.NewTokenizer(),
	}
}

type ParserError struct {
	Message string
	Line    int
	Fatal   bool
}

func (e ParserError) Error() string {
	return fmt.Sprintf("Parser error at line %d: %s", e.Line, e.Message)
}

func (p *Parser) hasNext() bool {
	return p.currentToken < len(p.tokens)-1
}

func (p *Parser) next() *tokenizer.Token {
	if p.currentToken < len(p.tokens) {
		p.currentToken++
	}
	return p.tokens[p.currentToken]
}

func (p *Parser) current() *tokenizer.Token {
	return p.tokens[p.currentToken]
}

func (p *Parser) peek() *tokenizer.Token {
	if p.currentToken < len(p.tokens) {
		return p.tokens[p.currentToken+1]
	}
	return p.tokens[p.currentToken]
}

func (p *Parser) Parse(prog string) (*ast.Programm, error) {
	p.tokenizer.Load(prog)
	p.tokens = make([]*tokenizer.Token, 0, 1000)
	for {
		token, err := p.tokenizer.Next()
		if err != nil {
			return nil, &ParserError{
				Message: err.Error(),
				Fatal:   true,
			}
		}
		if p.DebugLog {
			fmt.Print(token)
		}
		p.tokens = append(p.tokens, token)
		if token.Type == tokenizer.TypeEOF {
			break
		}
	}
	p.currentToken = 0
	parsed, err := p.parseProgram()
	if err == nil {
		return parsed, nil
	}
	return nil, err
}

func (p *Parser) parseProgram() (*ast.Programm, *ParserError) {
	p.log()
	ret := ast.Programm{
		Lines: make([]*ast.Line, 0),
	}
	for p.hasNext() {
		line, err := p.parseLine()
		if err != nil {
			return nil, err
		}
		ret.Lines = append(ret.Lines, line)
	}
	return &ret, nil
}

func (p *Parser) parseLine() (*ast.Line, *ParserError) {
	p.log()
	ret := ast.Line{
		Statements: make([]ast.Statement, 0),
	}

	for p.hasNext() {
		if p.current().Type == tokenizer.TypeNewline || p.current().Type == tokenizer.TypeEOF {
			p.next()
			return &ret, nil
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		ret.Statements = append(ret.Statements, stmt)
	}

	if p.current().Type == tokenizer.TypeEOF {
		return &ret, nil
	}

	return nil, p.getError("Missing newline", true)
}

func (p *Parser) parseStatement() (ast.Statement, *ParserError) {
	p.log()
	stmt, err := p.parseAssignment()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return stmt, nil
	}

	stmt2, err2 := p.parseIf()
	if err2 != nil && err2.Fatal {
		return nil, err2
	}
	if err2 == nil {
		return stmt2, nil
	}

	stmt3, err3 := p.parseGoto()
	if err3 != nil && err3.Fatal {
		return nil, err3
	}
	if err3 == nil {
		return stmt3, nil
	}

	stmt4, err4 := p.parsePreOrPostOperation()
	if err4 != nil && err4.Fatal {
		return nil, err4
	}
	if err4 == nil {
		return stmt4, nil
	}

	return nil, p.getErrorE("Expected assignment, if-statement or goto", false, err2)
}

func (p *Parser) parsePreOrPostOperation() (*ast.Dereference, *ParserError) {
	p.log()
	preOpVarDeref, err := p.parsePreOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return preOpVarDeref, nil
	}

	postOpVarDeref, err := p.parsePostOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return postOpVarDeref, nil
	}
	return nil, p.getError("No Pre- or Post expression-statement found", false)
}

func (p *Parser) parseGoto() (*ast.GoToStatement, *ParserError) {
	if p.current().Type == tokenizer.TypeKeyword && p.current().Value == "goto" {
		p.next()
		if p.current().Type != tokenizer.TypeNumber {
			return nil, p.getError("Goto must be followed by a line number", true)
		}
		line, err := strconv.Atoi(p.current().Value)
		if err != nil {
			return nil, p.getError("Goto must be followed by a line number", true)
		}
		p.next()
		stmt := ast.GoToStatement{
			Line: line,
		}
		return &stmt, nil
	}
	return nil, p.getError("Goto statements must start with 'goto'", false)
}

func (p *Parser) parseAssignment() (*ast.Assignment, *ParserError) {
	p.log()
	assignmentOperators := []string{"=", "+=", "-=", "*=", "/=", "%="}
	ret := ast.Assignment{}
	if p.current().Type != tokenizer.TypeID || !contains(assignmentOperators, p.peek().Value) {
		return nil, p.getError("Expected identifier and assignment operator", false)
	}
	ret.Variable = p.current().Value
	p.next()
	ret.Operator = p.current().Value
	p.next()
	exp, err := p.parseExpression()
	if err != nil {
		return nil, p.getErrorE("Expected expression on right side of assignment", true, err)
	}
	ret.Value = exp
	return &ret, nil
}

func (p *Parser) parseIf() (*ast.IfStatement, *ParserError) {
	p.log()
	ret := ast.IfStatement{}
	if p.current().Type != tokenizer.TypeKeyword || p.current().Value != "if" {
		return nil, p.getError("If-statements have to start with 'if'", false)
	}
	p.next()

	var err *ParserError
	ret.Condition, err = p.parseExpression()
	if err != nil {
		return nil, p.getErrorE("No expression found as if-condition", true, err)
	}

	if p.current().Type != tokenizer.TypeKeyword || p.current().Value != "then" {
		return nil, p.getError("Expected 'then' after condition", true)
	}
	p.next()

	stmt, err := p.parseStatement()
	if err != nil {
		return nil, p.getErrorE("Expected statement after 'then'", true, err)
	}
	ret.IfBlock = make([]ast.Statement, 0, 1)
	ret.IfBlock = append(ret.IfBlock, stmt)

	for {
		stmt2, err := p.parseStatement()
		if err != nil {
			break
		}
		ret.IfBlock = append(ret.IfBlock, stmt2)
	}

	if p.current().Type == tokenizer.TypeKeyword && p.current().Value == "else" {
		p.next()
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, p.getErrorE("Expected statement after 'else'", true, err)
		}
		ret.ElseBlock = make([]ast.Statement, 0, 1)
		ret.ElseBlock = append(ret.IfBlock, stmt)

		for {
			stmt2, err := p.parseStatement()
			if err != nil {
				break
			}
			ret.ElseBlock = append(ret.IfBlock, stmt2)
		}
	}

	if p.current().Type != tokenizer.TypeKeyword || p.current().Value != "end" {
		return nil, p.getError("Expected 'end' after if statement", true)
	}
	p.next()

	return &ret, nil
}

func (p *Parser) parseExpression() (ast.Expression, *ParserError) {
	p.log()
	return p.parseLogicExpression()
}

func (p *Parser) parseLogicExpression() (ast.Expression, *ParserError) {
	p.log()
	var exp ast.Expression
	var err *ParserError

	exp, err = p.parseCompareExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"or", "and"}

	for p.current().Type == tokenizer.TypeKeyword && contains(logOps, p.current().Value) {
		binexp := &ast.BinaryOperation{
			Operator: p.current().Value,
			Exp1:     exp,
		}
		p.next()
		var err *ParserError
		binexp.Exp2, err = p.parseCompareExpression()
		if err != nil {
			return nil, p.getErrorE(fmt.Sprintf("Expected expression on right side of %s", binexp.Operator), true, err)
		}
		exp = binexp
	}
	return exp, nil
}

func (p *Parser) parseCompareExpression() (ast.Expression, *ParserError) {
	p.log()
	exp1, err := p.parseSumExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"==", "!=", "<=", ">=", "<", ">"}

	if p.current().Type == tokenizer.TypeSymbol && contains(logOps, p.current().Value) {
		binexp := &ast.BinaryOperation{
			Operator: p.current().Value,
			Exp1:     exp1,
		}
		p.next()
		var err *ParserError
		binexp.Exp2, err = p.parseSumExpression()
		if err != nil {
			return nil, p.getErrorE(fmt.Sprintf("Expected expression on right side of %s", binexp.Operator), true, err)
		}
		return binexp, nil
	}
	return exp1, nil
}

func (p *Parser) parseSumExpression() (ast.Expression, *ParserError) {
	p.log()
	var exp ast.Expression
	var err *ParserError

	exp, err = p.parseProdExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"+", "-"}

	for p.current().Type == tokenizer.TypeSymbol && contains(logOps, p.current().Value) {
		binexp := &ast.BinaryOperation{
			Operator: p.current().Value,
			Exp1:     exp,
		}
		p.next()
		var err *ParserError
		binexp.Exp2, err = p.parseProdExpression()
		if err != nil {
			return nil, p.getErrorE(fmt.Sprintf("Expected expression on right side of %s", binexp.Operator), true, err)
		}
		exp = binexp
	}
	return exp, nil
}

func (p *Parser) parseProdExpression() (ast.Expression, *ParserError) {
	p.log()
	var exp ast.Expression
	var err *ParserError

	exp, err = p.parseUnaryExpression()
	if err != nil {
		return nil, err
	}
	logOps := []string{"*", "/", "%", "^"}

	for p.current().Type == tokenizer.TypeSymbol && contains(logOps, p.current().Value) {
		binexp := &ast.BinaryOperation{
			Operator: p.current().Value,
			Exp1:     exp,
		}
		p.next()
		var err *ParserError
		binexp.Exp2, err = p.parseUnaryExpression()
		if err != nil {
			return nil, p.getErrorE(fmt.Sprintf("Expected expression on right side of %s", binexp.Operator), true, err)
		}
		exp = binexp
	}
	return exp, nil
}

func (p *Parser) parseUnaryExpression() (ast.Expression, *ParserError) {
	p.log()
	preUnaryOps := []string{"not", "-"}
	if contains(preUnaryOps, p.current().Value) {
		unaryExp := &ast.UnaryOperation{
			Operator: p.current().Value,
		}
		p.next()
		subexp, err := p.parseUnaryExpression()
		if err != nil {
			return nil, p.getErrorE(fmt.Sprintf("Expected expression on right side of %s", unaryExp.Operator), true, err)
		}
		unaryExp.Exp = subexp
		return unaryExp, nil
	}
	return p.parseBracketExpression()
}

func (p *Parser) parseBracketExpression() (ast.Expression, *ParserError) {
	p.log()
	if p.current().Type == tokenizer.TypeSymbol && p.current().Value == "(" {
		p.next()
		innerExp, err := p.parseExpression()
		if err != nil {
			return nil, p.getErrorE("Expected expression after '('", true, err)
		}
		if p.current().Type == tokenizer.TypeSymbol && p.current().Value == ")" {
			p.next()
			return innerExp, nil
		}
		return nil, p.getError("Missing ')'", true)
	}
	return p.parseSingleExpression()
}

func (p *Parser) parseSingleExpression() (ast.Expression, *ParserError) {
	p.log()

	preOpVarDeref, err := p.parsePreOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return preOpVarDeref, nil
	}

	postOpVarDeref, err := p.parsePostOpExpression()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return postOpVarDeref, nil
	}

	funccall, err := p.parseFuncCall()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return funccall, nil
	}

	if p.current().Type == tokenizer.TypeID {
		value := p.current().Value
		p.next()
		return &ast.Dereference{
			Variable: value,
		}, nil
	}
	if p.current().Type == tokenizer.TypeString {
		value := p.current().Value
		p.next()
		return &ast.StringConstant{
			Value: value,
		}, nil
	}
	if p.current().Type == tokenizer.TypeNumber {
		value := p.current().Value
		p.next()
		return &ast.NumberConstant{
			Value: value,
		}, nil
	}
	return nil, p.getError(fmt.Sprintf("Expected expression, but found token '%s'", p.current()), true)
}

func (p *Parser) parseFuncCall() (*ast.FuncCall, *ParserError) {
	p.log()
	if p.current().Type != tokenizer.TypeID || p.peek().Type != tokenizer.TypeSymbol || p.peek().Value != "(" {
		return nil, p.getError("No function call detected", false)
	}
	fc := &ast.FuncCall{
		Function: p.current().Value,
	}
	p.next()
	p.next()
	arg, err := p.parseExpression()
	if err != nil {
		return nil, p.getErrorE("Expected expression as function argument", true, err)
	}

	if p.current().Type != tokenizer.TypeSymbol || p.current().Value != ")" {
		return nil, p.getError("Missing '(' on function call", true)
	}
	p.next()

	fc.Argument = arg
	return fc, nil
}

func (p *Parser) parsePreOpExpression() (*ast.Dereference, *ParserError) {
	p.log()
	if p.current().Type == tokenizer.TypeSymbol && (p.current().Value == "++" || p.current().Value == "--") {
		exp := ast.Dereference{
			Operator: p.current().Value,
			PrePost:  "Pre",
		}
		p.next()
		if p.current().Type != tokenizer.TypeID {
			return nil, p.getError("Pre- Increment/Decrement must be followed by variable", true)
		}
		exp.Variable = p.current().Value
		p.next()
		return &exp, nil
	}
	return nil, p.getError("No Pre-Operator found", false)
}

func (p *Parser) parsePostOpExpression() (*ast.Dereference, *ParserError) {
	p.log()
	if p.peek().Type == tokenizer.TypeSymbol && (p.peek().Value == "++" || p.peek().Value == "--") && p.current().Type == tokenizer.TypeID {
		exp := ast.Dereference{
			Variable: p.current().Value,
			Operator: p.peek().Value,
			PrePost:  "Post",
		}
		p.next()
		p.next()
		return &exp, nil
	}
	return nil, p.getError("No Post-Operator found", false)
}

func (p *Parser) getError(msg string, terminal bool) *ParserError {
	err := &ParserError{
		Message: msg + ". Found Token: '" + p.current().Value + "'(" + p.current().Type + ")",
		Line:    p.current().Line,
		Fatal:   terminal,
	}
	if p.DebugLog {
		fmt.Println("Created error for", callingFunctionName(), ":", err, "Fatal:", err.Fatal, "\n")
	}
	return err
}

func (p *Parser) getErrorE(msg string, terminal bool, prev *ParserError) *ParserError {
	err := &ParserError{
		Message: msg + ". Previous Error: " + prev.Message,
		Line:    p.current().Line,
		Fatal:   terminal,
	}
	if p.DebugLog {
		fmt.Println("Created error for", callingFunctionName(), ":", err, "Fatal:", err.Fatal, "\n")
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
