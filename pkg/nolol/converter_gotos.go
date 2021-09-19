package nolol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// getLineLabel is a case-insensitive getter for c.lineLabels
func (c *Converter) getLineLabel(name string) (int, bool) {
	name = strings.ToLower(name)
	val, exists := c.lineLabels[name]
	return val, exists
}

// storeLineLabel is a case-insensitive setter for c.lineLabels
func (c *Converter) storeLineLabel(name string, val int) {
	name = strings.ToLower(name)
	c.lineLabels[name] = val
}

func (c *Converter) gotoForLabel(label string) *ast.GoToStatement {
	c.storeLineLabel(label, -1)
	return &ast.GoToStatement{
		Position: ast.UnknownPosition,
		Line: &ast.Dereference{
			Variable: label,
		},
	}
}

func (c *Converter) gotoForLabelPos(label string, pos ast.Position) *ast.GoToStatement {
	c.storeLineLabel(label, -1)
	return &ast.GoToStatement{
		Position: pos,
		Line: &ast.Dereference{
			Position: pos.Add(4),
			Variable: label,
		},
	}
}

func (c *Converter) getGotoDestinationLabel(g *ast.GoToStatement) string {
	if deref, is := g.Line.(*ast.Dereference); is {
		if _, isLabel := c.getLineLabel(deref.Variable); isLabel {
			return deref.Variable
		}
	}
	return ""
}

func (c *Converter) isLineLabelDereference(deref *ast.Dereference) bool {
	if _, isLabel := c.getLineLabel(deref.Variable); isLabel {
		return true
	}
	return false
}

// findLineLabels finds all line-labels in the program
func (c *Converter) findLineLabels(p ast.Node, removeEmptyLines bool) error {
	c.lineLabels = make(map[string]int)
	linecounter := 0
	f := func(node ast.Node, visitType int) error {
		// skip all macro-definitions
		// lables inside macros are resolved upon macro-insertion
		if _, isMacro := node.(*nast.MacroDefinition); isMacro {
			return ast.NewNodeReplacementSkip(node)
		}
		if line, isLine := node.(*nast.StatementLine); isLine {
			if visitType == ast.PreVisit {
				linecounter++
				if line.Label != "" {
					_, exists := c.getLineLabel(line.Label)
					if exists {
						return &parser.Error{
							Message:       fmt.Sprintf("Duplicate declaration of line-label: %s", line.Label),
							StartPosition: line.Start(),
							EndPosition:   line.Start(),
						}
					}
					c.storeLineLabel(line.Label, linecounter)
				}
				// remove all empty lines
				if removeEmptyLines && len(line.Statements) == 0 && !line.HasEOL {
					linecounter--
					return ast.NewNodeReplacement()
				}
			}
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

// replaceLineLabels replaces all line labels with the appropriate line-number
func (c *Converter) replaceLineLabels(p ast.Node) error {
	f := func(node ast.Node, visitType int) error {
		if deref, is := node.(*ast.Dereference); is {
			line, exists := c.getLineLabel(deref.Variable)
			if deref.Variable == "_start" {
				line = 1
				exists = true
			}
			if exists {
				return ast.NewNodeReplacement(&ast.NumberConstant{
					Position: deref.Position,
					Value:    strconv.Itoa(line),
				})
			}
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

// removeDuplicateGotos removes gotos that are unreachable, because they are directly behind another goto
func (c *Converter) removeDuplicateGotos(p ast.Node) error {
	lastWasGoto := false
	f := func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *ast.GoToStatement:
			if lastWasGoto {
				return ast.NewNodeReplacement()
			}
			lastWasGoto = true
			break
		case *nast.StatementLine:
			if visitType == ast.PreVisit && n.Label != "" {
				lastWasGoto = false
			}
		default:
			if visitType < 0 {
				lastWasGoto = false
			}
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

// removeUnusedLabels removes all labels that are not used at least once
// this helpes in reducing the number of lines
func (c *Converter) removeUnusedLabels(p *nast.Program) error {
	used := make(map[string]bool)
	f := func(node ast.Node, visitType int) error {
		if deref, is := node.(*ast.Dereference); is {
			_, exists := c.getLineLabel(deref.Variable)
			if exists {
				used[deref.Variable] = true
			}
		}
		return nil
	}
	err := p.Accept(ast.VisitorFunc(f))
	if err != nil {
		return err
	}
	for _, element := range p.Elements {
		if line, isLine := element.(*nast.StatementLine); isLine {
			if _, isused := used[line.Label]; !isused {
				line.Label = ""
			}
		}
	}
	return nil
}

// resolveGotoChains replaces the target-label of gotos that lead to another goto with the target of the second goto
func (c *Converter) resolveGotoChains(p *nast.Program) error {

	// this function finds the goto another goto jumps to (if it exists)
	getTargetGoto := func(label string) *ast.GoToStatement {
		found := false
		for _, element := range p.Elements {
			if line, isLine := element.(*nast.StatementLine); isLine {
				if line.Label == label {
					found = true
				}
				if found {
					if len(line.Statements) > 0 {
						if gotostmt, is := line.Statements[0].(*ast.GoToStatement); is {
							return gotostmt
						}
						break
					}
					if line.HasEOL {
						break
					}
				}
			}
		}
		return nil
	}

	f := func(node ast.Node, visitType int) error {
		if gotostmt, isGoto := node.(*ast.GoToStatement); isGoto {
			targetLabel := c.getGotoDestinationLabel(gotostmt)
			if targetLabel != "" {
				targetgoto := getTargetGoto(targetLabel)
				if targetgoto != nil {
					gotostmt.Line = targetgoto.Line
				}
			}
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}

// addFinalGoto adds a goto 1 to the end of the programm to speed up execution
func (c *Converter) addFinalGoto(prog *nast.Program) error {
	pos := ast.UnknownPosition

	prog.Elements = append(prog.Elements, &nast.StatementLine{
		Line: ast.Line{
			Position: pos,
			Statements: []ast.Statement{
				c.gotoForLabelPos("_start", pos),
			},
		},
	})

	return nil
}

// removeFinalGotoIfNeeded removes the final goto added by addFinalGoto
// if it is the only reason for the programm to bust the line-limit
// as this is the last step of the converion, the input is a yolol ast
func (c *Converter) removeFinalGotoIfNeeded(prog *ast.Program) error {
	// one line too much
	if len(prog.Lines) == 21 {
		line := prog.Lines[len(prog.Lines)-1]
		// line has only one element
		if len(line.Statements) == 1 {
			// stmt is a goto
			if stmt, isgoto := line.Statements[0].(*ast.GoToStatement); isgoto {
				// goto target is line 1
				if number, isnumber := stmt.Line.(*ast.NumberConstant); isnumber {
					if number.Value == "1" {
						// remove the line
						prog.Lines = prog.Lines[:len(prog.Lines)-1]
					}
				}
			}
		}
	}
	return nil
}
