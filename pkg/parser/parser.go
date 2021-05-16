package parser

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Parser parses a yolol programm into an AST
type Parser struct {
	// The tokenizer used to tokenize the input
	Tokenizer    *ast.Tokenizer
	CurrentToken *ast.Token
	// using an interface of ourself to call the parsing-methods allows them to be overridden by 'subclasses'
	This YololParserFunctions
	// Contains all errors encountered during parsing
	Errors Errors
	// If true, return all found errors, not only one per line
	AllErrors bool
	// If true, print debug logs
	DebugLog bool
	// Is true when the current token was preceeded by whitespace
	skippedWhitespace bool
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
	ParsePreOrPostOperation() *ast.Dereference
	ParseGoto() ast.Statement
	ParseAssignment() ast.Statement
	ParseIf() ast.Statement
	ParseExpression() ast.Expression
	ParseBinaryExpression(int) ast.Expression
	ParseUnaryExpression() ast.Expression
	ParseUnaryNotExpression() ast.Expression
	ParseFactorioalExpression() ast.Expression
	ParseNegationExpression() ast.Expression
	ParseBracketExpression() ast.Expression
	ParseSingleExpression() ast.Expression
	ParsePreOpExpression() ast.Expression
	ParsePostOpExpression() ast.Expression
}

// YololParser is an interface that hides all overridable-methods from normal users
type YololParser interface {
	Parse(prog string) (*ast.Program, error)
	SetDebugLog(b bool)
	SetAllErrors(b bool)
}

// NewParser creates a new parser
func NewParser() YololParser {
	p := &Parser{
		Tokenizer: ast.NewTokenizer(),
	}
	p.This = p
	return p
}

// SetDebugLog enables/disables debug-logging
func (p *Parser) SetDebugLog(b bool) {
	p.DebugLog = b
}

// SetAllErrors enables/disables the logging of more then one error per line
func (p *Parser) SetAllErrors(b bool) {
	p.AllErrors = b
}

// ---------------------------------------------

// HasNext returns true if there is a next token
func (p *Parser) HasNext() bool {
	return !p.IsCurrentType(ast.TypeEOF)
}

// Advance advances the current token to the next (non whitespace) token in the list
func (p *Parser) Advance() *ast.Token {
	p.CurrentToken = p.Tokenizer.Next()
	p.skippedWhitespace = false
	for p.CurrentToken.Type == ast.TypeWhitespace {
		p.skippedWhitespace = true
		p.CurrentToken = p.Tokenizer.Next()
	}
	return p.CurrentToken
}

// IsCurrentType checks if the type of the current token is equal to the given type
func (p *Parser) IsCurrentType(t string) bool {
	return p.CurrentToken.Type == t
}

// IsCurrentValue checks if the value of the current token is equal to the given value
// The comparison is case-insensitive
func (p *Parser) IsCurrentValue(value string) bool {
	return strings.ToLower(p.CurrentToken.Value) == strings.ToLower(value)
}

// IsCurrent checks if the current token matches the given type and value
// The comparison of the value is case-insensitive
func (p *Parser) IsCurrent(t, value string) bool {
	return p.IsCurrentType(t) && p.IsCurrentValue(value)
}

// IsCurrentValueIn checks if the value of the current token is one of the provided values
// The comparison is case-insensitive
func (p *Parser) IsCurrentValueIn(values []string) bool {
	curval := strings.ToLower(p.CurrentToken.Value)
	for _, v := range values {
		if curval == strings.ToLower(v) {
			return true
		}
	}
	return false
}

// Error appends an error to the list of errors encountered during parsing
// if p.AllErrors is false, only the first error per line is appended to the list of errors
// Usually it is better to use one of the higher-level Error* methods
func (p *Parser) Error(err *Error) {

	// if not disabled, only log the first error for each line and discard the rest
	if !p.AllErrors {
		if len(p.Errors) > 0 {
			prevError := p.Errors[len(p.Errors)-1]
			if prevError.StartPosition.Line == err.StartPosition.Line {
				return
			}
		}
	}

	err.Message += ". Found Token: '" + p.CurrentToken.Value + "'(" + p.CurrentToken.Type + ")"

	p.Errors = append(p.Errors, err)
}

// ErrorString logs an error with the given message and type at the current location
func (p *Parser) ErrorString(message string, code string) {
	p.Error(&Error{
		Message:       message,
		Code:          code,
		StartPosition: p.CurrentToken.Position,
		EndPosition:   p.CurrentToken.Position,
	})
}

