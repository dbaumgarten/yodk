package nolol

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/optimizers"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// special error that is emitted if a nolol if can not be converted to an inline yolol-if
var errInlineIfImpossible = fmt.Errorf("Can not convert to inline if")

// reservedTimeVariable is the variable used to track passed time
var reservedTimeVariable = "_time"

// Converter can convert a nolol-ast to a yolol-ast
type Converter struct {
	files               FileSystem
	jumpLabels          map[string]int
	constants           map[string]interface{}
	usesTimeTracking    bool
	iflabelcounter      int
	waitlabelcounter    int
	whilelabelcounter   int
	sexpOptimizer       *optimizers.StaticExpressionOptimizer
	boolexpOptimizer    *optimizers.ExpressionInversionOptimizer
	varnameOptimizer    *optimizers.VariableNameOptimizer
	includecount        int
	macros              map[string]*nast.MacroDefinition
	macroLevel          []string
	macroInsertionCount int
	debug               bool
}

// NewConverter creates a new converter
func NewConverter() *Converter {
	return &Converter{
		jumpLabels:       make(map[string]int),
		constants:        make(map[string]interface{}),
		macros:           make(map[string]*nast.MacroDefinition),
		macroLevel:       make([]string, 0),
		sexpOptimizer:    optimizers.NewStaticExpressionOptimizer(),
		boolexpOptimizer: &optimizers.ExpressionInversionOptimizer{},
		varnameOptimizer: optimizers.NewVariableNameOptimizer(),
	}
}

// ConvertFile is a shortcut that loads a file from the file-system, parses it and directly convertes it.
// mainfile is the path to the file on the disk.
// All included are loaded relative to the mainfile.
func (c *Converter) ConvertFile(mainfile string) (*ast.Program, error) {
	files := DiskFileSystem{
		Dir: filepath.Dir(mainfile),
	}
	return c.ConvertFileEx(filepath.Base(mainfile), files)
}

// ConvertFileEx acts like ConvertFile, but allows the passing of a custom filesystem from which the source files
// are retrieved. This way, files that are not stored on disk can be converted
func (c *Converter) ConvertFileEx(mainfile string, files FileSystem) (*ast.Program, error) {
	file, err := files.Get(mainfile)
	if err != nil {
		return nil, err
	}
	p := NewParser()
	p.Debug(c.debug)
	parsed, err := p.Parse(file)
	if err != nil {
		return nil, err
	}
	return c.Convert(parsed, files)
}

// Debug enables/disables debug logging
func (c *Converter) Debug(b bool) {
	c.debug = b
}

// Convert converts a nolol-program to a yolol-program
// files is an object to access files that are referenced in prog's include directives
func (c *Converter) Convert(prog *nast.Program, files FileSystem) (*ast.Program, error) {
	c.files = files

	c.usesTimeTracking = usesTimeTracking(prog)
	// reserve a name for use in time-tracking
	c.varnameOptimizer.OptimizeVarName(reservedTimeVariable)

	err := c.convertNodes(prog)
	if err != nil {
		return nil, err
	}

	// merge the statemens of the program as good as possible
	merged, err := c.mergeNololElements(prog.Elements)
	if err != nil {
		return nil, err
	}
	prog.Elements = merged

	// find all line-labels
	err = c.findJumpLabels(prog)
	if err != nil {
		return nil, err
	}

	// resolve jump-labels
	err = c.replaceGotoLabels(prog)
	if err != nil {
		return nil, err
	}

	if c.usesTimeTracking {
		c.insertLineCounter(prog)
	}

	// at this point the program consists entirely of statement-lines which contain pure yolol-code
	out := &ast.Program{
		Lines: make([]*ast.Line, len(prog.Elements)),
	}

	for i, element := range prog.Elements {
		line := element.(*nast.StatementLine)
		out.Lines[i] = &ast.Line{
			Position:   line.Position,
			Statements: line.Statements,
		}
	}

	return out, nil
}

func (c *Converter) maxLineLength() int {
	if !c.usesTimeTracking {
		return 70
	}
	return 70 - 4
}

