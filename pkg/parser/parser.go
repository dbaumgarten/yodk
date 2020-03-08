package parser

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Parser parses a yolol programm into an AST
type Parser struct {
	Tokenizer    *ast.Tokenizer
	CurrentToken *ast.Token
	NextToken    *ast.Token
	PrevToken    *ast.Token
	// if true, there was whitespace between CurrentToken and NextToken
	NextWouldBeWhitespace bool
	// if true, current token was preceeded by whitespace
	SkippedWhitespace bool
	// using an interface of ourself to call the parsing-methods allows them to be overridden by 'subclasses'
	This YololParserFunctions
	// Contains all errors encountered during parsing
	Errors Errors
	// If true, return all found errors, not only one per line
	AllErrors bool
	// If true, print debug logs
	DebugLog bool
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
	ParseStatement() ast.Statement
	ParsePreOrPostOperation() ast.Statement
	ParseGoto() ast.Statement
	ParseAssignment() ast.Statement
	ParseIf() ast.Statement
	ParseExpression() ast.Expression
	ParseLogicExpression() ast.Expression
	ParseCompareExpression() ast.Expression
	ParseSumExpression() ast.Expression
	ParseProdExpression() ast.Expression
	ParseUnaryExpression() ast.Expression
	ParseBracketExpression() ast.Expression
	ParseSingleExpression() ast.Expression
	ParseFuncCall() ast.Expression
	ParsePreOpExpression() ast.Expression
	ParsePostOpExpression() ast.Expression
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
	return p.CurrentToken.Type != ast.TypeEOF
}

// Advance advances the current token to the next (non whitespace) token in the list
func (p *Parser) Advance() *ast.Token {
	if p.CurrentToken == nil || p.HasNext() {
		p.PrevToken = p.CurrentToken
		p.CurrentToken = p.NextToken
		p.NextToken = p.Tokenizer.Next()
		p.SkippedWhitespace = p.NextWouldBeWhitespace
		p.NextWouldBeWhitespace = false
		for p.NextToken.Type == ast.TypeWhitespace {
			if p.NextToken.Type == ast.TypeWhitespace {
				p.NextWouldBeWhitespace = true
			}
			p.NextToken = p.Tokenizer.Next()
		}

	}
	return p.CurrentToken
}