// Expect checks if the current token has the given type and value
// if true, the tokens position is returned, otherwise an error is logged
// alsways advances to the next token
func (p *Parser) Expect(tokenType string, tokenValue string) ast.Position {
	if !p.IsCurrent(tokenType, tokenValue) {
		var msg string
		if tokenType == ast.TypeNewline {
			msg = "Expected newline"
		} else {
			msg = fmt.Sprintf("Expected token '%s'(%s)", tokenValue, tokenType)
		}
		p.Error(&Error{
			Message: msg,
			Code:    ErrExpectedToken,
			ExpectedToken: &ast.Token{
				Type:     tokenType,
				Value:    tokenValue,
				Position: p.CurrentToken.Position,
			},
			StartPosition: p.CurrentToken.Position,
			EndPosition:   p.CurrentToken.Position,
		})
	}
	pos := p.CurrentToken.Position
	p.Advance()
	return pos
}

// ErrorExpectedExpression is a shortcut to log an error about an expected expression at where
func (p *Parser) ErrorExpectedExpression(where string) {
	p.ErrorString("Expected expression "+where, ErrExpectedExpression)
}

// ErrorExpectedStatement is a shortcut to log an error about an expected statement at where
func (p *Parser) ErrorExpectedStatement(where string) {
	p.ErrorString("Expected statement "+where, ErrExpectedStatement)
}

// Reset prepares all internal fields for a new parsing run
// is called automatically by Parse(). Overriding structs must call this in their Parse()
func (p *Parser) Reset() {
	p.Errors = make(Errors, 0)
	p.CurrentToken = nil
}

// ---------------------------------------------