func (c *Converter) convertNodes(node ast.Node) error {
	f := func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *ast.Assignment:
			if visitType == ast.PostVisit {
				return c.convertAssignment(n)
			}
		case *nast.ConstDeclaration:
			return c.convertConstDecl(n)
		case *nast.MacroDefinition:
			// using pre-visit here is important
			// the definition must be resolved, BEFORE its contents are processed
			if visitType == ast.PreVisit {
				return c.convertMacroDef(n)
			}
		case *nast.MacroInsetion:
			if visitType == ast.PreVisit {
				c.macroLevel = append(c.macroLevel, n.Name+":"+strconv.Itoa(n.Start().Line))
				return c.convertMacroInsertion(n)
			}
		case *nast.IncludeDirective:
			return c.convertInclude(n)
		case *nast.WaitDirective:
			if visitType == ast.PostVisit {
				return c.convertWait(n)
			}
		case *ast.FuncCall:
			if visitType == ast.PostVisit {
				return c.convertBuiltinFunction(n)
			}
		case *ast.Dereference:
			return c.convertDereference(n)
		case *nast.MultilineIf:
			if visitType == ast.PostVisit {
				return c.convertIf(n)
			}
		case *nast.WhileLoop:
			if visitType == ast.PostVisit {
				return c.convertWhileLoop(n)
			}
		case *ast.UnaryOperation:
		case *ast.BinaryOperation:
			if visitType == ast.PostVisit {
				repl := c.sexpOptimizer.OptimizeExpressionNonRecursive(n)
				if repl != nil {
					return ast.NewNodeReplacementSkip(repl)
				}
				return nil
			}
		case *nast.Trigger:
			if n.Kind == "macroleft" {
				c.macroLevel = c.macroLevel[:len(c.macroLevel)-1]
				return ast.NewNodeReplacement()
			}
		}

		return nil
	}
	return node.Accept(ast.VisitorFunc(f))
}

// convertAssignment optimizes the variable name and the expression of an assignment
func (c *Converter) convertAssignment(ass *ast.Assignment) error {
	ass.VariableDisplayName = c.varnameOptimizer.OptimizeVarName(ass.Variable)
	return nil
}

// convertMacroDef takes a macro definition, stores it for later use and removes the definition from the code
func (c *Converter) convertMacroDef(def *nast.MacroDefinition) error {
	c.macros[def.Name] = def
	// remove the node from the output-code
	return ast.NewNodeReplacementSkip()
}

// convert a macro insetion, by inserting the code defined by the macro
func (c *Converter) convertMacroInsertion(ins *nast.MacroInsetion) error {
	c.macroInsertionCount++

	if c.macroInsertionCount > 20 {
		return &parser.Error{
			Message:       "Error when processing macros: Macro-loop detected",
			StartPosition: ast.NewPosition("", 1, 1),
			EndPosition:   ast.NewPosition("", 20, 70),
		}
	}

	m, defined := c.macros[ins.Name]
	if !defined {
		return &parser.Error{
			Message:       fmt.Sprintf("No macro named '%s' defined", ins.Name),
			StartPosition: ins.Start(),
			EndPosition:   ins.End(),
		}
	}

	if len(m.Arguments) != len(ins.Arguments) {
		return &parser.Error{
			Message:       fmt.Sprintf("Wrong number of arguments for %s, got %d but want %d", ins.Name, len(ins.Arguments), len(m.Arguments)),
			StartPosition: ins.Start(),
			EndPosition:   ins.End(),
		}
	}

	copy := nast.CopyAst(m).(*nast.MacroDefinition)

	// gather replacements
	replacements := make(map[string]ast.Expression)
	for i := range ins.Arguments {
		replacements[m.Arguments[i]] = ins.Arguments[i]
	}

	performReplacements := func(node ast.Node, visitType int) error {
		// replace the variable name inside assignments
		if ass, is := node.(*ast.Assignment); is && visitType == ast.PreVisit {
			if replacement, exists := replacements[ass.VariableDisplayName]; exists {
				if replacementVariable, isvar := replacement.(*ast.Dereference); isvar && replacementVariable.Operator == "" {
					ass.Variable = replacementVariable.Variable
					ass.VariableDisplayName = replacementVariable.VariableDisplayName
				} else {
					return &parser.Error{
						Message:       "This argument must be a variable name (and not any other expression)",
						StartPosition: replacement.Start(),
						EndPosition:   replacement.End(),
					}
				}
			} else if !strings.HasPrefix(ass.Variable, ":") {
				// replace non-global vars (which are not arguments) with a insertion-scoped version
				ass.Variable = strings.Join(c.macroLevel, "_") + "_" + ass.Variable
			}
		}
		// replace the variable name of dereferences
		if deref, is := node.(*ast.Dereference); is && visitType == ast.SingleVisit {
			if replacement, exists := replacements[deref.VariableDisplayName]; exists {
				if replacementVariable, isvar := replacement.(*ast.Dereference); isvar {
					if deref.Operator != "" && replacementVariable.Operator != "" {
						return &parser.Error{
							Message:       "You can not use pre/post-operators for this particular argument",
							StartPosition: replacement.Start(),
							EndPosition:   replacement.End(),
						}
					}
					if deref.Operator != "" {
						replacementVariable.Operator = deref.Operator
						replacementVariable.PrePost = deref.PrePost
					}
				} else if deref.Operator != "" {
					return &parser.Error{
						Message:       "This argument must be a variable name (and not any other expression)",
						StartPosition: replacement.Start(),
						EndPosition:   replacement.End(),
					}
				}
				return ast.NewNodeReplacementSkip(replacement)
			} else if !strings.HasPrefix(deref.Variable, ":") {
				if _, isConst := c.constants[deref.VariableDisplayName]; !isConst {
					// replace non-global vars (which are not arguments) with a insertion-scoped version
					deref.Variable = strings.Join(c.macroLevel, "_") + "_" + deref.Variable
				}
			}
		}
		return nil
	}

	err := copy.Accept(ast.VisitorFunc(performReplacements))
	if err != nil {
		return err
	}

	nodes := make([]ast.Node, len(copy.Block.Elements)+1)
	for i, el := range copy.Block.Elements {
		nodes[i] = el
	}
	nodes[len(nodes)-1] = &nast.Trigger{
		Kind: "macroleft",
	}

	// remove the node from the output-code
	return ast.NewNodeReplacement(nodes...)
}

