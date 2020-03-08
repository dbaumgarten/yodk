package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/optimizers"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// special error that is emitted if a nolol if can not be converted to an inline yolol-if
var errInlineIfImpossible = fmt.Errorf("Can not convert to inline if")

// reservedTimeVariable is the variable used to track passed time
var reservedTimeVariable = "_time"

// reservedTimeVariableReplaced is the name of reservedTimeVariable after variable name optimization
var reservedTimeVariableReplaced = "t"

// Converter can convert a nolol-ast to a yolol-ast
type Converter struct {
	jumpLabels       map[string]int
	constants        map[string]interface{}
	usesTimeTracking bool
}

// NewConverter creates a new converter
func NewConverter() *Converter {
	return &Converter{}
}

// ConvertFromSource is a shortcut that parses and directly convertes a nolol program
func (c *Converter) ConvertFromSource(prog string) (*ast.Program, error) {
	p := NewParser()
	parsed, err := p.Parse(prog)
	if err != nil {
		return nil, err
	}
	return c.Convert(parsed)
}

// Convert converts a nolol-program to a yolol-program
func (c *Converter) Convert(prog *nast.Program) (*ast.Program, error) {
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
	// replace builtin functions with their yolol code
	err = c.insertBuiltinFunctions(prog)
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
	opt := optimizers.NewVariableNameOptimizer()
	// take special care that the special variables are replaced correctly
	opt.SpecialReplacement(reservedTimeVariable, reservedTimeVariableReplaced)
	err = opt.Optimize(prog)
	if err != nil {
		return nil, err
	}
	// convert block statements to yolol code
	err = c.convertBlockStatement(prog)
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

	if c.usesTimeTracking {
		// inset line counter at the beginning of each line
		c.insertLineCounter(prog)
	}

	// convert remaining nolol-types to yolol types
	newprog := c.convertLineTypes(prog)

	return newprog, nil
}