// Parse is the main method of the parser. Parses a yolol-program into an AST.
func (p *Parser) Parse(prog string) (*ast.Program, error) {
	p.Reset()
	p.Tokenizer.Load(prog)
	// Advance twice to fill CurrentToken
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
	for !p.IsCurrentType(ast.TypeNewline) && !p.IsCurrentType(ast.TypeEOF) {
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

	if p.IsCurrentType(ast.TypeComment) {
		ret.Comment = p.CurrentToken.Value
		p.Advance()
	}

	// not statements in this line
	if p.IsCurrentType(ast.TypeNewline) || p.IsCurrentType(ast.TypeEOF) {
		p.Advance()
		return &ret
	}

	for p.HasNext() {
		stmt := p.This.ParseStatement()
		if stmt == nil {
			p.ErrorExpectedStatement("")
			p.Advance()
		}
		ret.Statements = append(ret.Statements, stmt)
		hadWhitespace := p.skippedWhitespace

		if p.IsCurrentType(ast.TypeComment) {
			ret.Comment = p.CurrentToken.Value
			p.Advance()
		}

		// line ends after statement (or after comment)
		if p.IsCurrentType(ast.TypeNewline) || p.IsCurrentType(ast.TypeEOF) {
			p.Advance()
			return &ret
		}

		if !hadWhitespace {
			p.ErrorString("Statements on a line must be separated by spaces", "")
			return &ret
		}
	}

	if !p.IsCurrentType(ast.TypeEOF) {
		p.Error(&Error{
			Message:       "Missing newline",
			StartPosition: ret.Start(),
			EndPosition:   ret.End(),
		})
	}

	return &ret
}

// ParseStatement parses a statement-node
func (p *Parser) ParseStatement() ast.Statement {
	p.Log()

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

	stmt := p.This.ParseAssignment()
	if stmt != nil {
		return stmt
	}

	return nil
}

// ParsePreOrPostOperation parses a pre-/post operation (x++, ++x) as a statement
func (p *Parser) ParsePreOrPostOperation() *ast.Dereference {
	p.Log()
	preOpVarDeref := p.This.ParsePreOpExpression()
	if preOpVarDeref != nil {
		preOpVarDeref.(*ast.Dereference).IsStatement = true
		return preOpVarDeref.(*ast.Dereference)
	}

	postOpVarDeref := p.This.ParsePostOpExpression()
	if postOpVarDeref != nil {
		postOpVarDeref.(*ast.Dereference).IsStatement = true
		return postOpVarDeref.(*ast.Dereference)
	}

	return nil
}

// ParseGoto parse parses a goto-node
func (p *Parser) ParseGoto() ast.Statement {
	if p.IsCurrent(ast.TypeKeyword, "goto") {
		stmt := ast.GoToStatement{
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		stmt.Line = p.This.ParseExpression()
		if stmt.Line == nil {
			p.ErrorExpectedExpression("after goto")
		}
		if _, is := stmt.Line.(*ast.StringConstant); is {
			p.ErrorString("Can not goto a string", "")
		}
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
	if !p.IsCurrentType(ast.TypeID) {
		return nil
	}

	ret.Variable = p.CurrentToken.Value
	p.Advance()

	if !p.IsCurrentValueIn(assignmentOperators) {
		p.ErrorString("Expected an assignment-operator", ErrExpectedAssignop)
		return &ret
	}

	ret.Operator = p.CurrentToken.Value
	p.Advance()

	exp := p.This.ParseExpression()
	if exp == nil {
		p.ErrorExpectedExpression("on right side of assignment")
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
	if !p.IsCurrent(ast.TypeKeyword, "if") {
		return nil
	}
	p.Advance()

	ret.Condition = p.This.ParseExpression()
	if ret.Condition == nil {
		p.ErrorExpectedExpression("as if-condition")
	}

	p.Expect(ast.TypeKeyword, "then")
	ret.IfBlock = make([]ast.Statement, 0, 1)

	for p.HasNext() {
		hadWhitespace := p.skippedWhitespace
		stmt := p.This.ParseStatement()
		if stmt == nil {
			break
		}
		ret.IfBlock = append(ret.IfBlock, stmt)
		if len(ret.IfBlock) > 1 && !hadWhitespace {
			p.Error(&Error{
				Message:       "Statements on a line must be separated by spaces",
				StartPosition: stmt.Start(),
				EndPosition:   stmt.End(),
			})
			return &ret
		}
	}

	if p.IsCurrent(ast.TypeKeyword, "else") {
		p.Advance()

		ret.ElseBlock = make([]ast.Statement, 0, 1)

		for p.HasNext() {
			hadWhitespace := p.skippedWhitespace
			stmt := p.This.ParseStatement()
			if stmt == nil {
				break
			}
			ret.ElseBlock = append(ret.ElseBlock, stmt)
			if len(ret.ElseBlock) > 1 && !hadWhitespace {
				p.Error(&Error{
					Message:       "Statements on a line must be separated by spaces",
					StartPosition: stmt.Start(),
					EndPosition:   stmt.End(),
				})
				return &ret
			}
		}
	}

	p.Expect(ast.TypeKeyword, "end")

	return &ret
}

// ParseExpression parses an expression
func (p *Parser) ParseExpression() ast.Expression {
	p.Log()
	return p.This.ParseBinaryExpression(0)
}

// ParseBinaryExpression parses a binary expression
// The kind of binary-expression to be parsed is given as idx
func (p *Parser) ParseBinaryExpression(idx int) ast.Expression {
	p.LogI(idx)

	var exp *ast.BinaryOperation
	var ops []string
	expectedType := ast.TypeSymbol
	leftAssoc := true

	switch idx {
	case 0:
		ops = []string{"and"}
		expectedType = ast.TypeKeyword
		break
	case 1:
		ops = []string{"or"}
		expectedType = ast.TypeKeyword
		break
	case 2:
		return p.ParseUnaryNotExpression()
	case 3:
		ops = []string{"+", "-"}
		break
	case 4:
		ops = []string{"==", "!=", "<=", ">=", "<", ">"}
		break
	case 5:
		ops = []string{"*", "/", "%"}
		break
	case 6:
		ops = []string{"^"}
		leftAssoc = false
		break
	default:
		return p.This.ParseUnaryExpression()
	}

	idx++

	leftExp := p.This.ParseBinaryExpression(idx)
	if leftExp == nil {
		return nil
	}

	for p.IsCurrentType(expectedType) && p.IsCurrentValueIn(ops) {
		binexp := &ast.BinaryOperation{
			Operator: p.CurrentToken.Value,
		}

		p.Advance()
		rightExp := p.This.ParseBinaryExpression(idx)
		if rightExp == nil {
			p.ErrorExpectedExpression(fmt.Sprintf("on right side of %s", binexp.Operator))
			return binexp
		}

		if exp == nil {
			binexp.Exp1 = leftExp
			binexp.Exp2 = rightExp
			exp = binexp
			continue
		}

		if leftAssoc {
			binexp.Exp1 = exp
			binexp.Exp2 = rightExp
			exp = binexp
		} else {
			binexp.Exp1 = exp.Exp2
			binexp.Exp2 = rightExp
			exp.Exp2 = binexp
		}
	}

	if exp == nil {
		return leftExp
	}

	return exp
}

// ParseUnaryNotExpression parses a "not" expression
func (p *Parser) ParseUnaryNotExpression() ast.Expression {
	p.Log()
	if p.IsCurrent(ast.TypeKeyword, "not") {
		unaryExp := &ast.UnaryOperation{
			Operator: p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		subexp := p.This.ParseUnaryNotExpression()
		if subexp == nil {
			p.ErrorExpectedExpression(fmt.Sprintf("on right side of %s", unaryExp.Operator))
		}
		unaryExp.Exp = subexp
		return unaryExp
	}
	return p.This.ParseBinaryExpression(3)
}

// ParseUnaryExpression parses an unary expression
func (p *Parser) ParseUnaryExpression() ast.Expression {
	p.Log()
	preUnaryOps := []string{"abs", "sqrt", "sin", "cos", "tan", "asin", "acos", "atan"}
	if p.IsCurrentValueIn(preUnaryOps) && p.IsCurrentType(ast.TypeKeyword) {
		unaryExp := &ast.UnaryOperation{
			Operator: p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		subexp := p.This.ParseUnaryExpression()
		if subexp == nil {
			p.ErrorExpectedExpression(fmt.Sprintf("on right side of %s", unaryExp.Operator))
		}
		unaryExp.Exp = subexp
		return unaryExp
	}
	return p.This.ParseNegationExpression()
}

// ParseNegationExpression parses a negation
func (p *Parser) ParseNegationExpression() ast.Expression {
	p.Log()

	if p.IsCurrent(ast.TypeSymbol, "-") {
		unaryExp := &ast.UnaryOperation{
			Operator: p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		subexp := p.This.ParseNegationExpression()
		if subexp == nil {
			p.ErrorExpectedExpression(fmt.Sprintf("on right side of %s", unaryExp.Operator))
		}
		// Usualy, just parsing the positive number and wrapping it in a negation works fine
		// But 9223372036854775.808 can not be stored inside the 64bit number-type. -9223372036854775.808 can.
		// So as a workaround, a negation of this special number is folded into the constant itself
		if nconst, is := subexp.(*ast.NumberConstant); is {
			if nconst.Value == "9223372036854775.808" {
				nconst.Value = "-9223372036854775.808"
				return nconst
			}
		}
		unaryExp.Exp = subexp
		return unaryExp
	}
	return p.This.ParseFactorioalExpression()
}

// ParseFactorioalExpression parses a factorial
func (p *Parser) ParseFactorioalExpression() ast.Expression {
	p.Log()

	subexp := p.This.ParseBracketExpression()
	if subexp != nil {
		for p.IsCurrent(ast.TypeSymbol, "!") {
			subexp = &ast.UnaryOperation{
				Operator: "!",
				Position: p.CurrentToken.Position,
				Exp:      subexp,
			}
			p.Advance()
		}
	}
	return subexp
}

// ParseBracketExpression parses a racketed expression
func (p *Parser) ParseBracketExpression() ast.Expression {
	p.Log()
	if p.IsCurrent(ast.TypeSymbol, "(") {
		p.Advance()
		innerExp := p.This.ParseExpression()
		if innerExp == nil {
			p.ErrorExpectedExpression("after '('")
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

	if p.IsCurrentType(ast.TypeID) {
		defer p.Advance()
		return &ast.Dereference{
			Variable: p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
	}
	if p.IsCurrentType(ast.TypeString) {
		defer p.Advance()
		return &ast.StringConstant{
			Value:    p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
	}
	if p.IsCurrentType(ast.TypeNumber) {
		defer p.Advance()
		return &ast.NumberConstant{
			Value:    p.CurrentToken.Value,
			Position: p.CurrentToken.Position,
		}
	}

	// log error here and remove nil checks in other expression functions?
	return nil
}

// ParsePreOpExpression parse pre-expression
func (p *Parser) ParsePreOpExpression() ast.Expression {
	p.Log()
	if p.IsCurrentType(ast.TypeSymbol) && (p.IsCurrentValue("++") || p.IsCurrentValue("--")) {
		exp := ast.Dereference{
			Operator: p.CurrentToken.Value,
			PrePost:  "Pre",
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		if !p.IsCurrentType(ast.TypeID) {
			p.Error(&Error{
				Message:       "Pre- Increment/Decrement must be followed by a variable",
				StartPosition: exp.Start(),
				EndPosition:   exp.Start(),
			})
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
	nextToken := p.Tokenizer.Peek()
	if nextToken.Type == ast.TypeSymbol && (nextToken.Value == "++" || nextToken.Value == "--") && p.IsCurrentType(ast.TypeID) {
		exp := ast.Dereference{
			Variable: p.CurrentToken.Value,
			Operator: nextToken.Value,
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

// LogI logs the visiting of a parsing function with a given argument
func (p *Parser) LogI(arg int) {
	if p.DebugLog {
		// Print the name of the function
		fmt.Println("Called:", callingFunctionName(), "from line", callingLineNumber(), "with", p.CurrentToken.String(), "arg", arg)
	}
}