// convertConstDecl converts a const declarationto yolol by discarding it, but saving the declared value
func (c *Converter) convertConstDecl(decl *nast.ConstDeclaration) error {
	switch val := decl.Value.(type) {
	case *ast.StringConstant:
		c.constants[decl.Name] = val
		break
	case *ast.NumberConstant:
		c.constants[decl.Name] = val
		break
	default:
		return &parser.Error{
			Message:       "Only constant values can be the value of a constant declaration",
			StartPosition: decl.Start(),
			EndPosition:   decl.End(),
		}
	}
	return ast.NewNodeReplacement()
}

// resolveIncludes searches for include-directives and inserts the lines of the included files
func (c *Converter) convertInclude(include *nast.IncludeDirective) error {
	p := NewParser()

	c.includecount++
	if c.includecount > 20 {
		return &parser.Error{
			Message:       "Error when processing includes: Include-loop detected",
			StartPosition: ast.NewPosition("", 1, 1),
			EndPosition:   ast.NewPosition("", 20, 70),
		}
	}

	file, err := c.files.Get(include.File)
	if err != nil {
		return &parser.Error{
			Message:       fmt.Sprintf("Error when opening included file '%s': %s", include.File, err.Error()),
			StartPosition: include.Start(),
			EndPosition:   include.End(),
		}
	}
	p.SetFilename(include.File)
	parsed, err := p.Parse(file)
	if err != nil {
		// override the position of the error with the position of the include
		// this way the error gets displayed at the correct location
		// the message does contain the original location
		return &parser.Error{
			Message:       err.Error(),
			StartPosition: include.Start(),
			EndPosition:   include.End(),
		}
	}

	if usesTimeTracking(parsed) {
		c.usesTimeTracking = true
	}

	replacements := make([]ast.Node, len(parsed.Elements))
	for i := range parsed.Elements {
		replacements[i] = parsed.Elements[i]
	}
	return ast.NewNodeReplacement(replacements...)
}

