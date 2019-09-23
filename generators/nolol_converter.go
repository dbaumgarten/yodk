package generators

import (
	"fmt"

	"github.com/dbaumgarten/yodk/parser"
)

type NololConverter struct {
	jumpLabels        map[string]int
	constants         map[string]interface{}
	lineNumberChanges map[int]int
}

func NewNololConverter() *NololConverter {
	return &NololConverter{}
}
func (c *NololConverter) ConvertFromSource(prog string) (*parser.Programm, error) {
	p := parser.NewNololParser()
	parsed, err := p.Parse(prog)
	if err != nil {
		return nil, err
	}
	return c.Convert(parsed)
}

func (c *NololConverter) Convert(prog *parser.ExtProgramm) (*parser.Programm, error) {
	err := c.findConstantDeclarations(prog)
	if err != nil {
		return nil, err
	}
	err = c.insertConstants(prog)
	if err != nil {
		return nil, err
	}
	err = c.findJumpLabels(prog)
	if err != nil {
		return nil, err
	}
	err = c.convertLabelGoto(prog)
	if err != nil {
		return nil, err
	}
	newprog := c.convertProgramm(prog)

	return newprog, c.fixGotoLineNumbers(newprog)
}

func (c *NololConverter) findConstantDeclarations(p parser.Node) error {
	c.constants = make(map[string]interface{}, 0)
	f := func(node parser.Node, visitType int) error {
		if visitType == parser.PreVisit {
			if constDecl, is := node.(*parser.ConstDeclaration); is {
				_, exists := c.constants[constDecl.Name]
				if exists {
					return &parser.ParserError{
						Message:       fmt.Sprintf("Duplicate declaration of constant: %s", constDecl.Name),
						Fatal:         true,
						StartPosition: constDecl.Start(),
						EndPosition:   constDecl.End(),
					}
				}
				switch val := constDecl.Value.(type) {
				case *parser.StringConstant:
					c.constants[constDecl.Name] = val
					break
				case *parser.NumberConstant:
					c.constants[constDecl.Name] = val
					break
				default:
					return &parser.ParserError{
						Message:       "Only constant values can be the value of a constant declaration",
						Fatal:         true,
						StartPosition: constDecl.Start(),
						EndPosition:   constDecl.End(),
					}
				}
			}
		}
		return nil
	}
	return p.Accept(parser.VisitorFunc(f))
}

func (c *NololConverter) insertConstants(p parser.Node) error {
	f := func(node parser.Node, visitType int) error {
		if visitType == parser.SingleVisit {
			if deref, is := node.(*parser.Dereference); is {
				if deref.Operator == "" {
					if value, exists := c.constants[deref.Variable]; exists {
						var replacement parser.Expression
						switch val := value.(type) {
						case *parser.StringConstant:
							replacement = &parser.StringConstant{
								Value:    val.Value,
								Position: deref.Position,
							}
							break
						case *parser.NumberConstant:
							replacement = &parser.NumberConstant{
								Value:    val.Value,
								Position: deref.Position,
							}
							break
						}
						return parser.NewNodeReplacement(replacement)
					}
				}
			}
		}
		return nil
	}
	return p.Accept(parser.VisitorFunc(f))
}

func (c *NololConverter) findJumpLabels(p parser.Node) error {
	c.jumpLabels = make(map[string]int)
	f := func(node parser.Node, visitType int) error {
		if line, isExecutableLine := node.(*parser.ExecutableLine); isExecutableLine {
			if visitType == parser.PreVisit && line.Label != "" {
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
	return p.Accept(parser.VisitorFunc(f))
}

func (c *NololConverter) convertLabelGoto(p parser.Node) error {
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

func (c *NololConverter) fixGotoLineNumbers(p parser.Node) error {
	f := func(node parser.Node, visitType int) error {
		if gotostmt, is := node.(*parser.GoToStatement); is {
			newline, exists := c.lineNumberChanges[gotostmt.Line]
			if !exists {
				return parser.ParserError{
					Message:       fmt.Sprintf("Can not jump to line: %d", gotostmt.Line),
					Fatal:         true,
					StartPosition: gotostmt.Start(),
					EndPosition:   gotostmt.End(),
				}
			}
			gotostmt.Line = newline
		}
		return nil
	}
	return p.Accept(parser.VisitorFunc(f))
}

func (c *NololConverter) convertProgramm(p *parser.ExtProgramm) *parser.Programm {
	c.lineNumberChanges = make(map[int]int)
	newprog := parser.Programm{
		Lines: make([]*parser.Line, 0),
	}
	for _, rawline := range p.ExecutableLines {
		if line, isExecutableLine := rawline.(*parser.ExecutableLine); isExecutableLine {
			if len(line.Statements) == 0 {
				c.lineNumberChanges[line.Start().Line] = len(newprog.Lines) + 1
				continue
			}
			newline := &parser.Line{
				Statements: line.Statements,
			}
			newprog.Lines = append(newprog.Lines, newline)
			c.lineNumberChanges[line.Start().Line] = len(newprog.Lines)
		}
	}
	return &newprog
}
