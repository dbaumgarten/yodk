package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/optimizers"

	"github.com/dbaumgarten/yodk/parser"
)

// special error that is emitted if a nolol if can not be converted to an inline yolol-if
var errInlineIfImpossible = fmt.Errorf("Can not convert to inline if")

// Converter can convert a nolol-ast to a yolol-ast
type Converter struct {
	jumpLabels map[string]int
	constants  map[string]interface{}
}

// NewConverter creates a new converter
func NewConverter() *Converter {
	return &Converter{}
}

// ConvertFromSource is a shortcut that parses and directly convertes a nolol program
func (c *Converter) ConvertFromSource(prog string) (*parser.Program, error) {
	p := NewParser()
	parsed, err := p.Parse(prog)
	if err != nil {
		return nil, err
	}
	return c.Convert(parsed)
}

// Convert converts a nolol-program to a yolol-program
func (c *Converter) Convert(prog *Program) (*parser.Program, error) {
	// get all constant declarations
	err := c.findConstantDeclarations(prog)
	if err != nil {
		return nil, err
	}
	// fill all constants with the declared values
	err = c.insertConstants(prog)
	if err != nil {
		return nil, err
	}
	// optimize static expressions
	err = optimizers.NewStaticExpressionOptimizer().Optimize(prog)
	if err != nil {
		return nil, err
	}
	// remove useless lines
	err = c.filterLines(prog)
	if err != nil {
		return nil, err
	}
	// shorten variable names
	err = optimizers.NewVariableNameOptimizer().Optimize(prog)
	if err != nil {
		return nil, err
	}
	// convert nolol ifs to yolol code
	err = c.convertIf(prog)
	if err != nil {
		return nil, err
	}
	// convert while to yolol code
	err = c.convertWhileLoops(prog)
	if err != nil {
		return nil, err
	}
	// merge lines if possible
	newlines, err := c.mergeNololLines(prog.Lines)
	if err != nil {
		return nil, err
	}
	prog.Lines = newlines
	// find all line-labels
	err = c.findJumpLabels(prog)
	if err != nil {
		return nil, err
	}
	// resolve jump-labels
	err = c.convertLabelGoto(prog)
	if err != nil {
		return nil, err
	}
	// convert remaining nolol-types to yolol types
	newprog := c.convertLineTypes(prog)

	return newprog, nil
}

