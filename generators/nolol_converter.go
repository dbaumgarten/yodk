package generators

import (
	"fmt"

	"github.com/dbaumgarten/yodk/parser"
)

type NololConverter struct {
	jumpLabels map[string]int
}

func NewNololConverter() *NololConverter {
	return &NololConverter{}
}

func (c *NololConverter) Convert(prog *parser.ExtProgramm) (*parser.Programm, error) {
	err := c.findJumpLabels(prog)
	if err != nil {
		return nil, err
	}
	err = c.convertLabelGoto(prog)
	if err != nil {
		return nil, err
	}
	return c.convertProgramm(prog), nil
}

func (c *NololConverter) findJumpLabels(p *parser.ExtProgramm) error {
	c.jumpLabels = make(map[string]int)
	for _, line := range p.ExecutableLines {
		if line.Label != "" {
			_, exists := c.jumpLabels[line.Label]
			if exists {
				return &parser.ParserError{
					Message:       fmt.Sprintf("Duplicate declaration of jump-label: %s", line.Label),
					Fatal:         true,
					StartPosition: line.Start(),
					EndPosition:   line.End(),
				}
			}
			c.jumpLabels[line.Label] = line.Start().Line
		}
	}
	return nil
}

func (c *NololConverter) convertLabelGoto(p *parser.ExtProgramm) error {
	f := func(node parser.Node, visitType int) error {
		if gotostmt, is := node.(*parser.GoToLabelStatement); is {
			line, exists := c.jumpLabels[gotostmt.Label]
			if !exists {
				return parser.ParserError{
					Message:       "Unknown jump-label: " + gotostmt.Label,
					Fatal:         true,
					StartPosition: gotostmt.Start(),
					EndPosition:   gotostmt.End(),
				}
			}
			repl := &parser.GoToStatement{
				Position: gotostmt.Position,
				Line:     line,
			}
			return parser.NewNodeReplacement(repl)
		}
		return nil
	}
	return p.Accept(parser.VisitorFunc(f))
}

func (c *NololConverter) convertProgramm(p *parser.ExtProgramm) *parser.Programm {
	newprog := parser.Programm{
		Lines: make([]*parser.Line, 0),
	}
	for _, line := range p.ExecutableLines {
		newline := &parser.Line{
			Statements: line.Statements,
		}
		newprog.Lines = append(newprog.Lines, newline)
	}
	return &newprog
}
