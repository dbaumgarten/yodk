package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/parser"
)

// Parser pasres a nolol-program
type Parser struct {
	*parser.Parser
}

// NewNololParser creates and returns a nolol parser
func NewNololParser() *Parser {
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
	p.Tokens = make([]*parser.Token, 0, 1000)
	for {
		token, err := p.Tokenizer.Next()
		if err != nil {
			errors = append(errors, err.(*parser.Error))
		} else {
			if p.DebugLog {
				fmt.Print(token)
			}
			p.Tokens = append(p.Tokens, token)
			if token.Type == parser.TypeEOF {
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
		Position: p.Current().Position,
	}

	// get line-label if it exists
	if p.Current().Type == parser.TypeID && (p.Next().Type == parser.TypeSymbol && p.Next().Value == ">") {
		ret.Label = p.Current().Value
		p.Advance()
		p.Advance()
	}

	// the line has no statements
	if p.Current().Type == parser.TypeEOF || p.Current().Type == parser.TypeNewline {
		p.Advance()
		return &ret, nil
	}

	stmt, err := p.This.ParseStatement()
	if err != nil {
		return nil, err
	}
	ret.Statements = append(ret.Statements, stmt)

	if p.Current().Type == parser.TypeEOF || p.Current().Type == parser.TypeNewline {
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
	for p.Current().Type == parser.TypeNewline {
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
	if p.Current().Type != parser.TypeKeyword || p.Current().Value != "const" {
		return nil, p.NewError("Const declaration must start with const", false, p.Current().Position, p.Current().Position)
	}
	startpos := p.Current().Position
	p.Advance()
	if p.Current().Type != parser.TypeID {
		return nil, p.NewError("const keyword must be followed by identifier", true, p.Current().Position, p.Current().Position)
	}
	decl := &ConstDeclaration{
		Name:     p.Current().Value,
		Position: startpos,
	}
	p.Advance()
	if p.Current().Type != parser.TypeSymbol || p.Current().Value != "=" {
		return nil, p.NewError("Identifier in const declaration must be followed by '='", true, p.Current().Position, p.Current().Position)
	}
	p.Advance()
	value, err := p.ParseSingleExpression()
	if err != nil {
		err.Fatal = true
		return nil, err
	}
	decl.Value = value
	if p.Current().Type != parser.TypeNewline && p.Current().Type != parser.TypeEOF {
		return nil, p.NewError("Const declaration must be followed by newline or EOF", true, p.Current().Position, p.Current().Position)
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
		Position: p.Current().Position,
	}
	if p.Current().Type != parser.TypeKeyword || p.Current().Value != "if" {
		return nil, p.NewError("If-statements have to start with 'if'", false, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	var err *parser.Error
	mlif.Condition, err = p.This.ParseExpression()
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("No expression found as if-condition"))
	}

	if p.Current().Type != parser.TypeKeyword || p.Current().Value != "then" {
		return nil, p.NewError("Expected 'then' after condition", true, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	if p.Current().Type != parser.TypeNewline {
		return nil, p.NewError("Expected newline after then", true, p.Current().Position, p.Current().Position)
	}
	p.Advance()
	// mulitline if

	mlif.IfBlock, err = p.ParseLinesUntil(func() bool {
		return p.Current().Type == parser.TypeKeyword && (p.Current().Value == "end" || p.Current().Value == "else")
	})
	if err != nil {
		return nil, err
	}

	if p.Current().Type == parser.TypeKeyword && p.Current().Value == "else" {
		p.Advance()
		mlif.ElseBlock, err = p.ParseLinesUntil(func() bool {
			return p.Current().Type == parser.TypeKeyword && p.Current().Value == "end"
		})
		if err != nil {
			return nil, err
		}
	}

	if p.Current().Type != parser.TypeKeyword || p.Current().Value != "end" {
		return nil, p.NewError("Expected 'end' after if statement", true, mlif.Start(), mlif.Start())
	}
	p.Advance()

	if p.Current().Type != parser.TypeNewline && p.Current().Type != parser.TypeEOF {
		return nil, p.NewError("End must be followed by newline or EOF", true, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	return &mlif, nil
}

// ParseWhile pasres a nolol while
func (p *Parser) ParseWhile() (Line, *parser.Error) {
	p.Log()
	loop := WhileLoop{
		Position: p.Current().Position,
	}
	if p.Current().Type != parser.TypeKeyword || p.Current().Value != "while" {
		return nil, p.NewError("While-statements have to start with 'while'", false, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	var err *parser.Error
	loop.Condition, err = p.This.ParseExpression()
	if err != nil {
		err.Fatal = true
		return nil, err.Append(fmt.Errorf("No expression found as loop-condition"))
	}

	if p.Current().Type != parser.TypeKeyword || p.Current().Value != "do" {
		return nil, p.NewError("Expected 'do' after condition", true, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	if p.Current().Type != parser.TypeNewline {
		return nil, p.NewError("Expected newline after do", true, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	loop.Block, err = p.ParseLinesUntil(func() bool {
		return p.Current().Type == parser.TypeKeyword && p.Current().Value == "end"
	})
	if err != nil {
		return nil, err
	}

	if p.Current().Type != parser.TypeKeyword || p.Current().Value != "end" {
		return nil, p.NewError("Expected 'end' after if statement", true, loop.Start(), loop.Start())
	}
	p.Advance()

	if p.Current().Type != parser.TypeNewline && p.Current().Type != parser.TypeEOF {
		return nil, p.NewError("End must be followed by newline or EOF", true, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	return &loop, nil

}

// ParseIf overrides and disables the old yolol-style inline ifs
func (p *Parser) ParseIf() (parser.Statement, *parser.Error) {
	p.Log()
	return nil, p.NewError("Inline if is not supported by nolol.", false, p.Current().Position, p.Current().Position)
}

// ParseGoto allows labeled-gotos and forbids line-based gotos
func (p *Parser) ParseGoto() (parser.Statement, *parser.Error) {
	if p.Current().Type == parser.TypeKeyword && p.Current().Value == "goto" {
		p.Advance()

		if p.Current().Type == parser.TypeID {
			stmt := &GoToLabelStatement{
				Position: p.Current().Position,
				Label:    p.Current().Value,
			}
			p.Advance()
			return stmt, nil
		}

		return nil, p.NewError("Goto must be followed by an identifier", true, p.Current().Position, p.Current().Position)
	}
	return nil, p.NewError("Goto statements must start with 'goto'", false, p.Current().Position, p.Current().Position)
}