// findConstantDeclarations searches the programm for constant declarations and stores them for later use
func (c *Converter) findConstantDeclarations(p parser.Node) error {
	c.constants = make(map[string]interface{}, 0)
	f := func(node parser.Node, visitType int) error {
		if visitType == parser.PreVisit {
			if constDecl, is := node.(*ConstDeclaration); is {
				_, exists := c.constants[constDecl.Name]
				if exists {
					return &parser.Error{
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
					return &parser.Error{
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

// insertConstants fills in the values of defined constants
func (c *Converter) insertConstants(p parser.Node) error {
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

// findJumpLabels finds all line-labels in the program
func (c *Converter) findJumpLabels(p parser.Node) error {
	c.jumpLabels = make(map[string]int)
	linecounter := 0
	f := func(node parser.Node, visitType int) error {
		if line, isExecutableLine := node.(*StatementLine); isExecutableLine {
			if visitType == parser.PreVisit {
				linecounter++
				if line.Label != "" {
					_, exists := c.jumpLabels[line.Label]
					if exists {
						return &parser.Error{
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

// convertLabelGoto converts a nolol-style label-goto to a plain yolol-goto
func (c *Converter) convertLabelGoto(p parser.Node) error {
	f := func(node parser.Node, visitType int) error {
		if gotostmt, is := node.(*GoToLabelStatement); is {
			line, exists := c.jumpLabels[gotostmt.Label]
			if !exists {
				return &parser.Error{
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

// convertIf converts nolol's multiline-ifs to yolol ifs
func (c *Converter) convertIf(p parser.Node) error {
	counter := 0
	f := func(node parser.Node, visitType int) error {
		if mlif, is := node.(*MultilineIf); is && visitType == parser.PostVisit {
			// first, try to convert to inline if
			result := c.convertIfInline(mlif)
			if result != errInlineIfImpossible {
				return result
			}
			// inline if is not possible. Do a multiline if instead
			return c.convertIfMultiline(mlif, &counter)
		}
		return nil
	}
	return p.Accept(parser.VisitorFunc(f))
}

// convertIfInline converts a nolol-if directly to a yolol-if, if possible
func (c *Converter) convertIfInline(mlif *MultilineIf) error {
	linelength := len("if  then  end")
	mergedIfLines, _ := c.mergeExecutableLines(mlif.IfBlock)
	var mergedElseLines []ExecutableLine
	var elseBlock []parser.Statement

	if len(mergedIfLines) > 1 || mergedIfLines[0].(*StatementLine).Label != "" {
		return errInlineIfImpossible
	}
	linelength += getLengthOfLine(&mergedIfLines[0].(*StatementLine).Line)

	if mlif.ElseBlock != nil {
		mergedElseLines, _ = c.mergeExecutableLines(mlif.ElseBlock)
		if len(mergedElseLines) > 1 || mergedElseLines[0].(*StatementLine).Label != "" {
			return errInlineIfImpossible
		}
		linelength += len(" else ")
		linelength += getLengthOfLine(&mergedElseLines[0].(*StatementLine).Line)
		elseBlock = mergedElseLines[0].(*StatementLine).Line.Statements
	}

	if linelength > 70 {
		return errInlineIfImpossible
	}

	repl := &StatementLine{
		Position: mlif.Position,
		Line: parser.Line{
			Statements: []parser.Statement{
				&parser.IfStatement{
					Position:  mlif.Position,
					Condition: mlif.Condition,
					IfBlock:   mergedIfLines[0].(*StatementLine).Line.Statements,
					ElseBlock: elseBlock,
				},
			},
		},
	}

	return parser.NewNodeReplacement(repl)
}

// convertIfMultiline combines if, lables and gotos to implement multiline ifs with yolol
func (c *Converter) convertIfMultiline(mlif *MultilineIf, counter *int) error {
	skipIf := fmt.Sprintf("iflbl%d", *counter)
	skipElse := fmt.Sprintf("elselbl%d", *counter)
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

	*counter++
	return parser.NewNodeReplacement(repl...)
}

// convertWhileLoops converts while loops into yolol-code
func (c *Converter) convertWhileLoops(p parser.Node) error {
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

// filterLines removes empty lines and constant declarations from the program
func (c *Converter) filterLines(p parser.Node) error {
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

// mergeExecutableLines is a type-wrapper for mergeStatementLines
func (c *Converter) mergeExecutableLines(lines []ExecutableLine) ([]ExecutableLine, error) {
	inp := make([]*StatementLine, len(lines))
	for i, elem := range lines {
		inp[i] = elem.(*StatementLine)
	}
	interm, err := c.mergeStatementLines(inp)
	if err != nil {
		return nil, err
	}
	outp := make([]ExecutableLine, len(interm))
	for i, elem := range interm {
		outp[i] = elem
	}
	return outp, nil
}

// mergeNololLines is a type-wrapper for mergeStatementLines
func (c *Converter) mergeNololLines(lines []Line) ([]Line, error) {
	inp := make([]*StatementLine, len(lines))
	for i, elem := range lines {
		inp[i] = elem.(*StatementLine)
	}
	interm, err := c.mergeStatementLines(inp)
	if err != nil {
		return nil, err
	}
	outp := make([]Line, len(interm))
	for i, elem := range interm {
		outp[i] = elem
	}
	return outp, nil
}

// mergeStatementLines merges consectuive statementlines into as few lines as possible
func (c *Converter) mergeStatementLines(lines []*StatementLine) ([]*StatementLine, error) {
	maxlen := 70
	newLines := make([]*StatementLine, 0, len(lines))
	i := 0
	for i < len(lines) {
		current := &StatementLine{
			Line: parser.Line{
				Statements: []parser.Statement{},
			},
			Label:    lines[i].Label,
			Position: lines[i].Position,
		}
		current.Statements = append(current.Statements, lines[i].Statements...)
		newLines = append(newLines, current)
		for i+1 < len(lines) {
			currlen := getLengthOfLine(&current.Line)
			nextline := lines[i+1]
			nextlen := getLengthOfLine(&nextline.Line)

			if nextline.Label == "" && currlen+nextlen <= maxlen {
				current.Statements = append(current.Statements, nextline.Statements...)
				i++
			} else {
				break
			}
		}
		i++
	}
	return newLines, nil
}

// getLengthOfLine returns the amount of characters needed to represent the given line as yolol-code
func getLengthOfLine(line *parser.Line) int {
	ygen := parser.Printer{}
	ygen.UnknownHandlerFunc = func(node parser.Node, visitType int) (string, error) {
		if _, is := node.(*GoToLabelStatement); is {
			return "goto XX", nil
		}
		return "", fmt.Errorf("Unknown node-type: %T", node)
	}
	generated, err := ygen.Print(line)
	if err != nil {
		panic(err)
	}

	return len(generated)
}

// convertLineTypes converts nolol line-types to yolol-types
func (c *Converter) convertLineTypes(p *Program) *parser.Program {
	newprog := parser.Program{
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
