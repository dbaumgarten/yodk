package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

type loopinfo struct {
	Number            int
	HasBreakStatement bool
}

func (li loopinfo) StartLabel() string {
	return fmt.Sprintf("while%d", li.Number)
}

func (li loopinfo) EndLabel() string {
	return fmt.Sprintf("endwhile%d", li.Number)
}

// getCurrentLoop returns information about the innermost loop that is currently being processed
func (c *Converter) getCurrentLoop() *loopinfo {
	return &c.loopLevel[len(c.loopLevel)-1]
}

// convertWhileLoop converts while loops into yolol-code
func (c *Converter) convertWhileLoop(loop *nast.WhileLoop, visitType int) error {

	if visitType == ast.PreVisit {
		c.loopcounter++
		c.loopLevel = append(c.loopLevel, loopinfo{
			Number: c.loopcounter,
		})
		return nil
	}

	if visitType != ast.PostVisit {
		return nil
	}

	defer func() {
		c.loopLevel = c.loopLevel[:len(c.loopLevel)-1]
	}()

	currentloop := c.getCurrentLoop()

	conditionIsAlwaysTrue := false
	condition := loop.Condition
	isstatic, value := c.isStaticValue(condition)
	if isstatic && value.IsNumber() && value.Number() != 0 {
		conditionIsAlwaysTrue = true
	}

	if !conditionIsAlwaysTrue {
		inlineloop, err := c.convertWhileLoopInline(loop)
		if err == nil {
			return ast.NewNodeReplacementSkip(inlineloop...)
		}
	}

	repl := []ast.Node{
		&nast.StatementLine{
			Label: currentloop.StartLabel(),
			Line: ast.Line{
				Position:   loop.Position,
				Statements: []ast.Statement{},
			},
		},
	}

	// if the condition is always true, we do not need to add a condition-check
	// this makes infinite loops smaller
	if !conditionIsAlwaysTrue {
		condition = c.boolexpOptimizer.OptimizeExpression(&ast.UnaryOperation{
			Operator: "not",
			Exp:      condition,
			Position: condition.Start(),
		})

		repl[0].(*nast.StatementLine).Line.Statements = []ast.Statement{
			&ast.IfStatement{
				Position:  loop.Condition.Start(),
				Condition: condition,
				IfBlock: []ast.Statement{
					c.gotoForLabelPos(currentloop.EndLabel(), loop.Condition.End()),
				},
			},
		}
	}

	for _, blockline := range loop.Block.Elements {
		repl = append(repl, blockline)
	}
	repl = append(repl, &nast.StatementLine{
		Line: ast.Line{
			Position: ast.UnknownPosition,
			Statements: []ast.Statement{
				c.gotoForLabel(currentloop.StartLabel()),
			},
		},
	})

	repl = append(repl, &nast.StatementLine{
		Label: currentloop.EndLabel(),
		Line: ast.Line{
			Position:   loop.Position,
			Statements: []ast.Statement{},
		},
	})

	return ast.NewNodeReplacementSkip(repl...)
}

func (c *Converter) convertWhileLoopInline(loop *nast.WhileLoop) ([]ast.Node, error) {
	mergedIfElements, _ := c.mergeNololNestableElements(loop.Block.Elements)

	if len(mergedIfElements) > 1 || (len(mergedIfElements) > 0 && mergedIfElements[0].(*nast.StatementLine).Label != "") {
		return nil, errInlineIfImpossible
	}

	statements := []ast.Statement{}
	if len(mergedIfElements) > 0 {
		statements = mergedIfElements[0].(*nast.StatementLine).Line.Statements
	}
	statements = append(statements, c.gotoForLabel(c.getCurrentLoop().StartLabel()))

	repl := &nast.StatementLine{
		Label: c.getCurrentLoop().StartLabel(),
		Line: ast.Line{
			Position: loop.Position,
			Statements: []ast.Statement{
				&ast.IfStatement{
					Position:  loop.Position,
					Condition: loop.Condition,
					IfBlock:   statements,
				},
			},
		},
	}

	repllist := make([]ast.Node, 1, 2)
	repllist[0] = repl

	if c.getLengthOfLine(&repl.Line) > c.maxLineLength() {
		return nil, errInlineIfImpossible
	}

	if c.getCurrentLoop().HasBreakStatement {
		repllist = append(repllist, &nast.StatementLine{
			Label: c.getCurrentLoop().EndLabel(),
			Line: ast.Line{
				Position:   loop.Position,
				Statements: []ast.Statement{},
			},
		})
	}

	return repllist, nil

}

// convertBreakStatement converts the beak keyword
func (c *Converter) convertBreakStatement(brk *nast.BreakStatement) error {
	if len(c.loopLevel) == 0 {
		return &parser.Error{
			Message:       "The break keyword can only be used inside loops",
			StartPosition: brk.Start(),
			EndPosition:   brk.End(),
		}
	}
	current := c.getCurrentLoop()
	current.HasBreakStatement = true
	return ast.NewNodeReplacementSkip(c.gotoForLabelPos(current.EndLabel(), brk.Position))
}

// convertContinueStatement converts the continue keyword
func (c *Converter) convertContinueStatement(cnt *nast.ContinueStatement) error {
	if len(c.loopLevel) == 0 {
		return &parser.Error{
			Message:       "The continue keyword can only be used inside loops",
			StartPosition: cnt.Start(),
			EndPosition:   cnt.End(),
		}
	}
	return ast.NewNodeReplacementSkip(c.gotoForLabelPos(c.getCurrentLoop().StartLabel(), cnt.Position))
}
