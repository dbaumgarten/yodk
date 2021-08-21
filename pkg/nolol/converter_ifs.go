package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// special error that is emitted if a nolol if can not be converted to an inline yolol-if
var errInlineIfImpossible = fmt.Errorf("Can not convert to inline if")

// convertIf converts nolol multiline-ifs to yolol
func (c *Converter) convertIf(mlif *nast.MultilineIf, visitType int) error {

	if visitType != ast.PostVisit {
		return nil
	}

	// try if we can convert to the simplest if
	simple, err := c.convertIfTrivial(mlif)
	if err == nil {
		return ast.NewNodeReplacementSkip(simple)
	} else {
		if err != errInlineIfImpossible {
			return err
		}
	}

	// the simplest if is not possible. Fall back to more complicated versions
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
		Label: endif,
		Line: ast.Line{
			Position:   mlif.End(),
			Statements: []ast.Statement{},
		},
	})

	c.iflabelcounter++
	return ast.NewNodeReplacementSkip(repl...)
}

// convertIfTrivial tries the most simple conversion possible. A simple yolol-style if-then-else-end without any gotos.
func (c *Converter) convertIfTrivial(mlif *nast.MultilineIf) (ast.Node, error) {
	if len(mlif.Conditions) != 1 {
		return nil, errInlineIfImpossible
	}

	mergedIfElements, err := c.mergeNololNestableElements(mlif.Blocks[0].Elements)
	if err != nil {
		return nil, err
	}
	if len(mergedIfElements) > 1 || (len(mergedIfElements) > 0 && mergedIfElements[0].(*nast.StatementLine).Label != "") {
		return nil, errInlineIfImpossible
	}

	var elseStatements []ast.Statement
	if mlif.ElseBlock != nil {
		mergedElseElements, err := c.mergeNololNestableElements(mlif.ElseBlock.Elements)
		if err != nil {
			return nil, err
		}
		if len(mergedElseElements) > 1 || (len(mergedElseElements) > 0 && mergedElseElements[0].(*nast.StatementLine).Label != "") {
			return nil, errInlineIfImpossible
		}
		elseStatements = []ast.Statement{}
		if len(mergedElseElements) > 0 {
			elseStatements = mergedElseElements[0].(*nast.StatementLine).Line.Statements
		}
	}

	statements := []ast.Statement{}
	if len(mergedIfElements) > 0 {
		statements = mergedIfElements[0].(*nast.StatementLine).Line.Statements
	}

	repl := &nast.StatementLine{
		Line: ast.Line{
			Position: mlif.Positions[0],
			Statements: []ast.Statement{
				&ast.IfStatement{
					Position:  mlif.Positions[0],
					Condition: mlif.Conditions[0],
					IfBlock:   statements,
					ElseBlock: elseStatements,
				},
			},
		},
	}

	if c.getLengthOfLine(&repl.Line) > c.maxLineLength() {
		return nil, errInlineIfImpossible
	}

	return repl, nil
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
			statements = append(statements, c.gotoForLabel(endlabel))
		}
	}

	repl := &nast.StatementLine{
		Line: ast.Line{
			Position: mlif.Positions[index],
			Statements: []ast.Statement{
				&ast.IfStatement{
					Position:  mlif.Positions[index],
					Condition: mlif.Conditions[index],
					IfBlock:   statements,
				},
			},
		},
	}

	// if the user used goto inside the block, our additional goto may be useless
	c.removeDuplicateGotos(repl)

	if c.getLengthOfLine(&repl.Line) > c.maxLineLength() {
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
			Line: ast.Line{
				Position: mlif.Positions[index],
				Statements: []ast.Statement{
					&ast.IfStatement{
						Position:  mlif.Positions[index],
						Condition: condition,
						IfBlock: []ast.Statement{
							c.gotoForLabel(skipIf),
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
			Line: ast.Line{
				Position: mlif.Blocks[index].End(),
				Statements: []ast.Statement{
					c.gotoForLabel(endlabel),
				},
			},
		})
	}

	repl = append(repl, &nast.StatementLine{
		Label: skipIf,
		Line: ast.Line{
			Position:   mlif.Blocks[index].End(),
			Statements: []ast.Statement{},
		},
	})

	return repl
}
