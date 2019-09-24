package nolol

import (
	"fmt"
	"strconv"

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

func (p *NololParser) Parse(prog string) (*ExtProgramm, error) {
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

func (p *NololParser) ParseProgram() (*ExtProgramm, parser.ParserErrors) {
	p.Log()
	errors := make(parser.ParserErrors, 0)
	ret := ExtProgramm{
		Lines: make([]ExtLine, 0),
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

func (p *NololParser) ParseLine() (ExtLine, *parser.ParserError) {
	p.Log()

	constDecl, err := p.ParseConstantDeclaration()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return constDecl, nil
	}

	ret := ExecutableLine{
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

func (p *NololParser) ParseGoto() (parser.Statement, *parser.ParserError) {
	if p.Current().Type == parser.TypeKeyword && p.Current().Value == "goto" {
		p.Advance()
		if p.Current().Type == parser.TypeNumber {
			stmt := &parser.GoToStatement{
				Position: p.Current().Position,
			}
			line, _ := strconv.Atoi(p.Current().Value)
			stmt.Line = line
			p.Advance()
			return stmt, nil
		}

		if p.Current().Type == parser.TypeID {
			stmt := &GoToLabelStatement{
				Position: p.Current().Position,
				Label:    p.Current().Value,
			}
			p.Advance()
			return stmt, nil
		}

		return nil, p.NewError("Goto must be followed by number or identifier", true, p.Current().Position, p.Current().Position)
	}
	return nil, p.NewError("Goto statements must start with 'goto'", false, p.Current().Position, p.Current().Position)
}