// Error appends an error to the list of errors encountered during parsing
// if p.AllErrors is false, only the first error per line is appended to the list of errors
func (p *Parser) Error(msg string, start ast.Position, end ast.Position) {

	// if not disabled, only log the first error for each line and discard the rest
	if !p.AllErrors {
		if len(p.Errors) > 0 {
			prevError := p.Errors[len(p.Errors)-1]
			if prevError.StartPosition.Line == start.Line {
				return
			}
		}
	}

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
func (p *Parser) Expect(tokenType string, tokenValue string) ast.Position {
	if p.CurrentToken.Type != tokenType || p.CurrentToken.Value != tokenValue {
		var msg string
		if tokenType == ast.TypeNewline {
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
	p.Tokenizer = ast.NewTokenizer()
	p.Errors = make(Errors, 0)
	p.CurrentToken = nil
	p.PrevToken = nil
	p.NextToken = nil
	p.NextWouldBeWhitespace = false
	p.SkippedWhitespace = false
}

// ---------------------------------------------

// Parse is the main method of the parser. Parses a yolol-program into an AST.
func (p *Parser) Parse(prog string) (*ast.Program, error) {
	p.init()
	p.Tokenizer.Load(prog)
	// Advance twice to fill CurrentToken and NextToken
	p.Advance()
	p.Advance()
	parsed := p.ParseProgram()
	if len(p.Errors) == 0 {
		return parsed, nil
	}
	return nil, p.Errors
}

// ParseProgram parses a programm-node
func (p *Parser) ParseProgram() *ast.Program {
	p.Log()
	ret := ast.Program{
		Lines: make([]*ast.Line, 0),
	}
	for p.HasNext() {
		ret.Lines = append(ret.Lines, p.ParseLine())
	}
	return &ret
}

// SkipLine skips tokens up to the next newline
func (p *Parser) SkipLine() {
	for p.CurrentToken.Type != ast.TypeNewline && p.CurrentToken.Type != ast.TypeEOF {
		p.Advance()
	}
}

// ParseLine parses a line-node
func (p *Parser) ParseLine() *ast.Line {
	p.Log()
	ret := ast.Line{
		Position:   p.CurrentToken.Position,
		Statements: make([]ast.Statement, 0),
	}

	if p.CurrentToken.Type == ast.TypeComment {
		ret.Comment = p.CurrentToken.Value
		p.Advance()
	}

	// not statements in this line
	if p.CurrentToken.Type == ast.TypeNewline || p.CurrentToken.Type == ast.TypeEOF {
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

		if p.CurrentToken.Type == ast.TypeComment {
			ret.Comment = p.CurrentToken.Value
			p.Advance()
		}

		// line ends after statement (or after comment)
		if p.CurrentToken.Type == ast.TypeNewline || p.CurrentToken.Type == ast.TypeEOF {
			p.Advance()
			return &ret
		}

		// more statements on this line?
		if !p.SkippedWhitespace {
			p.ErrorCurrent("Statements must be followed by space or newline")
		}
	}

	if p.CurrentToken.Type != ast.TypeEOF {
		p.Error("Missing newline", ret.Start(), ret.End())
	}

	return &ret
}

// ParseStatement parses a statement-node
func (p *Parser) ParseStatement() ast.Statement {
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
func (p *Parser) ParsePreOrPostOperation() ast.Statement {
	p.Log()
	preOpVarDeref := p.This.ParsePreOpExpression()
	if preOpVarDeref != nil {
		preOpVarDeref.(*ast.Dereference).IsStatement = true
		return preOpVarDeref
	}

	postOpVarDeref := p.This.ParsePostOpExpression()
	if postOpVarDeref != nil {
		postOpVarDeref.(*ast.Dereference).IsStatement = true
		return postOpVarDeref
	}

	return nil
}

// ParseGoto parse parses a goto-node
func (p *Parser) ParseGoto() ast.Statement {
	if p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "goto" {
		stmt := ast.GoToStatement{
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		line, err := strconv.Atoi(p.CurrentToken.Value)
		if p.CurrentToken.Type != ast.TypeNumber || err != nil {
			p.Error("Goto must be followed by a line number", stmt.Start(), stmt.Start())
		}
		stmt.Line = line
		p.Advance()
		return &stmt
	}
	return nil
}

// ParseAssignment parses an assignment-node
func (p *Parser) ParseAssignment() ast.Statement {
	p.Log()
	assignmentOperators := []string{"=", "+=", "-=", "*=", "/=", "%="}
	ret := ast.Assignment{
		Position: p.CurrentToken.Position,
	}
	if p.CurrentToken.Type != ast.TypeID || !contains(assignmentOperators, p.NextToken.Value) {
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
func (p *Parser) ParseIf() ast.Statement {
	p.Log()
	ret := ast.IfStatement{
		Position: p.CurrentToken.Position,
	}
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "if" {
		return nil
	}
	p.Advance()

	ret.Condition = p.This.ParseExpression()
	if ret.Condition == nil {
		p.ErrorCurrent("No expression found as if-condition")
	}

	p.Expect(ast.TypeKeyword, "then")

	stmt := p.This.ParseStatement()
	if stmt == nil {
		p.ErrorCurrent("If-block needs at least one statement")
	}
	ret.IfBlock = make([]ast.Statement, 0, 1)
	ret.IfBlock = append(ret.IfBlock, stmt)

	for {
		stmt2 := p.This.ParseStatement()
		if stmt2 == nil {
			break
		}
		ret.IfBlock = append(ret.IfBlock, stmt2)
	}

	if p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "else" {
		p.Advance()
		stmt := p.This.ParseStatement()
		if stmt == nil {
			p.ErrorCurrent("Else-block needs at least one statement")
		}
		ret.ElseBlock = make([]ast.Statement, 0, 1)
		ret.ElseBlock = append(ret.ElseBlock, stmt)

		for {
			stmt2 := p.This.ParseStatement()
			if stmt2 == nil {
				break
			}
			ret.ElseBlock = append(ret.IfBlock, stmt2)
		}
	}

	p.Expect(ast.TypeKeyword, "end")

	return &ret
}

// ParseExpression parses an expression
func (p *Parser) ParseExpression() ast.Expression {
	p.Log()
	return p.This.ParseLogicExpression()
}

// ParseLogicExpression parses a logical expression
func (p *Parser) ParseLogicExpression() ast.Expression {
	p.Log()
	var exp ast.Expression

	exp = p.This.ParseCompareExpression()
	if exp == nil {
		return nil
	}
	logOps := []string{"or", "and"}

	for p.CurrentToken.Type == ast.TypeKeyword && contains(logOps, p.CurrentToken.Value) {
		binexp := &ast.BinaryOperation{
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
func (p *Parser) ParseCompareExpression() ast.Expression {
	p.Log()
	exp1 := p.This.ParseSumExpression()
	if exp1 == nil {
		return nil
	}
	logOps := []string{"==", "!=", "<=", ">=", "<", ">"}

	if p.CurrentToken.Type == ast.TypeSymbol && contains(logOps, p.CurrentToken.Value) {
		binexp := &ast.BinaryOperation{
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
func (p *Parser) ParseSumExpression() ast.Expression {
	p.Log()
	var exp ast.Expression

	exp = p.This.ParseProdExpression()
	if exp == nil {
		return nil
	}
	logOps := []string{"+", "-"}

	for p.CurrentToken.Type == ast.TypeSymbol && contains(logOps, p.CurrentToken.Value) {
		binexp := &ast.BinaryOperation{
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
func (p *Parser) ParseProdExpression() ast.Expression {
	p.Log()
	var exp ast.Expression

	exp = p.This.ParseUnaryExpression()
	if exp == nil {
		return nil
	}
	logOps := []string{"*", "/", "%", "^"}

	for p.CurrentToken.Type == ast.TypeSymbol && contains(logOps, p.CurrentToken.Value) {
		binexp := &ast.BinaryOperation{
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
func (p *Parser) ParseUnaryExpression() ast.Expression {
	p.Log()
	preUnaryOps := []string{"not", "-"}
	if contains(preUnaryOps, p.CurrentToken.Value) {
		unaryExp := &ast.UnaryOperation{
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
func (p *Parser) ParseBracketExpression() ast.Expression {
	p.Log()
	if p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == "(" {
		p.Advance()
		innerExp := p.This.ParseExpression()
		if innerExp == nil {
			p.ErrorCurrent(fmt.Sprintf("Expected expression after '('"))
		}
		p.Expect(ast.TypeSymbol, ")")
		return innerExp
	}
	return p.This.ParseSingleExpression()
}

// ParseSingleExpression parses a single expression
func (p *Parser) ParseSingleExpression() ast.Expression {
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

	if p.CurrentToken.Type == ast.TypeID {
		defer p.Advance()
		return &ast.Dereference{
			Variable: p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
	}
	if p.CurrentToken.Type == ast.TypeString {
		defer p.Advance()
		return &ast.StringConstant{
			Value:    p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
	}
	if p.CurrentToken.Type == ast.TypeNumber {
		defer p.Advance()
		return &ast.NumberConstant{
			Value:    p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
	}

	// log error here and remove nil checks in other expression functions?
	return nil
}

// ParseFuncCall parse a function call
func (p *Parser) ParseFuncCall() ast.Expression {
	p.Log()
	if p.CurrentToken.Type != ast.TypeID || p.NextToken.Type != ast.TypeSymbol || p.NextToken.Value != "(" {
		return nil
	}
	fc := &ast.FuncCall{
		Position: p.CurrentToken.Position,
		Function: p.CurrentToken.Value,
	}
	p.Advance()
	p.Advance()
	arg := p.This.ParseExpression()
	fc.Argument = arg
	if arg == nil {
		p.ErrorCurrent("Functions need exactly one argument")
	}

	p.Expect(ast.TypeSymbol, ")")

	return fc
}

// ParsePreOpExpression parse pre-expression
func (p *Parser) ParsePreOpExpression() ast.Expression {
	p.Log()
	if p.CurrentToken.Type == ast.TypeSymbol && (p.CurrentToken.Value == "++" || p.CurrentToken.Value == "--") {
		exp := ast.Dereference{
			Operator: p.CurrentToken.Value,
			PrePost:  "Pre",
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		if p.CurrentToken.Type != ast.TypeID {
			p.Error("Pre- Increment/Decrement must be followed by a variable", exp.Start(), exp.Start())
		}
		exp.Variable = p.CurrentToken.Value
		p.Advance()
		return &exp
	}
	return nil
}

// ParsePostOpExpression parse post-expression
func (p *Parser) ParsePostOpExpression() ast.Expression {
	p.Log()
	if p.NextToken.Type == ast.TypeSymbol && (p.NextToken.Value == "++" || p.NextToken.Value == "--") && p.CurrentToken.Type == ast.TypeID {
		exp := ast.Dereference{
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
