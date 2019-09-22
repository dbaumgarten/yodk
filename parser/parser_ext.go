package parser

import (
	"fmt"
	"strconv"
)

type ExtParser struct {
	*Parser
}

func NewExtParser() *ExtParser {
	ep := &ExtParser{
		Parser: NewParser(),
	}
	ep.this = ep
	return ep
}

func (p *ExtParser) Parse(prog string) (*ExtProgramm, error) {
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

func (p *ExtParser) parseProgram() (*ExtProgramm, ParserErrors) {
	p.log()
	errors := make(ParserErrors, 0)
	ret := ExtProgramm{
		ExecutableLines: make([]*ExtLine, 0),
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

func (p *ExtParser) parseLine() (*ExtLine, *ParserError) {
	p.log()
	ret := ExtLine{
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

func (p *ExtParser) parseGoto() (Statement, *ParserError) {
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
