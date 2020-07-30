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
