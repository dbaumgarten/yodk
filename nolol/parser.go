package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/parser"
)

// Parser pasres a nolol-program
type Parser struct {
	*parser.Parser
}

// NewParser creates and returns a nolol parser
func NewParser() *Parser {
	ep := &Parser{
		Parser: parser.NewParser(),
	}
	ep.Tokenizer = NewNololTokenizer()
	ep.This = ep
	return ep
}

// Parse is the entry point for parsing
func (p *Parser) Parse(prog string) (*Program, error) {
	errors := make(parser.Errors, 0)
	p.Tokenizer.Load(prog)
	// Advance twice to fill CurrentToken and NextToken
	p.Advance()
	p.Advance()
	parsed, err := p.ParseProgram()
	errors = append(errors, err...)
	if len(errors) > 0 {
		return nil, errors
	}
	return parsed, nil
}

// ParseProgram parses the program
func (p *Parser) ParseProgram() (*Program, parser.Errors) {
	p.Log()
	errors := make(parser.Errors, 0)
	ret := Program{
		Lines: make([]Line, 0),
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

// ParseStatementLine parses a statement line
func (p *Parser) ParseStatementLine() (*StatementLine, *parser.Error) {
	ret := StatementLine{
		Line: parser.Line{
			Statements: make([]parser.Statement, 0, 1),
		},
		Position: p.CurrentToken.Position,
	}

	// get line-label if it exists
	if p.CurrentToken.Type == parser.TypeID && (p.NextToken.Type == parser.TypeSymbol && p.NextToken.Value == ">") {
		ret.Label = p.CurrentToken.Value
		p.Advance()
		p.Advance()
	}

	// the line has no statements
	if p.CurrentToken.Type == parser.TypeEOF || p.CurrentToken.Type == parser.TypeNewline {
		p.Advance()
		return &ret, nil
	}

	stmt, err := p.This.ParseStatement()
	if err != nil {
		return nil, err
	}
	ret.Statements = append(ret.Statements, stmt)

	if p.CurrentToken.Type == parser.TypeEOF || p.CurrentToken.Type == parser.TypeNewline {
		p.Advance()
		return &ret, nil
	}

	return nil, p.NewError("Expected newline after statement", true, ret.Start(), ret.End())
}

// ParseExecutableLine parses an if, while or statement-line
func (p *Parser) ParseExecutableLine() (ExecutableLine, *parser.Error) {

	ifline, err := p.ParseMultilineIf()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return ifline, nil
	}

	whileline, err := p.ParseWhile()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return whileline, nil
	}

	return p.ParseStatementLine()
}

// ParseLine parses any kind of line
func (p *Parser) ParseLine() (Line, *parser.Error) {
	p.Log()

	// skip empty lines and stray newlines
	for p.CurrentToken.Type == parser.TypeNewline {
		p.Advance()
	}

	constDecl, err := p.ParseConstantDeclaration()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return constDecl, nil
	}

	return p.ParseExecutableLine()
}

