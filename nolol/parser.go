package nolol

import (
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
	p.Errors = make(parser.Errors, 0)
	p.Comments = make([]*parser.Token, 0)
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

// ParseProgram parses the program
func (p *Parser) ParseProgram() *Program {
	p.Log()
	ret := Program{
		Lines: make([]Line, 0),
	}
	for p.HasNext() {
		ret.Lines = append(ret.Lines, p.ParseLine())
	}
	return &ret
}

// ParseStatementLine parses a statement line
func (p *Parser) ParseStatementLine() *StatementLine {
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
		return &ret
	}

	stmt := p.This.ParseStatement()
	ret.Statements = append(ret.Statements, stmt)

	if p.CurrentToken.Type != parser.TypeEOF {
		p.Expect(parser.TypeNewline, "")
	}

	return &ret
}

// ParseExecutableLine parses an if, while or statement-line
func (p *Parser) ParseExecutableLine() ExecutableLine {

	ifline := p.ParseMultilineIf()
	if ifline != nil {
		return ifline
	}

	whileline := p.ParseWhile()
	if whileline != nil {
		return whileline
	}

	return p.ParseStatementLine()
}

// ParseLine parses any kind of line
func (p *Parser) ParseLine() Line {
	p.Log()

	constDecl := p.ParseConstantDeclaration()
	if constDecl != nil {
		return constDecl
	}

	return p.ParseExecutableLine()
}

// ParseConstantDeclaration parses a constant declaration
func (p *Parser) ParseConstantDeclaration() *ConstDeclaration {
	if p.CurrentToken.Type != parser.TypeKeyword || p.CurrentToken.Value != "const" {
		return nil
	}
	startpos := p.CurrentToken.Position
	p.Advance()
	if p.CurrentToken.Type != parser.TypeID {
		p.ErrorCurrent("const keyword must be followed by an identifier")
	}
	decl := &ConstDeclaration{
		Name:     p.CurrentToken.Value,
		Position: startpos,
	}
	p.Advance()
	p.Expect(parser.TypeSymbol, "=")
	value := p.ParseSingleExpression()
	if value == nil {
		p.ErrorCurrent("The = of a const declaration must be followed by an expression")
	}
	decl.Value = value
	if p.CurrentToken.Type != parser.TypeNewline {
		p.Expect(parser.TypeNewline, "")
	}
	return decl
}

// ParseLinesUntil parse lines until stop() returns true
func (p *Parser) ParseLinesUntil(stop func() bool) []ExecutableLine {
	lines := make([]ExecutableLine, 0)
	for p.HasNext() && !stop() {
		line := p.ParseExecutableLine()
		if line == nil {
			break
		}
		lines = append(lines, line)
	}
	return lines
}

// ParseMultilineIf parses a nolol-style multiline if
func (p *Parser) ParseMultilineIf() Line {
	p.Log()
	mlif := MultilineIf{
		Position: p.CurrentToken.Position,
	}
	if p.CurrentToken.Type != parser.TypeKeyword || p.CurrentToken.Value != "if" {
		return nil
	}
	p.Advance()

	mlif.Condition = p.This.ParseExpression()
	if mlif.Condition == nil {
		p.ErrorCurrent("No expression found as if-condition")
	}

	p.Expect(parser.TypeKeyword, "then")
	p.Expect(parser.TypeNewline, "")

	mlif.IfBlock = p.ParseLinesUntil(func() bool {
		return p.CurrentToken.Type == parser.TypeKeyword && (p.CurrentToken.Value == "end" || p.CurrentToken.Value == "else")
	})

	if p.CurrentToken.Type == parser.TypeKeyword && p.CurrentToken.Value == "else" {
		p.Advance()
		mlif.ElseBlock = p.ParseLinesUntil(func() bool {
			return p.CurrentToken.Type == parser.TypeKeyword && p.CurrentToken.Value == "end"
		})
	}

	p.Expect(parser.TypeKeyword, "end")

	if p.CurrentToken.Type != parser.TypeEOF {
		p.Expect(parser.TypeNewline, "")
	}

	return &mlif
}

// ParseWhile pasres a nolol while
func (p *Parser) ParseWhile() Line {
	p.Log()
	loop := WhileLoop{
		Position: p.CurrentToken.Position,
	}
	if p.CurrentToken.Type != parser.TypeKeyword || p.CurrentToken.Value != "while" {
		return nil
	}
	p.Advance()

	loop.Condition = p.This.ParseExpression()
	if loop.Condition == nil {
		p.ErrorCurrent("No expression found as loop-condition")
	}

	p.Expect(parser.TypeKeyword, "do")
	p.Expect(parser.TypeNewline, "")

	loop.Block = p.ParseLinesUntil(func() bool {
		return p.CurrentToken.Type == parser.TypeKeyword && p.CurrentToken.Value == "end"
	})

	p.Expect(parser.TypeKeyword, "end")

	if p.CurrentToken.Type != parser.TypeEOF {
		p.Expect(parser.TypeNewline, "")
	}

	return &loop
}

// ParseIf overrides and disables the old yolol-style inline ifs
func (p *Parser) ParseIf() parser.Statement {
	p.Log()
	//Inline if is not supported by nolol. Always return nil
	return nil
}

// ParseGoto allows labeled-gotos and forbids line-based gotos
func (p *Parser) ParseGoto() parser.Statement {
	if p.CurrentToken.Type == parser.TypeKeyword && p.CurrentToken.Value == "goto" {
		p.Advance()

		stmt := &GoToLabelStatement{
			Position: p.CurrentToken.Position,
			Label:    p.CurrentToken.Value,
		}

		if p.CurrentToken.Type != parser.TypeID {
			p.ErrorCurrent("Goto must be followed by an identifier")
		} else {
			p.Advance()
		}

		return stmt
	}
	return nil
}
