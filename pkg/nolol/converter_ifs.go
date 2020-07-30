package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// special error that is emitted if a nolol if can not be converted to an inline yolol-if
var errInlineIfImpossible = fmt.Errorf("Can not convert to inline if")

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
		Position: mlif.End(),
		Label:    endif,
		Line: ast.Line{
			Statements: []ast.Statement{},
		},
	})

	c.iflabelcounter++
	return ast.NewNodeReplacementSkip(repl...)
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
			var pos ast.Position
			if len(statements) == 0 {
				pos = mlif.Positions[index]
			} else {
				pos = statements[len(statements)-1].End()
			}
			statements = append(statements, &nast.GoToLabelStatement{
				Label:    endlabel,
				Position: pos,
			})
		}
	}

	repl := &nast.StatementLine{
		Position: mlif.Positions[index],
		Line: ast.Line{
			Statements: []ast.Statement{
				&ast.IfStatement{
					Position:  mlif.Positions[index],
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
			Position: mlif.Positions[index],
			Line: ast.Line{
				Statements: []ast.Statement{
					&ast.IfStatement{
						Position:  mlif.Positions[index],
						Condition: condition,
						IfBlock: []ast.Statement{
							&nast.GoToLabelStatement{
								Position: mlif.Positions[index],
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
			Position: mlif.Blocks[index].End(),
			Line: ast.Line{
				Position: mlif.Blocks[index].End(),
				Statements: []ast.Statement{
					&nast.GoToLabelStatement{
						Position: mlif.Blocks[index].End(),
						Label:    endlabel,
					},
				},
			},
		})
	}

	repl = append(repl, &nast.StatementLine{
		Position: mlif.Blocks[index].End(),
		Label:    skipIf,
		Line: ast.Line{
			Statements: []ast.Statement{},
		},
	})

	return repl
}