// convert a wait directive to yolol
func (c *Converter) convertWait(wait *nast.WaitDirective) error {
	label := fmt.Sprintf("wait%d", c.waitlabelcounter)
	return ast.NewNodeReplacement(&nast.StatementLine{
		Label:  label,
		HasEOL: true,
		Line: ast.Line{
			Position: wait.Start(),
			Statements: []ast.Statement{
				&ast.IfStatement{
					Position:  wait.Start(),
					Condition: wait.Condition,
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

// convert a built-in function to yolol
func (c *Converter) convertBuiltinFunction(function *ast.FuncCall) error {
	switch function.Function {
	case "time":
		c.usesTimeTracking = true
		return ast.NewNodeReplacementSkip(&ast.Dereference{
			Variable:            reservedTimeVariable,
			VariableDisplayName: c.varnameOptimizer.OptimizeVarName(reservedTimeVariable),
		})
	}
	return nil
}

// checkes, if the program uses nolols time-tracking feature
func usesTimeTracking(n ast.Node) bool {
	uses := false
	f := func(node ast.Node, visitType int) error {
		if function, is := node.(*ast.FuncCall); is {
			if function.Function == "time" {
				uses = true
			}
		}
		return nil
	}
	n.Accept(ast.VisitorFunc(f))
	return uses
}

// convertDereference replaces mentionings of constants with the value of the constant
func (c *Converter) convertDereference(deref *ast.Dereference) error {
	if deref.Operator == "" {
		// we are dereferencing a constant
		if value, exists := c.constants[deref.VariableDisplayName]; exists {
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
	// we are dereferencing a variable
	deref.VariableDisplayName = c.varnameOptimizer.OptimizeVarName(deref.Variable)
	return nil
}

// findJumpLabels finds all line-labels in the program
func (c *Converter) findJumpLabels(p ast.Node) error {
	c.jumpLabels = make(map[string]int)
	linecounter := 0
	f := func(node ast.Node, visitType int) error {
		if line, isLine := node.(*nast.StatementLine); isLine {
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

// replaceGotoLabels replaces all goto labels with the appropriate line-number
func (c *Converter) replaceGotoLabels(p ast.Node) error {
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

// convertIf converts nolol multiline-ifs to yolol
func (c *Converter) convertIf(mlif *nast.MultilineIf) error {
	endif := fmt.Sprintf("endif%d", c.iflabelcounter)
	repl := []ast.Node{}
	for i := range mlif.Conditions {
		endlabel := ""
		if mlif.ElseBlock != nil || i < len(mlif.Conditions)-1 {
			endlabel = endif
		}
		condline, err := c.convertConditionInline(mlif, i, endlabel)
		if err == nil {
			repl = append(repl, condline)
		} else {
			condlines := c.convertConditionMultiline(mlif, i, endlabel)
			repl = append(repl, condlines...)
		}
	}

	if mlif.ElseBlock != nil {
		for _, elseline := range mlif.ElseBlock.Elements {
			repl = append(repl, elseline)
		}
	}

	repl = append(repl, &nast.StatementLine{
		Position: mlif.Position,
		Label:    endif,
		Line: ast.Line{
			Statements: []ast.Statement{},
		},
	})

	c.iflabelcounter++
	return ast.NewNodeReplacement(repl...)
}

// convertConditionInline converts a single conditional block of a multiline if and tries to produce a single yolol if
func (c *Converter) convertConditionInline(mlif *nast.MultilineIf, index int, endlabel string) (ast.Node, error) {
	mergedIfElements, _ := c.mergeNololNestableElements(mlif.Blocks[index].Elements)

	if len(mergedIfElements) > 1 || (len(mergedIfElements) > 0 && mergedIfElements[0].(*nast.StatementLine).Label != "") {
		return nil, errInlineIfImpossible
	}

	statements := []ast.Statement{}
	if len(mergedIfElements) > 0 {
		statements = mergedIfElements[0].(*nast.StatementLine).Line.Statements
		if endlabel != "" {
			statements = append(statements, &nast.GoToLabelStatement{
				Label: endlabel,
			})
		}
	}

	repl := &nast.StatementLine{
		Position: mlif.Position,
		Line: ast.Line{
			Statements: []ast.Statement{
				&ast.IfStatement{
					Position:  mlif.Position,
					Condition: mlif.Conditions[index],
					IfBlock:   statements,
				},
			},
		},
	}

	if getLengthOfLine(&repl.Line) > c.maxLineLength() {
		return nil, errInlineIfImpossible
	}

	return repl, nil
}

// convertConditionMultiline converts a single conditional block of a multiline if and produces
// multiple lines, because a single-line if would become too long
func (c *Converter) convertConditionMultiline(mlif *nast.MultilineIf, index int, endlabel string) []ast.Node {
	skipIf := fmt.Sprintf("iflbl%d-%d", c.iflabelcounter, index)
	condition := c.boolexpOptimizer.OptimizeExpression(&ast.UnaryOperation{
		Operator: "not",
		Exp:      mlif.Conditions[index],
	})
	repl := []ast.Node{
		&nast.StatementLine{
			Position: mlif.Position,
			Line: ast.Line{
				Statements: []ast.Statement{
					&ast.IfStatement{
						Position:  mlif.Position,
						Condition: condition,
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

	for _, ifblling := range mlif.Blocks[index].Elements {
		repl = append(repl, ifblling)
	}

	if endlabel != "" {
		repl = append(repl, &nast.StatementLine{
			Position: mlif.Position,
			Line: ast.Line{
				Statements: []ast.Statement{
					&nast.GoToLabelStatement{
						Position: mlif.Position,
						Label:    endlabel,
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

	return repl
}

// convertWhileLoop converts while loops into yolol-code
func (c *Converter) convertWhileLoop(loop *nast.WhileLoop) error {
	startLabel := fmt.Sprintf("while%d", c.whilelabelcounter)
	endLabel := fmt.Sprintf("endwhile%d", c.whilelabelcounter)

	condition := c.boolexpOptimizer.OptimizeExpression(&ast.UnaryOperation{
		Operator: "not",
		Exp:      loop.Condition,
	})
	repl := []ast.Node{
		&nast.StatementLine{
			Position: loop.Position,
			Label:    startLabel,
			Line: ast.Line{
				Statements: []ast.Statement{
					&ast.IfStatement{
						Position:  loop.Condition.Start(),
						Condition: condition,
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

	for _, blockline := range loop.Block.Elements {
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

	c.whilelabelcounter++
	return ast.NewNodeReplacement(repl...)

}

// mergeNololNestableElements is a type-wrapper for mergeStatementElements
func (c *Converter) mergeNololNestableElements(lines []nast.NestableElement) ([]nast.NestableElement, error) {
	inp := make([]*nast.StatementLine, len(lines))
	for i, elem := range lines {
		line, isline := elem.(*nast.StatementLine)
		if !isline {
			return nil, parser.Error{
				Message: fmt.Sprintf("Err: Found unconverted nolol-element: %T", elem),
			}
		}
		inp[i] = line
	}
	interm, err := c.mergeStatementElements(inp)
	if err != nil {
		return nil, err
	}
	outp := make([]nast.NestableElement, len(interm))
	for i, elem := range interm {
		outp[i] = elem
	}
	return outp, nil
}

// mergeNololElements is a type-wrapper for mergeStatementElements
func (c *Converter) mergeNololElements(lines []nast.Element) ([]nast.Element, error) {
	inp := make([]*nast.StatementLine, len(lines))
	for i, elem := range lines {
		line, isline := elem.(*nast.StatementLine)
		if !isline {
			return nil, parser.Error{
				Message: fmt.Sprintf("Err: Found unconverted nolol-element: %T", elem),
			}
		}
		inp[i] = line
	}
	interm, err := c.mergeStatementElements(inp)
	if err != nil {
		return nil, err
	}
	outp := make([]nast.Element, len(interm))
	for i, elem := range interm {
		outp[i] = elem
	}
	return outp, nil
}

// mergeStatementElements merges consectuive statementlines into as few lines as possible
func (c *Converter) mergeStatementElements(lines []*nast.StatementLine) ([]*nast.StatementLine, error) {
	maxlen := c.maxLineLength()
	newElements := make([]*nast.StatementLine, 0, len(lines))
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
		newElements = append(newElements, current)

		if current.HasEOL {
			// no lines may MUST be appended to a line having EOL
			i++
			continue
		}

		for i+1 < len(lines) {
			currlen := getLengthOfLine(&current.Line)

			if currlen > maxlen {
				return newElements, &parser.Error{
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
	return newElements, nil
}

//getLengthOfLine returns the amount of characters needed to represent the given line as yolol-code
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

// inserts the line-counting statement into the beginning of each line
func (c *Converter) insertLineCounter(p *nast.Program) {
	for _, line := range p.Elements {
		if stmtline, is := line.(*nast.StatementLine); is {
			stmts := make([]ast.Statement, 1, len(stmtline.Statements)+1)
			stmts[0] = &ast.Dereference{
				Variable:            reservedTimeVariable,
				VariableDisplayName: c.varnameOptimizer.OptimizeVarName(reservedTimeVariable),
				Operator:            "++",
				PrePost:             "Post",
				IsStatement:         true,
			}
			stmts = append(stmts, stmtline.Statements...)
			stmtline.Statements = stmts
		}
	}
}
