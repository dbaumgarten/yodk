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
	CurrentToken *Token
	NextToken    *Token
	PrevToken    *Token
	// if true, there was whitespace between CurrentToken and NextToken
	NextWouldBeWhitespace bool
	// if true, current token was preceeded by whitespace
	SkippedWhitespace bool
	// using an interface of ourself to call the parsing-methods allows them to be overridden by 'subclasses'
	This YololParserFunctions
	// Contains all comments encountered during parsing
	Comments []*Token
	// Contains all errors encountered during parsing
	Errors Errors
}

/*
How parsing functions (parseXYZ) MUST behave:

If the parsing function can not parse the current token and is not sure, that it is the only function that could
parse the current token, it must return nil. This way another function can try different functions and watch
which of them returns a non-nil, and does not need to know exactly what function to call beforehand.

A function returning nil MUST NOT have advanced the current token and MUST NOT have logged an error, because it
can not be sure that no other function is called to try to parse the current token.

If a parsing function is sure, that the current token is meant for it (e.g. the parseIf function finds an if token)
it MUST return a non-nil value, to indicate that it is in charge of the current token.

If a parsing function is sure that it is charge (will return non-nil) and encounters an error, it logs the
error using the Error()-Function, but will still NOT return nil. It MUST return any non-nil value and is allowed
to leave vital fields empty in the returned value.
*/

// YololParserFunctions is used together with Parser.This to allow 'subclasses' to override 'virtual functions'
type YololParserFunctions interface {
	ParseStatement() Statement
	ParsePreOrPostOperation() Statement
	ParseGoto() Statement
	ParseAssignment() Statement
	ParseIf() Statement
	ParseExpression() Expression
	ParseLogicExpression() Expression
	ParseCompareExpression() Expression
	ParseSumExpression() Expression
	ParseProdExpression() Expression
	ParseUnaryExpression() Expression
	ParseBracketExpression() Expression
	ParseSingleExpression() Expression
	ParseFuncCall() Expression
	ParsePreOpExpression() Expression
	ParsePostOpExpression() Expression
}

// NewParser creates a new parser
func NewParser() *Parser {
	p := &Parser{}
	p.init()
	p.This = p
	return p
}

// ---------------------------------------------

// HasNext returns true if there is a next token
func (p *Parser) HasNext() bool {
	return p.CurrentToken.Type != TypeEOF
}

// Advance advances the current token to the next (non whitespace) token in the list
func (p *Parser) Advance() *Token {
	if p.CurrentToken == nil || p.HasNext() {
		p.PrevToken = p.CurrentToken
		p.CurrentToken = p.NextToken
		p.NextToken = p.Tokenizer.Next()
		p.SkippedWhitespace = p.NextWouldBeWhitespace
		p.NextWouldBeWhitespace = false
		for p.NextToken.Type != TypeEOF && (p.NextToken.Type == TypeWhitespace || p.NextToken.Type == TypeComment) {
			if p.NextToken.Type == TypeWhitespace {
				p.NextWouldBeWhitespace = true
			} else {
				// next token is a comment. Store it.
				p.Comments = append(p.Comments, p.NextToken)
			}
			p.NextToken = p.Tokenizer.Next()
		}

	}
	return p.CurrentToken
}

// Error appends an error to the list of errors encountered during parsing
func (p *Parser) Error(msg string, start Position, end Position) {
	err := &Error{
		Message:       msg + ". Found Token: '" + p.CurrentToken.Value + "'(" + p.CurrentToken.Type + ")",
		StartPosition: start,
		EndPosition:   end,
	}
	p.Errors = append(p.Errors, err)
}

// ErrorCurrent calls Error() with the position of the current token
func (p *Parser) ErrorCurrent(msg string) {
	p.Error(msg, p.CurrentToken.Position, p.CurrentToken.Position)
}

// Expect checks if the current token has the given type and value
// if true, the tokens position is returned, otherwise an error is logged
// alsways advances to the next token
func (p *Parser) Expect(tokenType string, tokenValue string) Position {
	if p.CurrentToken.Type != tokenType || p.CurrentToken.Value != tokenValue {
		var msg string
		if tokenType == TypeNewline {
			msg = "Expected newline"
		} else {
			msg = fmt.Sprintf("Expected token '%s'(%s)", tokenValue, tokenType)
		}
		p.ErrorCurrent(msg)
	}
	pos := p.CurrentToken.Position
	p.Advance()
	return pos
}