// ParseConstantDeclaration parses a constant declaration
func (p *Parser) ParseConstantDeclaration() (*ConstDeclaration, *parser.Error) {
	if p.CurrentToken.Type != parser.TypeKeyword || p.CurrentToken.Value != "const" {
		return nil, p.NewError("Const declaration must start with const", false, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	startpos := p.CurrentToken.Position
	p.Advance()
	if p.CurrentToken.Type != parser.TypeID {
		return nil, p.NewError("const keyword must be followed by identifier", true, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	decl := &ConstDeclaration{
		Name:     p.CurrentToken.Value,
		Position: startpos,
	}
	p.Advance()
	if p.CurrentToken.Type != parser.TypeSymbol || p.CurrentToken.Value != "=" {
		return nil, p.NewError("Identifier in const declaration must be followed by '='", true, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	p.Advance()
	value, err := p.ParseSingleExpression()
	if err != nil {
		err.Fatal = true
		return nil, err
	}
	decl.Value = value
	if p.CurrentToken.Type != parser.TypeNewline && p.CurrentToken.Type != parser.TypeEOF {
		return nil, p.NewError("Const declaration must be followed by newline or EOF", true, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	p.Advance()
	return decl, nil
}

// ParseLinesUntil parse lines until stop() returns true
func (p *Parser) ParseLinesUntil(stop func() bool) ([]ExecutableLine, *parser.Error) {
	lines := make([]ExecutableLine, 0)
	for p.HasNext() && !stop() {
		line, err := p.ParseExecutableLine()
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// ParseMultilineIf parses a nolol-style multiline if
func (p *Parser) ParseMultilineIf() (Line, *parser.Error) {
	p.Log()
	mlif := MultilineIf{
		Position: p.CurrentToken.Position,
	}
	if p.CurrentToken.Type != parser.TypeKeyword || p.CurrentToken.Value != "if" {
		return nil, p.NewError("If-statements have to start with 'if'", false, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	p.Advance()

	var err *parser.Error
	mlif.Condition, err = p.This.ParseExpression()
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("No expression found as if-condition"))
	}

	if p.CurrentToken.Type != parser.TypeKeyword || p.CurrentToken.Value != "then" {
		return nil, p.NewError("Expected 'then' after condition", true, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	p.Advance()

	if p.CurrentToken.Type != parser.TypeNewline {
		return nil, p.NewError("Expected newline after then", true, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	p.Advance()
	// mulitline if

	mlif.IfBlock, err = p.ParseLinesUntil(func() bool {
		return p.CurrentToken.Type == parser.TypeKeyword && (p.CurrentToken.Value == "end" || p.CurrentToken.Value == "else")
	})
	if err != nil {
		return nil, err
	}

	if p.CurrentToken.Type == parser.TypeKeyword && p.CurrentToken.Value == "else" {
		p.Advance()
		mlif.ElseBlock, err = p.ParseLinesUntil(func() bool {
			return p.CurrentToken.Type == parser.TypeKeyword && p.CurrentToken.Value == "end"
		})
		if err != nil {
			return nil, err
		}
	}

	if p.CurrentToken.Type != parser.TypeKeyword || p.CurrentToken.Value != "end" {
		return nil, p.NewError("Expected 'end' after if statement", true, mlif.Start(), mlif.Start())
	}
	p.Advance()

	if p.CurrentToken.Type != parser.TypeNewline && p.CurrentToken.Type != parser.TypeEOF {
		return nil, p.NewError("End must be followed by newline or EOF", true, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	p.Advance()

	return &mlif, nil
}

// ParseWhile pasres a nolol while
func (p *Parser) ParseWhile() (Line, *parser.Error) {
	p.Log()
	loop := WhileLoop{
		Position: p.CurrentToken.Position,
	}
	if p.CurrentToken.Type != parser.TypeKeyword || p.CurrentToken.Value != "while" {
		return nil, p.NewError("While-statements have to start with 'while'", false, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	p.Advance()

	var err *parser.Error
	loop.Condition, err = p.This.ParseExpression()
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("No expression found as loop-condition"))
	}

	if p.CurrentToken.Type != parser.TypeKeyword || p.CurrentToken.Value != "do" {
		return nil, p.NewError("Expected 'do' after condition", true, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	p.Advance()

	if p.CurrentToken.Type != parser.TypeNewline {
		return nil, p.NewError("Expected newline after do", true, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	p.Advance()

	loop.Block, err = p.ParseLinesUntil(func() bool {
		return p.CurrentToken.Type == parser.TypeKeyword && p.CurrentToken.Value == "end"
	})
	if err != nil {
		return nil, err
	}

	if p.CurrentToken.Type != parser.TypeKeyword || p.CurrentToken.Value != "end" {
		return nil, p.NewError("Expected 'end' after if statement", true, loop.Start(), loop.Start())
	}
	p.Advance()

	if p.CurrentToken.Type != parser.TypeNewline && p.CurrentToken.Type != parser.TypeEOF {
		return nil, p.NewError("End must be followed by newline or EOF", true, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	p.Advance()

	return &loop, nil

}

// ParseIf overrides and disables the old yolol-style inline ifs
func (p *Parser) ParseIf() (parser.Statement, *parser.Error) {
	p.Log()
	return nil, p.NewError("Inline if is not supported by nolol.", false, p.CurrentToken.Position, p.CurrentToken.Position)
}

// ParseGoto allows labeled-gotos and forbids line-based gotos
func (p *Parser) ParseGoto() (parser.Statement, *parser.Error) {
	if p.CurrentToken.Type == parser.TypeKeyword && p.CurrentToken.Value == "goto" {
		p.Advance()

		if p.CurrentToken.Type == parser.TypeID {
			stmt := &GoToLabelStatement{
				Position: p.CurrentToken.Position,
				Label:    p.CurrentToken.Value,
			}
			p.Advance()
			return stmt, nil
		}

		return nil, p.NewError("Goto must be followed by an identifier", true, p.CurrentToken.Position, p.CurrentToken.Position)
	}
	return nil, p.NewError("Goto statements must start with 'goto'", false, p.CurrentToken.Position, p.CurrentToken.Position)
}
