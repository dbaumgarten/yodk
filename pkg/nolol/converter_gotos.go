package nolol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// getJumpLabel is a case-insensitive getter for c.jumpLabels
func (c *Converter) getJumpLabel(name string) (int, bool) {
	name = strings.ToLower(name)
	val, exists := c.jumpLabels[name]
	return val, exists
}

// setJumpLabel is a case-insensitive setter for c.jumpLabels
func (c *Converter) setJumpLabel(name string, val int) {
	name = strings.ToLower(name)
	c.jumpLabels[name] = val
}

// removeDuplicateGotos removes gotos that are unreachable, because they are directly behin another goto
func (c *Converter) removeDuplicateGotos(p ast.Node) error {
	lastWasGoto := false
	f := func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *nast.GoToLabelStatement:
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

// removeUnusedLabels removes all labels that are not used by at least one goto
// this helpes in reducing the number of lines
func (c *Converter) removeUnusedLabels(p *nast.Program) error {
	used := make(map[string]bool)
	f := func(node ast.Node, visitType int) error {
		if gotostmt, isGoto := node.(*nast.GoToLabelStatement); isGoto {
			used[gotostmt.Label] = true
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
	getTargetGoto := func(label string) *nast.GoToLabelStatement {
		foundlabel := false
		for _, element := range p.Elements {
			if line, isLine := element.(*nast.StatementLine); isLine {
				if line.Label == label {
					foundlabel = true
				}
				if foundlabel {
					if len(line.Statements) > 0 {
						if gotostmt, is := line.Statements[0].(*nast.GoToLabelStatement); is {
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
		if gotostmt, isGoto := node.(*nast.GoToLabelStatement); isGoto {
			targetgoto := getTargetGoto(gotostmt.Label)
			if targetgoto != nil {
				gotostmt.Label = targetgoto.Label
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
		if line, isLine := node.(*nast.StatementLine); isLine {
			if visitType == ast.PreVisit {
				linecounter++
				if line.Label != "" {
					_, exists := c.getJumpLabel(line.Label)
					if exists {
						return &parser.Error{
							Message:       fmt.Sprintf("Duplicate declaration of jump-label: %s", line.Label),
							StartPosition: line.Start(),
							EndPosition:   line.Start(),
						}
					}
					c.setJumpLabel(line.Label, linecounter)
				}
				// remove all empty lines
				if len(line.Statements) == 0 && !line.HasEOL {
					linecounter--
					return ast.NewNodeReplacement()
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
			line, exists := c.getJumpLabel(gotostmt.Label)
			if !exists {
				return &parser.Error{
					Message:       "Unknown jump-label: " + gotostmt.Label,
					StartPosition: gotostmt.Start(),
					EndPosition:   gotostmt.End(),
				}
			}
			repl := &ast.GoToStatement{
				Position: gotostmt.Position,
				Line: &ast.NumberConstant{
					Position: gotostmt.Start(),
					Value:    strconv.Itoa(line),
				},
			}
			return ast.NewNodeReplacement(repl)
		}
		return nil
	}
	return p.Accept(ast.VisitorFunc(f))
}
