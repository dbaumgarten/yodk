package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/parser"
)

type NololParser struct {
	*parser.Parser
}

func NewNololParser() *NololParser {
	ep := &NololParser{
		Parser: parser.NewParser(),
	}
	ep.Tokenizer = NewNololTokenizer()
	ep.This = ep
	return ep
}

func (p *NololParser) Parse(prog string) (*NololProgramm, error) {
	errors := make(parser.ParserErrors, 0)
	p.Tokenizer.Load(prog)
	p.Tokens = make([]*parser.Token, 0, 1000)
	for {
		token, err := p.Tokenizer.Next()
		if err != nil {
			errors = append(errors, err.(*parser.ParserError))
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

func (p *NololParser) ParseProgram() (*NololProgramm, parser.ParserErrors) {
	p.Log()
	errors := make(parser.ParserErrors, 0)
	ret := NololProgramm{
		Lines: make([]NololLine, 0),
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

func (p *NololParser) ParseStatementLine() (*StatementLine, *parser.ParserError) {
	ret := StatementLine{
		Line: parser.Line{
			Statements: make([]parser.Statement, 0),
		},
		Position: p.Current().Position,
	}

	// get line-label if it exists
	if p.Current().Type == parser.TypeID && (p.Next().Type == parser.TypeSymbol && p.Next().Value == ">") {
		ret.Label = p.Current().Value
		p.Advance()
		p.Advance()
	}

	for p.HasNext() {
		if p.Current().Type == parser.TypeNewline || p.Current().Type == parser.TypeEOF {
			p.Advance()
			return &ret, nil
		}
		stmt, err := p.This.ParseStatement()
		if err != nil {
			return nil, err
		}
		ret.Statements = append(ret.Statements, stmt)
	}

	if p.Current().Type == parser.TypeEOF {
		return &ret, nil
	}

	return nil, p.NewError("Missing newline", true, ret.Start(), ret.End())
}

func (p *NololParser) ParseExecutableLine() (ExecutableLine, *parser.ParserError) {

	ifline, err := p.ParseIfLine()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return ifline, nil
	}

	whileline, err := p.ParseWhileLine()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return whileline, nil
	}

	return p.ParseStatementLine()
}

func (p *NololParser) ParseLine() (NololLine, *parser.ParserError) {
	p.Log()

	constDecl, err := p.ParseConstantDeclaration()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return constDecl, nil
	}

	return p.ParseExecutableLine()
}

func (p *NololParser) ParseConstantDeclaration() (*ConstDeclaration, *parser.ParserError) {
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

func (p *NololParser) ParseLinesUntil(stop func() bool) ([]ExecutableLine, *parser.ParserError) {
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

func (p *NololParser) ParseIfLine() (NololLine, *parser.ParserError) {
	p.Log()
	mlif := MultilineIf{
		Position: p.Current().Position,
	}
	if p.Current().Type != parser.TypeKeyword || p.Current().Value != "if" {
		return nil, p.NewError("If-statements have to start with 'if'", false, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	var err *parser.ParserError
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
	return &mlif, nil
}

func (p *NololParser) ParseWhileLine() (NololLine, *parser.ParserError) {
	p.Log()
	loop := WhileLoop{
		Position: p.Current().Position,
	}
	if p.Current().Type != parser.TypeKeyword || p.Current().Value != "while" {
		return nil, p.NewError("While-statements have to start with 'while'", false, p.Current().Position, p.Current().Position)
	}
	p.Advance()

	var err *parser.ParserError
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
	return &loop, nil

}

func (p *NololParser) ParseIf() (parser.Statement, *parser.ParserError) {
	p.Log()
	return nil, p.NewError("Inline if is not supported.", false, p.Current().Position, p.Current().Position)
}

func (p *NololParser) ParseGoto() (parser.Statement, *parser.ParserError) {
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