// init prepares all internal fields for a new parsing run
func (p *Parser) init() {
	p.Tokenizer = NewTokenizer()
	p.Errors = make(Errors, 0)
	p.Comments = make([]*Token, 0)
	p.CurrentToken = nil
	p.PrevToken = nil
	p.NextToken = nil
	p.NextWouldBeWhitespace = false
	p.SkippedWhitespace = false
}

// ---------------------------------------------

// Parse is the main method of the parser. Parses a yolol-program into an AST.
func (p *Parser) Parse(prog string) (*Program, error) {
	p.init()
	p.Tokenizer.Load(prog)
	// Advance twice to fill CurrentToken and NextToken
	p.Advance()
	p.Advance()
	parsed := p.ParseProgram()
	parsed.Comments = p.Comments
	if len(p.Errors) == 0 {
		return parsed, nil
	}
	return nil, p.Errors
}

// ParseProgram parses a programm-node
func (p *Parser) ParseProgram() *Program {
	p.Log()
	ret := Program{
		Lines: make([]*Line, 0),
	}
	for p.HasNext() {
		ret.Lines = append(ret.Lines, p.ParseLine())
	}
	return &ret
}

// SkipLine skips tokens up to the next newline
func (p *Parser) SkipLine() {
	for p.CurrentToken.Type != TypeNewline && p.CurrentToken.Type != TypeEOF {
		p.Advance()
	}
}

// ParseLine parses a line-node
func (p *Parser) ParseLine() *Line {
	p.Log()
	ret := Line{
		Position:   p.CurrentToken.Position,
		Statements: make([]Statement, 0),
	}

	// empty line
	if p.CurrentToken.Type == TypeNewline || p.CurrentToken.Type == TypeEOF {
		p.Advance()
		return &ret
	}

	for p.HasNext() {
		stmt := p.This.ParseStatement()
		if stmt == nil {
			p.ErrorCurrent("Expected a statement")
			p.Advance()
		}
		ret.Statements = append(ret.Statements, stmt)

		// line ends after statement
		if p.CurrentToken.Type == TypeNewline || p.CurrentToken.Type == TypeEOF {
			p.Advance()
			return &ret
		}

		// more statements on this line?
		if !p.SkippedWhitespace {
			p.ErrorCurrent("Statements must be followed by space or newline")
		}
	}

	if p.CurrentToken.Type != TypeEOF {
		p.Error("Missing newline", ret.Start(), ret.End())
	}

	return &ret
}

// ParseStatement parses a statement-node
func (p *Parser) ParseStatement() Statement {
	p.Log()

	stmt := p.This.ParseAssignment()
	if stmt != nil {
		return stmt
	}

	stmt2 := p.This.ParseIf()
	if stmt2 != nil {
		return stmt2
	}

	stmt3 := p.This.ParseGoto()
	if stmt3 != nil {
		return stmt3
	}

	stmt4 := p.This.ParsePreOrPostOperation()
	if stmt4 != nil {
		return stmt4
	}

	return nil
}

// ParsePreOrPostOperation parses a pre-/post operation (x++, ++x) as a statement
func (p *Parser) ParsePreOrPostOperation() Statement {
	p.Log()
	preOpVarDeref := p.This.ParsePreOpExpression()
	if preOpVarDeref != nil {
		preOpVarDeref.(*Dereference).IsStatement = true
		return preOpVarDeref
	}

	postOpVarDeref := p.This.ParsePostOpExpression()
	if postOpVarDeref != nil {
		postOpVarDeref.(*Dereference).IsStatement = true
		return postOpVarDeref
	}

	return nil
}

