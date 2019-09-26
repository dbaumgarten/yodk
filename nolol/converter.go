package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/optimizers"

	"github.com/dbaumgarten/yodk/parser"
)

type NololConverter struct {
	jumpLabels map[string]int
	constants  map[string]interface{}
}

func NewNololConverter() *NololConverter {
	return &NololConverter{}
}
func (c *NololConverter) ConvertFromSource(prog string) (*parser.Programm, error) {
	p := NewNololParser()
	parsed, err := p.Parse(prog)
	if err != nil {
		return nil, err
	}
	return c.Convert(parsed)
}

func (c *NololConverter) Convert(prog *NololProgramm) (*parser.Programm, error) {
	err := c.convertMultilineIf(prog)
	if err != nil {
		return nil, err
	}
	err = c.convertWhileLoops(prog)
	if err != nil {
		return nil, err
	}
	err = optimizers.NewStaticExpressionOptimizer().Optimize(prog)
	if err != nil {
		return nil, err
	}
	err = c.findConstantDeclarations(prog)
	if err != nil {
		return nil, err
	}
	err = c.insertConstants(prog)
	if err != nil {
		return nil, err
	}
	err = c.filterLines(prog)
	if err != nil {
		return nil, err
	}
	err = optimizers.NewVariableNameOptimizer().Optimize(prog)
	if err != nil {
		return nil, err
	}
	err = c.mergeLines(prog)
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
	newprog := c.convertToYololLines(prog)

	return newprog, nil
}

