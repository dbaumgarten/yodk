package parser

import (
	"fmt"
	"strconv"
)

type NololParser struct {
	*Parser
}

func NewNololParser() *NololParser {
	ep := &NololParser{
		Parser: NewParser(),
	}
	ep.tokenizer = NewNololTokenizer()
	ep.this = ep
	return ep
}

func (p *NololParser) Parse(prog string) (*ExtProgramm, error) {
	errors := make(ParserErrors, 0)
	p.tokenizer.Load(prog)
	p.tokens = make([]*Token, 0, 1000)
	for {
		token, err := p.tokenizer.Next()
		if err != nil {
			errors = append(errors, err.(*ParserError))
		} else {
			if p.DebugLog {
				fmt.Print(token)
			}
			p.tokens = append(p.tokens, token)
			if token.Type == TypeEOF {
				break
			}
		}
	}
	p.currentToken = 0
	parsed, err := p.parseProgram()
	errors = append(errors, err...)
	if len(errors) > 0 {
		return nil, errors
	}
	return parsed, nil
}

func (p *NololParser) parseProgram() (*ExtProgramm, ParserErrors) {
	p.log()
	errors := make(ParserErrors, 0)
	ret := ExtProgramm{
		ExecutableLines: make([]ExtLine, 0),
	}
	for p.hasNext() {
		line, err := p.parseLine()
		if err != nil {
			errors = append(errors, err)
			p.skipLine()
		}
		ret.ExecutableLines = append(ret.ExecutableLines, line)
	}
	return &ret, errors
}

func (p *NololParser) parseLine() (ExtLine, *ParserError) {
	p.log()

	constDecl, err := p.parseConstantDeclaration()
	if err != nil && err.Fatal {
		return nil, err
	}
	if err == nil {
		return constDecl, nil
	}

	ret := ExecutableLine{
		Line: Line{
			Statements: make([]Statement, 0),
		},
	}

	// get line-label if it exists
	if p.current().Type == TypeID && (p.next().Type == TypeSymbol && p.next().Value == ">") {
		ret.Label = p.current().Value
		p.advance()
		p.advance()
	}

	for p.hasNext() {
		if p.current().Type == TypeNewline || p.current().Type == TypeEOF {
			p.advance()
			return &ret, nil
		}
		stmt, err := p.this.parseStatement()
		if err != nil {
			return nil, err
		}
		ret.Statements = append(ret.Statements, stmt)
	}

	if p.current().Type == TypeEOF {
		return &ret, nil
	}

	return nil, p.newError("Missing newline", true, ret.Start(), ret.End())
}

func (p *NololParser) parseConstantDeclaration() (*ConstDeclaration, *ParserError) {
	if p.current().Type != TypeKeyword || p.current().Value != "const" {
		return nil, p.newError("Const declaration must start with const", false, p.current().Position, p.current().Position)
	}
	startpos := p.current().Position
	p.advance()
	if p.current().Type != TypeID {
		return nil, p.newError("const keyword must be followed by identifier", true, p.current().Position, p.current().Position)
	}
	decl := &ConstDeclaration{
		Name:     p.current().Value,
		Position: startpos,
	}
	p.advance()
	if p.current().Type != TypeSymbol || p.current().Value != "=" {
		return nil, p.newError("Identifier in const declaration must be followed by '='", true, p.current().Position, p.current().Position)
	}
	p.advance()
	value, err := p.parseSingleExpression()
	if err != nil {
		err.Fatal = true
		return nil, err
	}
	decl.Value = value
	if p.current().Type != TypeNewline && p.current().Type != TypeEOF {
		return nil, p.newError("Const declaration must be followed by newline or EOF", true, p.current().Position, p.current().Position)
	}
	p.advance()
	return decl, nil
}

func (p *NololParser) parseGoto() (Statement, *ParserError) {
	if p.current().Type == TypeKeyword && p.current().Value == "goto" {
		p.advance()
		if p.current().Type == TypeNumber {
			stmt := &GoToStatement{
				Position: p.current().Position,
			}
			line, _ := strconv.Atoi(p.current().Value)
			stmt.Line = line
			p.advance()
			return stmt, nil
		}

		if p.current().Type == TypeID {
			stmt := &GoToLabelStatement{
				Position: p.current().Position,
				Label:    p.current().Value,
			}
			p.advance()
			return stmt, nil
		}

		return nil, p.newError("Goto must be followed by number or identifier", true, p.current().Position, p.current().Position)
	}
	return nil, p.newError("Goto statements must start with 'goto'", false, p.current().Position, p.current().Position)
}