// ParseGoto parse parses a goto-node
func (p *Parser) ParseGoto() Statement {
	if p.CurrentToken.Type == TypeKeyword && p.CurrentToken.Value == "goto" {
		stmt := GoToStatement{
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		line, err := strconv.Atoi(p.CurrentToken.Value)
		if p.CurrentToken.Type != TypeNumber || err != nil {
			p.Error("Goto must be followed by a line number", stmt.Start(), stmt.Start())
		}
		stmt.Line = line
		p.Advance()
		return &stmt
	}
	return nil
}

// ParseAssignment parses an assignment-node
func (p *Parser) ParseAssignment() Statement {
	p.Log()
	assignmentOperators := []string{"=", "+=", "-=", "*=", "/=", "%="}
	ret := Assignment{
		Position: p.CurrentToken.Position,
	}
	if p.CurrentToken.Type != TypeID || !contains(assignmentOperators, p.NextToken.Value) {
		return nil
	}
	ret.Variable = p.CurrentToken.Value
	p.Advance()
	ret.Operator = p.CurrentToken.Value
	p.Advance()
	exp := p.This.ParseExpression()
	if exp == nil {
		p.ErrorCurrent("Expected expression on right side of assignment")
	}
	ret.Value = exp
	return &ret
}

// ParseIf parses an if-node
func (p *Parser) ParseIf() Statement {
	p.Log()
	ret := IfStatement{
		Position: p.CurrentToken.Position,
	}
	if p.CurrentToken.Type != TypeKeyword || p.CurrentToken.Value != "if" {
		return nil
	}
	p.Advance()

	ret.Condition = p.This.ParseExpression()
	if ret.Condition == nil {
		p.ErrorCurrent("No expression found as if-condition")
	}

	p.Expect(TypeKeyword, "then")

	stmt := p.This.ParseStatement()
	if stmt == nil {
		p.ErrorCurrent("If-block needs at least one statement")
	}
	ret.IfBlock = make([]Statement, 0, 1)
	ret.IfBlock = append(ret.IfBlock, stmt)

	for {
		stmt2 := p.This.ParseStatement()
		if stmt2 == nil {
			break
		}
		ret.IfBlock = append(ret.IfBlock, stmt2)
	}

	if p.CurrentToken.Type == TypeKeyword && p.CurrentToken.Value == "else" {
		p.Advance()
		stmt := p.This.ParseStatement()
		if stmt == nil {
			p.ErrorCurrent("Else-block needs at least one statement")
		}
		ret.ElseBlock = make([]Statement, 0, 1)
		ret.ElseBlock = append(ret.ElseBlock, stmt)

		for {
			stmt2 := p.This.ParseStatement()
			if stmt2 == nil {
				break
			}
			ret.ElseBlock = append(ret.IfBlock, stmt2)
		}
	}

	p.Expect(TypeKeyword, "end")

	return &ret
}

// ParseExpression parses an expression
func (p *Parser) ParseExpression() Expression {
	p.Log()
	return p.This.ParseLogicExpression()
}

// ParseLogicExpression parses a logical expression
func (p *Parser) ParseLogicExpression() Expression {
	p.Log()
	var exp Expression

	exp = p.This.ParseCompareExpression()
	if exp == nil {
		return nil
	}
	logOps := []string{"or", "and"}

	for p.CurrentToken.Type == TypeKeyword && contains(logOps, p.CurrentToken.Value) {
		binexp := &BinaryOperation{
			Operator: p.CurrentToken.Value,
			Exp1:     exp,
		}
		p.Advance()
		binexp.Exp2 = p.This.ParseCompareExpression()
		if binexp.Exp2 == nil {
			p.ErrorCurrent(fmt.Sprintf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp
}

// ParseCompareExpression parses a compare expression
func (p *Parser) ParseCompareExpression() Expression {
	p.Log()
	exp1 := p.This.ParseSumExpression()
	if exp1 == nil {
		return nil
	}
	logOps := []string{"==", "!=", "<=", ">=", "<", ">"}

	if p.CurrentToken.Type == TypeSymbol && contains(logOps, p.CurrentToken.Value) {
		binexp := &BinaryOperation{
			Operator: p.CurrentToken.Value,
			Exp1:     exp1,
		}
		p.Advance()
		binexp.Exp2 = p.This.ParseSumExpression()
		if binexp.Exp2 == nil {
			p.ErrorCurrent(fmt.Sprintf("Expected expression on right side of %s", binexp.Operator))
		}
		return binexp
	}
	return exp1
}

// ParseSumExpression parses a sum-expression
func (p *Parser) ParseSumExpression() Expression {
	p.Log()
	var exp Expression

	exp = p.This.ParseProdExpression()
	if exp == nil {
		return nil
	}
	logOps := []string{"+", "-"}

	for p.CurrentToken.Type == TypeSymbol && contains(logOps, p.CurrentToken.Value) {
		binexp := &BinaryOperation{
			Operator: p.CurrentToken.Value,
			Exp1:     exp,
		}
		p.Advance()
		binexp.Exp2 = p.This.ParseProdExpression()
		if binexp.Exp2 == nil {
			p.ErrorCurrent(fmt.Sprintf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp
}

// ParseProdExpression parses a product expression
func (p *Parser) ParseProdExpression() Expression {
	p.Log()
	var exp Expression

	exp = p.This.ParseUnaryExpression()
	if exp == nil {
		return nil
	}
	logOps := []string{"*", "/", "%", "^"}

	for p.CurrentToken.Type == TypeSymbol && contains(logOps, p.CurrentToken.Value) {
		binexp := &BinaryOperation{
			Operator: p.CurrentToken.Value,
			Exp1:     exp,
		}
		p.Advance()
		binexp.Exp2 = p.This.ParseUnaryExpression()
		if binexp.Exp2 == nil {
			p.ErrorCurrent(fmt.Sprintf("Expected expression on right side of %s", binexp.Operator))
		}
		exp = binexp
	}
	return exp
}

// ParseUnaryExpression parses an unary expression
func (p *Parser) ParseUnaryExpression() Expression {
	p.Log()
	preUnaryOps := []string{"not", "-"}
	if contains(preUnaryOps, p.CurrentToken.Value) {
		unaryExp := &UnaryOperation{
			Operator: p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		subexp := p.This.ParseUnaryExpression()
		if subexp == nil {
			p.ErrorCurrent(fmt.Sprintf("Expected expression on right side of %s", unaryExp.Operator))
		}
		unaryExp.Exp = subexp
		return unaryExp
	}
	return p.This.ParseBracketExpression()
}

// ParseBracketExpression parses a racketed expression
func (p *Parser) ParseBracketExpression() Expression {
	p.Log()
	if p.CurrentToken.Type == TypeSymbol && p.CurrentToken.Value == "(" {
		p.Advance()
		innerExp := p.This.ParseExpression()
		if innerExp == nil {
			p.ErrorCurrent(fmt.Sprintf("Expected expression after '('"))
		}
		p.Expect(TypeSymbol, ")")
		return innerExp
	}
	return p.This.ParseSingleExpression()
}

// ParseSingleExpression parses a single expression
func (p *Parser) ParseSingleExpression() Expression {
	p.Log()

	preOpVarDeref := p.This.ParsePreOpExpression()
	if preOpVarDeref != nil {
		return preOpVarDeref
	}

	postOpVarDeref := p.This.ParsePostOpExpression()
	if postOpVarDeref != nil {
		return postOpVarDeref
	}

	funccall := p.This.ParseFuncCall()
	if funccall != nil {
		return funccall
	}

	if p.CurrentToken.Type == TypeID {
		defer p.Advance()
		return &Dereference{
			Variable: p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
	}
	if p.CurrentToken.Type == TypeString {
		defer p.Advance()
		return &StringConstant{
			Value:    p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
	}
	if p.CurrentToken.Type == TypeNumber {
		defer p.Advance()
		return &NumberConstant{
			Value:    p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
	}

	// log error here and remove nil checks in other expression functions?
	return nil
}

// ParseFuncCall parse a function call
func (p *Parser) ParseFuncCall() Expression {
	p.Log()
	if p.CurrentToken.Type != TypeID || p.NextToken.Type != TypeSymbol || p.NextToken.Value != "(" {
		return nil
	}
	fc := &FuncCall{
		Function: p.CurrentToken.Value,
	}
	p.Advance()
	p.Advance()
	arg := p.This.ParseExpression()
	fc.Argument = arg
	if arg == nil {
		p.ErrorCurrent("Functions need exactly one argument")
	}

	p.Expect(TypeSymbol, ")")

	return fc
}

// ParsePreOpExpression parse pre-expression
func (p *Parser) ParsePreOpExpression() Expression {
	p.Log()
	if p.CurrentToken.Type == TypeSymbol && (p.CurrentToken.Value == "++" || p.CurrentToken.Value == "--") {
		exp := Dereference{
			Operator: p.CurrentToken.Value,
			PrePost:  "Pre",
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		if p.CurrentToken.Type != TypeID {
			p.Error("Pre- Increment/Decrement must be followed by a variable", exp.Start(), exp.Start())
		}
		exp.Variable = p.CurrentToken.Value
		p.Advance()
		return &exp
	}
	return nil
}

// ParsePostOpExpression parse post-expression
func (p *Parser) ParsePostOpExpression() Expression {
	p.Log()
	if p.NextToken.Type == TypeSymbol && (p.NextToken.Value == "++" || p.NextToken.Value == "--") && p.CurrentToken.Type == TypeID {
		exp := Dereference{
			Variable: p.CurrentToken.Value,
			Operator: p.NextToken.Value,
			PrePost:  "Post",
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		p.Advance()
		return &exp
	}
	return nil
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
		fmt.Println("Called:", callingFunctionName(), "from line", callingLineNumber(), "with", p.CurrentToken.String())
	}
}
