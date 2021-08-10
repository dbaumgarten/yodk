package optimizers

import (
	"strconv"

	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

type CommonSubexpressionOptimizer struct {
	subexpressions    map[string]*subexpression
	printer           parser.Printer
	modifiedVariables map[string]bool
}

type subexpression struct {
	Expression    ast.Expression
	Count         int
	Replacement   string
	HasAssignment bool
}

func (e subexpression) shouldBeReplaced() bool {
	return e.Count > 1
}

func (c *CommonSubexpressionOptimizer) Optimize(node ast.Node) error {
	c.printer = parser.Printer{}

	switch n := node.(type) {
	case *ast.Line:
		c.subexpressions = make(map[string]*subexpression)
		c.modifiedVariables = make(map[string]bool)
		c.find(n)
		c.filter()
		c.replace(n)
		c.subexpressions = make(map[string]*subexpression)
		c.modifiedVariables = make(map[string]bool)
		c.find(n)
		c.filter()
		c.replace(n)
	case *ast.Program:
		c.Optimize(n.Lines[0])
	}

	return nil
}

func (c *CommonSubexpressionOptimizer) find(line *ast.Line) {
	f := func(node ast.Node, visitType int) error {
		if visitType == ast.PreVisit || visitType == ast.SingleVisit {
			switch n := node.(type) {
			case *ast.Assignment:
				hash, err := c.printer.Print(n.Value)
				if err != nil {
					panic(err)
				}

				if _, exists := c.subexpressions[hash]; !exists {
					c.subexpressions[hash] = &subexpression{
						Expression:    n.Value,
						Count:         0,
						HasAssignment: true,
						Replacement:   n.Variable,
					}
				}

				c.modifiedVariables[n.Variable] = true

			case ast.Expression:

				if deref, is := n.(*ast.Dereference); is {
					if deref.Operator != "" {
						c.modifiedVariables[deref.Variable] = true
					}
				}

				hash, err := c.printer.Print(n)
				if err != nil {
					panic(err)
				}

				if len(hash) < 2 {
					return nil
				}

				if sube, exists := c.subexpressions[hash]; exists {
					sube.Count++
				} else {
					c.subexpressions[hash] = &subexpression{
						Expression: n,
						Count:      1,
					}
				}
			}
		}
		return nil
	}
	line.Accept(ast.VisitorFunc(f))
}

func (c *CommonSubexpressionOptimizer) filter() {
	filtered := make(map[string]*subexpression)

	for hash, sube := range c.subexpressions {
		if sube.Count < 2 {
			continue
		}

		if c.containsModifiedVariable(sube.Expression) {
			continue
		}

		filtered[hash] = sube
	}

	c.subexpressions = filtered
}

func (c *CommonSubexpressionOptimizer) containsModifiedVariable(node ast.Node) bool {
	contains := false
	f := func(node ast.Node, visitType int) error {
		if deref, is := node.(*ast.Dereference); is {
			if _, exists := c.modifiedVariables[deref.Variable]; exists {
				contains = true
			}
		}
		return nil
	}
	node.Accept(ast.VisitorFunc(f))

	return contains
}

func (c *CommonSubexpressionOptimizer) replace(line *ast.Line) {

	counter := 0
	var currentAssignment *ast.Assignment

	f := func(node ast.Node, visitType int) error {
		if visitType == ast.PreVisit || visitType == ast.SingleVisit {
			switch n := node.(type) {
			case *ast.Assignment:
				currentAssignment = n
			case ast.Expression:
				hash, err := c.printer.Print(n)
				if err != nil {
					panic(err)
				}

				if sube, exists := c.subexpressions[hash]; exists && sube.shouldBeReplaced() {
					if sube.Replacement == "" {
						sube.Replacement = "_tmp" + strconv.Itoa(counter)
						counter++
					}
					if currentAssignment != nil && currentAssignment.Variable == sube.Replacement {
						return ast.NewNodeReplacementSkip(n)
					}
					return ast.NewNodeReplacementSkip(&ast.Dereference{
						Position: ast.UnknownPosition,
						Variable: sube.Replacement,
					})
				}
			}
		}

		if _, is := node.(*ast.Assignment); is && visitType == ast.PostVisit {
			currentAssignment = nil
		}

		return nil
	}
	line.Accept(ast.VisitorFunc(f))

	assignments := make([]ast.Statement, 0, len(c.subexpressions))
	for _, sube := range c.subexpressions {
		if sube.Replacement != "" && !sube.HasAssignment {
			sube.HasAssignment = true
			assignments = append(assignments, &ast.Assignment{
				Position: ast.UnknownPosition,
				Variable: sube.Replacement,
				Value:    sube.Expression,
				Operator: "=",
			})
		}
	}

	line.Statements = append(assignments, line.Statements...)
}