// findConstantDeclarations searches the programm for constant declarations and stores them for later use
func (c *Converter) findConstantDeclarations(p ast.Node) error {
	c.constants = make(map[string]interface{}, 0)
	f := func(node ast.Node, visitType int) error {
		if visitType == ast.PreVisit {
			if constDecl, is := node.(*nast.ConstDeclaration); is {
				_, exists := c.constants[constDecl.Name]
				if exists {
					return &parser.Error{
						Message:       fmt.Sprintf("Duplicate declaration of constant: %s", constDecl.Name),
						StartPosition: constDecl.Start(),
						EndPosition:   constDecl.End(),
					}
				}
				switch val := constDecl.Value.(type) {
				case *ast.StringConstant:
					c.constants[constDecl.Name] = val
					break
				case *ast.NumberConstant:
					c.constants[constDecl.Name] = val
					break
				default:
					return &parser.Error{
						Message:       "Only constant values can be the value of a constant declaration",
						StartPosition: constDecl.Start(),
						EndPosition:   constDecl.End(),
					}
				}
			}
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

func (c *Converter) convertBlockStatement(p ast.Node) error {
	counter := 0
	f := func(node ast.Node, visitType int) error {
		if block, is := node.(*nast.BlockStatement); is {
			label := fmt.Sprintf("block%d", counter)
			return ast.NewNodeReplacement(&nast.StatementLine{
				Label:  label,
				HasEOL: true,
				Line: ast.Line{
					Position: node.Start(),
					Statements: []ast.Statement{
						&ast.IfStatement{
							Position:  node.Start(),
							Condition: block.Condition,
							IfBlock: []ast.Statement{
								&nast.GoToLabelStatement{
									Label: label,
								},
							},
						},
					},
				},
			})
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

func (c *Converter) insertBuiltinFunctions(p ast.Node) error {

	f := func(node ast.Node, visitType int) error {
		if function, isFunction := node.(*ast.FuncCall); isFunction {
			switch function.Function {
			case "time":
				c.usesTimeTracking = true
				return ast.NewNodeReplacement(&ast.Dereference{
					Variable: reservedTimeVariable,
				})
			}
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

// insertConstants fills in the values of defined constants
func (c *Converter) insertConstants(p ast.Node) error {
	f := func(node ast.Node, visitType int) error {
		if visitType == ast.SingleVisit {
			if deref, is := node.(*ast.Dereference); is {
				if deref.Operator == "" {
					if value, exists := c.constants[deref.Variable]; exists {
						var replacement ast.Expression
						switch val := value.(type) {
						case *ast.StringConstant:
							replacement = &ast.StringConstant{
								Value:    val.Value,
								Position: deref.Position,
							}
							break
						case *ast.NumberConstant:
							replacement = &ast.NumberConstant{
								Value:    val.Value,
								Position: deref.Position,
							}
							break
						}
						return ast.NewNodeReplacement(replacement)
					}
				}
			}
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

// findJumpLabels finds all line-labels in the program
func (c *Converter) findJumpLabels(p ast.Node) error {
	c.jumpLabels = make(map[string]int)
	linecounter := 0
	f := func(node ast.Node, visitType int) error {
		if line, isExecutableLine := node.(*nast.StatementLine); isExecutableLine {
			if visitType == ast.PreVisit {
				linecounter++
				if line.Label != "" {
					_, exists := c.jumpLabels[line.Label]
					if exists {
						return &parser.Error{
							Message:       fmt.Sprintf("Duplicate declaration of jump-label: %s", line.Label),
							StartPosition: line.Start(),
							EndPosition:   line.Start(),
						}
					}
					c.jumpLabels[line.Label] = linecounter

					if len(line.Statements) == 0 {
						linecounter--
						return ast.NewNodeReplacement()
					}
				}
			}
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

// convertLabelGoto converts a nolol-style label-goto to a plain yolol-goto
func (c *Converter) convertLabelGoto(p ast.Node) error {
	f := func(node ast.Node, visitType int) error {
		if gotostmt, is := node.(*nast.GoToLabelStatement); is {
			line, exists := c.jumpLabels[gotostmt.Label]
			if !exists {
				return &parser.Error{
					Message:       "Unknown jump-label: " + gotostmt.Label,
					StartPosition: gotostmt.Start(),
					EndPosition:   gotostmt.End(),
				}
			}
			repl := &ast.GoToStatement{
				Position: gotostmt.Position,
				Line:     line,
			}
			return ast.NewNodeReplacement(repl)
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

// convertIf converts nolol's multiline-ifs to yolol ifs
func (c *Converter) convertIf(p ast.Node) error {
	counter := 0
	f := func(node ast.Node, visitType int) error {
		if mlif, is := node.(*nast.MultilineIf); is && visitType == ast.PostVisit {
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
	return p.Accept(ast.VisitorFunc(f))
}

// convertIfInline converts a nolol-if directly to a yolol-if, if possible
func (c *Converter) convertIfInline(mlif *nast.MultilineIf) error {
	linelength := len("if  then  end")
	mergedIfLines, _ := c.mergeExecutableLines(mlif.IfBlock)
	var mergedElseLines []nast.ExecutableLine
	var elseBlock []ast.Statement

	if len(mergedIfLines) > 1 || mergedIfLines[0].(*nast.StatementLine).Label != "" {
		return errInlineIfImpossible
	}
	linelength += getLengthOfLine(&mergedIfLines[0].(*nast.StatementLine).Line)
	linelength += getLengthOfLine(mlif.Condition)

	if mlif.ElseBlock != nil {
		mergedElseLines, _ = c.mergeExecutableLines(mlif.ElseBlock)
		if len(mergedElseLines) > 1 || mergedElseLines[0].(*nast.StatementLine).Label != "" {
			return errInlineIfImpossible
		}
		linelength += len(" else ")
		linelength += getLengthOfLine(&mergedElseLines[0].(*nast.StatementLine).Line)
		elseBlock = mergedElseLines[0].(*nast.StatementLine).Line.Statements
	}

	if linelength > 70 {
		return errInlineIfImpossible
	}

	repl := &nast.StatementLine{
		Position: mlif.Position,
		Line: ast.Line{
			Statements: []ast.Statement{
				&ast.IfStatement{
					Position:  mlif.Position,
					Condition: mlif.Condition,
					IfBlock:   mergedIfLines[0].(*nast.StatementLine).Line.Statements,
					ElseBlock: elseBlock,
				},
			},
		},
	}

	return ast.NewNodeReplacement(repl)
}

// convertIfMultiline combines if, lables and gotos to implement multiline ifs with yolol
func (c *Converter) convertIfMultiline(mlif *nast.MultilineIf, counter *int) error {
	skipIf := fmt.Sprintf("iflbl%d", *counter)
	skipElse := fmt.Sprintf("elselbl%d", *counter)
	repl := []ast.Node{
		&nast.StatementLine{
			Position: mlif.Position,
			Line: ast.Line{
				Statements: []ast.Statement{
					&ast.IfStatement{
						Position: mlif.Position,
						Condition: &ast.UnaryOperation{
							Operator: "not",
							Exp:      mlif.Condition,
						},
						IfBlock: []ast.Statement{
							&nast.GoToLabelStatement{
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
		repl = append(repl, &nast.StatementLine{
			Position: mlif.Position,
			Line: ast.Line{
				Statements: []ast.Statement{
					&nast.GoToLabelStatement{
						Position: mlif.Position,
						Label:    skipElse,
					},
				},
			},
		})
	}

	repl = append(repl, &nast.StatementLine{
		Position: mlif.Position,
		Label:    skipIf,
		Line: ast.Line{
			Statements: []ast.Statement{},
		},
	})

	if mlif.ElseBlock != nil {
		for _, ifblling := range mlif.ElseBlock {
			repl = append(repl, ifblling)
		}
	}

	repl = append(repl, &nast.StatementLine{
		Position: mlif.Position,
		Label:    skipElse,
		Line: ast.Line{
			Statements: []ast.Statement{},
		},
	})

	*counter++
	return ast.NewNodeReplacement(repl...)
}

// convertWhileLoops converts while loops into yolol-code
func (c *Converter) convertWhileLoops(p ast.Node) error {
	counter := 0
	f := func(node ast.Node, visitType int) error {
		if loop, is := node.(*nast.WhileLoop); is && visitType == ast.PostVisit {
			startLabel := fmt.Sprintf("while%d", counter)
			endLabel := fmt.Sprintf("endwhile%d", counter)
			repl := []ast.Node{
				&nast.StatementLine{
					Position: loop.Position,
					Label:    startLabel,
					Line: ast.Line{
						Statements: []ast.Statement{
							&ast.IfStatement{
								Position: loop.Condition.Start(),
								Condition: &ast.UnaryOperation{
									Operator: "not",
									Exp:      loop.Condition,
								},
								IfBlock: []ast.Statement{
									&nast.GoToLabelStatement{
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
			repl = append(repl, &nast.StatementLine{
				Position: loop.Position,
				Line: ast.Line{
					Statements: []ast.Statement{
						&nast.GoToLabelStatement{
							Position: loop.Position,
							Label:    startLabel,
						},
					},
				},
			})

			repl = append(repl, &nast.StatementLine{
				Position: loop.Position,
				Label:    endLabel,
				Line: ast.Line{
					Statements: []ast.Statement{},
				},
			})

			counter++
			return ast.NewNodeReplacement(repl...)
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

// filterLines removes empty lines and constant declarations from the program
func (c *Converter) filterLines(p ast.Node) error {
	f := func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *nast.StatementLine:
			if n.Label == "" && len(n.Statements) == 0 && !n.HasEOL && !n.HasBOL {
				// empty line
				return ast.NewNodeReplacement()
			}
		case *nast.ConstDeclaration:
			return ast.NewNodeReplacement()
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

// mergeExecutableLines is a type-wrapper for mergeStatementLines
func (c *Converter) mergeExecutableLines(lines []nast.ExecutableLine) ([]nast.ExecutableLine, error) {
	inp := make([]*nast.StatementLine, len(lines))
	for i, elem := range lines {
		inp[i] = elem.(*nast.StatementLine)
	}
	interm, err := c.mergeStatementLines(inp)
	if err != nil {
		return nil, err
	}
	outp := make([]nast.ExecutableLine, len(interm))
	for i, elem := range interm {
		outp[i] = elem
	}
	return outp, nil
}

// mergeNololLines is a type-wrapper for mergeStatementLines
func (c *Converter) mergeNololLines(lines []nast.Line) ([]nast.Line, error) {
	inp := make([]*nast.StatementLine, len(lines))
	for i, elem := range lines {
		inp[i] = elem.(*nast.StatementLine)
	}
	interm, err := c.mergeStatementLines(inp)
	if err != nil {
		return nil, err
	}
	outp := make([]nast.Line, len(interm))
	for i, elem := range interm {
		outp[i] = elem
	}
	return outp, nil
}

// mergeStatementLines merges consectuive statementlines into as few lines as possible
func (c *Converter) mergeStatementLines(lines []*nast.StatementLine) ([]*nast.StatementLine, error) {
	maxlen := 70
	newLines := make([]*nast.StatementLine, 0, len(lines))
	i := 0
	for i < len(lines) {
		current := &nast.StatementLine{
			Line: ast.Line{
				Statements: []ast.Statement{},
			},
			Label:    lines[i].Label,
			Position: lines[i].Position,
			HasEOL:   lines[i].HasEOL,
		}
		current.Statements = append(current.Statements, lines[i].Statements...)
		newLines = append(newLines, current)

		if current.HasEOL {
			// no lines may MUST be appended to a line having EOL
			i++
			continue
		}

		for i+1 < len(lines) {
			currlen := getLengthOfLine(&current.Line)

			if currlen > maxlen {
				return newLines, &parser.Error{
					Message:       "The line is too long (>70 characters) to be converted to yolol, even after optimization.",
					StartPosition: current.Start(),
					EndPosition:   current.End(),
				}
			}

			nextline := lines[i+1]
			nextlen := getLengthOfLine(&nextline.Line)

			if nextline.Label == "" && currlen+nextlen <= maxlen && !nextline.HasBOL {
				current.Statements = append(current.Statements, nextline.Statements...)
				i++
				if nextline.HasEOL {
					break
				}
			} else {
				break
			}
		}
		i++
	}
	return newLines, nil
}

// getLengthOfLine returns the amount of characters needed to represent the given line as yolol-code
func getLengthOfLine(line ast.Node) int {
	ygen := parser.Printer{}
	ygen.UnknownHandlerFunc = func(node ast.Node, visitType int) (string, error) {
		if _, is := node.(*nast.GoToLabelStatement); is {
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

func (c *Converter) insertLineCounter(p *nast.Program) {
	counterVar := reservedTimeVariableReplaced
	for _, line := range p.Lines {
		if stmtline, is := line.(*nast.StatementLine); is {
			stmts := make([]ast.Statement, 1, len(stmtline.Statements)+1)
			stmts[0] = &ast.Dereference{
				Variable:    counterVar,
				Operator:    "++",
				PrePost:     "Post",
				IsStatement: true,
			}
			stmts = append(stmts, stmtline.Statements...)
			stmtline.Statements = stmts
		}
	}
}

// convertLineTypes converts nolol line-types to yolol-types
func (c *Converter) convertLineTypes(p *nast.Program) *ast.Program {
	newprog := ast.Program{
		Lines: make([]*ast.Line, 0),
	}
	for _, rawline := range p.Lines {
		switch line := rawline.(type) {
		case *nast.StatementLine:
			newline := &ast.Line{
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