func (c *NololConverter) findConstantDeclarations(p parser.Node) error {
	c.constants = make(map[string]interface{}, 0)
	f := func(node parser.Node, visitType int) error {
		if visitType == parser.PreVisit {
			if constDecl, is := node.(*ConstDeclaration); is {
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
	linecounter := 0
	f := func(node parser.Node, visitType int) error {
		if line, isExecutableLine := node.(*StatementLine); isExecutableLine {
			if visitType == parser.PreVisit {
				linecounter++
				if line.Label != "" {
					_, exists := c.jumpLabels[line.Label]
					if exists {
						return &parser.ParserError{
							Message:       fmt.Sprintf("Duplicate declaration of jump-label: %s", line.Label),
							Fatal:         true,
							StartPosition: line.Start(),
							EndPosition:   line.Start(),
						}
					}
					c.jumpLabels[line.Label] = linecounter

					if len(line.Statements) == 0 {
						linecounter--
						return parser.NewNodeReplacement()
					}
				}
			}
		}
		return nil
	}
	return p.Accept(parser.VisitorFunc(f))
}

func (c *NololConverter) convertLabelGoto(p parser.Node) error {
	f := func(node parser.Node, visitType int) error {
		if gotostmt, is := node.(*GoToLabelStatement); is {
			line, exists := c.jumpLabels[gotostmt.Label]
			if !exists {
				return &parser.ParserError{
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

func (c *NololConverter) convertMultilineIf(p parser.Node) error {
	counter := 0
	f := func(node parser.Node, visitType int) error {
		if mlif, is := node.(*MultilineIf); is && visitType == parser.PostVisit {
			skipIf := fmt.Sprintf("iflbl%d", counter)
			skipElse := fmt.Sprintf("elselbl%d", counter)
			repl := []parser.Node{
				&StatementLine{
					Position: mlif.Position,
					Line: parser.Line{
						Statements: []parser.Statement{
							&parser.IfStatement{
								Position: mlif.Position,
								Condition: &parser.UnaryOperation{
									Operator: "not",
									Exp:      mlif.Condition,
								},
								IfBlock: []parser.Statement{
									&GoToLabelStatement{
										Position: mlif.Position,
										Label:    skipIf,
									},
								},
							},
						},
					},
				},
			}

			for _, ifblling := range mlif.IfBlock {
				repl = append(repl, ifblling)
			}

			if mlif.ElseBlock != nil {
				repl = append(repl, &StatementLine{
					Position: mlif.Position,
					Line: parser.Line{
						Statements: []parser.Statement{
							&GoToLabelStatement{
								Position: mlif.Position,
								Label:    skipElse,
							},
						},
					},
				})
			}

			repl = append(repl, &StatementLine{
				Position: mlif.Position,
				Label:    skipIf,
				Line: parser.Line{
					Statements: []parser.Statement{},
				},
			})

			if mlif.ElseBlock != nil {
				for _, ifblling := range mlif.ElseBlock {
					repl = append(repl, ifblling)
				}
			}

			repl = append(repl, &StatementLine{
				Position: mlif.Position,
				Label:    skipElse,
				Line: parser.Line{
					Statements: []parser.Statement{},
				},
			})

			counter++
			return parser.NewNodeReplacement(repl...)
		}
		return nil
	}
	return p.Accept(parser.VisitorFunc(f))
}

func (c *NololConverter) convertWhileLoops(p parser.Node) error {
	counter := 0
	f := func(node parser.Node, visitType int) error {
		if loop, is := node.(*WhileLoop); is && visitType == parser.PostVisit {
			startLabel := fmt.Sprintf("while%d", counter)
			endLabel := fmt.Sprintf("endwhile%d", counter)
			repl := []parser.Node{
				&StatementLine{
					Position: loop.Position,
					Label:    startLabel,
					Line: parser.Line{
						Statements: []parser.Statement{
							&parser.IfStatement{
								Position: loop.Condition.Start(),
								Condition: &parser.UnaryOperation{
									Operator: "not",
									Exp:      loop.Condition,
								},
								IfBlock: []parser.Statement{
									&GoToLabelStatement{
										Position: loop.Position,
										Label:    endLabel,
									},
								},
							},
						},
					},
				},
			}

			for _, blockline := range loop.Block {
				repl = append(repl, blockline)
			}
			repl = append(repl, &StatementLine{
				Position: loop.Position,
				Line: parser.Line{
					Statements: []parser.Statement{
						&GoToLabelStatement{
							Position: loop.Position,
							Label:    startLabel,
						},
					},
				},
			})

			repl = append(repl, &StatementLine{
				Position: loop.Position,
				Label:    endLabel,
				Line: parser.Line{
					Statements: []parser.Statement{},
				},
			})

			counter++
			return parser.NewNodeReplacement(repl...)
		}
		return nil
	}
	return p.Accept(parser.VisitorFunc(f))
}

func (c *NololConverter) filterLines(p parser.Node) error {
	f := func(node parser.Node, visitType int) error {
		switch n := node.(type) {
		case *StatementLine:
			if n.Label == "" && len(n.Statements) == 0 {
				// empty line
				return parser.NewNodeReplacement()
			}
		case *ConstDeclaration:
			return parser.NewNodeReplacement()
		}
		return nil
	}
	return p.Accept(parser.VisitorFunc(f))
}

func (c *NololConverter) mergeLines(p *NololProgramm) error {
	maxlen := 70
	newLines := make([]NololLine, 0, len(p.Lines))
	i := 0
	for i < len(p.Lines)-1 {
		if current, is := p.Lines[i].(*StatementLine); is {
			nextcounter := 1
			for i+nextcounter < len(p.Lines) {
				currlen := getLengthOfLine(&current.Line)
				nextline := p.Lines[i+nextcounter].(*StatementLine)
				nextlen := getLengthOfLine(&nextline.Line)
				if nextline.Label != "" || currlen+nextlen > maxlen {
					newLines = append(newLines, current)
					current = nextline
					i += nextcounter
					nextcounter = 1
					if i+nextcounter == len(p.Lines) {
						//the next line has not been merged, but it is the last line in the programm
						//append it to the list so it is not left out
						newLines = append(newLines, current)
					}
					continue
				}
				current.Statements = append(current.Statements, nextline.Statements...)
				nextcounter++
			}
		} else {
			panic("mergeLines can only work with Statement-lines")
		}
	}
	p.Lines = newLines
	return nil
}

func getLengthOfLine(line *parser.Line) int {
	ygen := parser.YololGenerator{}
	ygen.UnknownHandlerFunc = func(node parser.Node) (string, error) {
		if _, is := node.(*GoToLabelStatement); is {
			return "goto XX", nil
		}
		return "", fmt.Errorf("Unknown node-type: %T", node)
	}
	generated, err := ygen.Generate(line)
	if err != nil{
		panic(err)
	}

	return len(generated)
}

func (c *NololConverter) convertToYololLines(p *NololProgramm) *parser.Programm {
	newprog := parser.Programm{
		Lines: make([]*parser.Line, 0),
	}
	for _, rawline := range p.Lines {
		switch line := rawline.(type) {
		case *StatementLine:
			newline := &parser.Line{
				Statements: line.Statements,
			}
			newprog.Lines = append(newprog.Lines, newline)
			break
		default:
			fmt.Printf("Unexpected type: %T\n", rawline)
		}

	}
	return &newprog
}
