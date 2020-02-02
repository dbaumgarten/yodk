package nolol

import (
	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Parser parses a nolol-program
type Parser struct {
	*parser.Parser
}

// NewParser creates and returns a nolol parser
func NewParser() *Parser {
	ep := &Parser{
		Parser: parser.NewParser(),
	}
	ep.Tokenizer = nast.NewNololTokenizer()
	ep.This = ep
	return ep
}

// Parse is the entry point for parsing
func (p *Parser) Parse(prog string) (*nast.Program, error) {
	p.Errors = make(parser.Errors, 0)
	p.Comments = make([]*ast.Token, 0)
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
func (p *Parser) ParseProgram() *nast.Program {
	p.Log()
	ret := nast.Program{
		Lines: make([]nast.Line, 0),
	}
	for p.HasNext() {
		ret.Lines = append(ret.Lines, p.ParseLine())
	}
	return &ret
}

// ParseStatementLine parses a statement line
func (p *Parser) ParseStatementLine() *nast.StatementLine {
	ret := nast.StatementLine{
		Line: ast.Line{
			Statements: make([]ast.Statement, 0, 1),
		},
		Position: p.CurrentToken.Position,
	}

	// get line-label if it exists
	if p.CurrentToken.Type == ast.TypeID && (p.NextToken.Type == ast.TypeSymbol && p.NextToken.Value == ">") {
		ret.Label = p.CurrentToken.Value
		p.Advance()
		p.Advance()
	}

	if p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == "$" {
		ret.HasBOL = true
		p.Advance()
	}

	// the line has no statements
	if p.CurrentToken.Type == ast.TypeEOF || p.CurrentToken.Type == ast.TypeNewline {
		p.Advance()
		// if a line has no statements, its BOL is also its EOL
		ret.HasEOL = ret.HasBOL
		return &ret
	}

	stmt := p.This.ParseStatement()
	// a line may have 0 statements and may still be usefull because of an EOL
	if stmt != nil {
		ret.Statements = append(ret.Statements, stmt)
	}

	for p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == ";" {
		p.Advance()
		stmt = p.This.ParseStatement()
		if stmt != nil {
			ret.Statements = append(ret.Statements, stmt)
		} else {
			p.ErrorCurrent(("Expected a statement after ';'"))
		}
	}

	if p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == "$" {
		ret.HasEOL = true
		p.Advance()
	}

	if p.CurrentToken.Type != ast.TypeEOF {
		p.Expect(ast.TypeNewline, "")
	}

	return &ret
}

// ParseBlockStatement parses a NOLOL block statement
func (p *Parser) ParseBlockStatement() *nast.BlockStatement {
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "block" {
		return nil
	}
	p.Advance()
	st := &nast.BlockStatement{
		Position: p.CurrentToken.Position,
	}

	st.Condition = p.This.ParseExpression()
	if st.Condition == nil {
		p.ErrorCurrent("Expected an expression after 'block'")
	}

	return st
}

// ParseExecutableLine parses an if, while or statement-line
func (p *Parser) ParseExecutableLine() nast.ExecutableLine {

	ifline := p.ParseMultilineIf()
	if ifline != nil {
		return ifline
	}

	whileline := p.ParseWhile()
	if whileline != nil {
		return whileline
	}

	block := p.ParseBlockStatement()
	if block != nil {
		return block
	}

	return p.ParseStatementLine()
}

// ParseLine parses any kind of line
func (p *Parser) ParseLine() nast.Line {
	p.Log()

	constDecl := p.ParseConstantDeclaration()
	if constDecl != nil {
		return constDecl
	}

	return p.ParseExecutableLine()
}

// ParseConstantDeclaration parses a constant declaration
func (p *Parser) ParseConstantDeclaration() *nast.ConstDeclaration {
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "const" {
		return nil
	}
	startpos := p.CurrentToken.Position
	p.Advance()
	if p.CurrentToken.Type != ast.TypeID {
		p.ErrorCurrent("const keyword must be followed by an identifier")
	}
	decl := &nast.ConstDeclaration{
		Name:     p.CurrentToken.Value,
		Position: startpos,
	}
	p.Advance()
	p.Expect(ast.TypeSymbol, "=")
	value := p.ParseSingleExpression()
	if value == nil {
		p.ErrorCurrent("The = of a const declaration must be followed by an expression")
	}
	decl.Value = value
	if p.CurrentToken.Type != ast.TypeNewline {
		p.Expect(ast.TypeNewline, "")
	}
	return decl
}

// ParseLinesUntil parse lines until stop() returns true
func (p *Parser) ParseLinesUntil(stop func() bool) []nast.ExecutableLine {
	lines := make([]nast.ExecutableLine, 0)
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
func (p *Parser) ParseMultilineIf() nast.Line {
	p.Log()
	mlif := nast.MultilineIf{
		Position: p.CurrentToken.Position,
	}
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "if" {
		return nil
	}
	p.Advance()

	mlif.Condition = p.This.ParseExpression()
	if mlif.Condition == nil {
		p.ErrorCurrent("No expression found as if-condition")
	}

	p.Expect(ast.TypeKeyword, "then")
	p.Expect(ast.TypeNewline, "")

	mlif.IfBlock = p.ParseLinesUntil(func() bool {
		return p.CurrentToken.Type == ast.TypeKeyword && (p.CurrentToken.Value == "end" || p.CurrentToken.Value == "else")
	})

	if p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "else" {
		p.Advance()
		mlif.ElseBlock = p.ParseLinesUntil(func() bool {
			return p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "end"
		})
	}

	p.Expect(ast.TypeKeyword, "end")

	if p.CurrentToken.Type != ast.TypeEOF {
		p.Expect(ast.TypeNewline, "")
	}

	return &mlif
}

// ParseWhile pasres a nolol while
func (p *Parser) ParseWhile() nast.Line {
	p.Log()
	loop := nast.WhileLoop{
		Position: p.CurrentToken.Position,
	}
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "while" {
		return nil
	}
	p.Advance()

	loop.Condition = p.This.ParseExpression()
	if loop.Condition == nil {
		p.ErrorCurrent("No expression found as loop-condition")
	}

	p.Expect(ast.TypeKeyword, "do")
	p.Expect(ast.TypeNewline, "")

	loop.Block = p.ParseLinesUntil(func() bool {
		return p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "end"
	})

	p.Expect(ast.TypeKeyword, "end")

	if p.CurrentToken.Type != ast.TypeEOF {
		p.Expect(ast.TypeNewline, "")
	}

	return &loop
}

// ParseIf overrides and disables the old yolol-style inline ifs
func (p *Parser) ParseIf() ast.Statement {
	p.Log()
	//Inline if is not supported by nolol. Always return nil
	return nil
}

// ParseGoto allows labeled-gotos and forbids line-based gotos
func (p *Parser) ParseGoto() ast.Statement {
	if p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "goto" {
		p.Advance()

		stmt := &nast.GoToLabelStatement{
			Position: p.CurrentToken.Position,
			Label:    p.CurrentToken.Value,
		}

		if p.CurrentToken.Type != ast.TypeID {
			p.ErrorCurrent("Goto must be followed by an identifier")
		} else {
			p.Advance()
		}

		return stmt
	}
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

	if p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == ")" {
		p.Advance()
		return fc
	}

	arg := p.This.ParseExpression()
	fc.Argument = arg
	if arg == nil {
		p.ErrorCurrent("Expected a function argument or ')'")
	}

	p.Expect(ast.TypeSymbol, ")")

	return fc
}
