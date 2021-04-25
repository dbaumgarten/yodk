package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/number"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
	"github.com/dbaumgarten/yodk/pkg/vm"
)

// getCurrentLoopNumber returns the number of the current (innermost) loop that is beeing converted
func (c *Converter) getCurrentLoopNumber() int {
	return c.loopLevel[len(c.loopLevel)-1]
}

// convertWhileLoop converts while loops into yolol-code
func (c *Converter) convertWhileLoop(loop *nast.WhileLoop, visitType int) error {

	if visitType == ast.PreVisit {
		c.loopcounter++
		c.loopLevel = append(c.loopLevel, c.loopcounter)
		return nil
	}

	if visitType != ast.PostVisit {
		return nil
	}

	defer func() {
		c.loopLevel = c.loopLevel[:len(c.loopLevel)-1]
	}()

	loopnr := c.getCurrentLoopNumber()
	startLabel := fmt.Sprintf("while%d", loopnr)
	endLabel := fmt.Sprintf("endwhile%d", loopnr)

	inlineloop, err := c.convertWhileLoopInline(loop, startLabel)
	if err == nil {
		return ast.NewNodeReplacementSkip(inlineloop)
	}

	repl := []ast.Node{
		&nast.StatementLine{
			Label: startLabel,
			Line: ast.Line{
				Position:   loop.Position,
				Statements: []ast.Statement{},
			},
		},
	}

	condition := c.sexpOptimizer.OptimizeExpression(loop.Condition)

	conditionIsAlwaysTrue := false
	if numberconst, is := condition.(*ast.NumberConstant); is {
		variable := vm.VariableFromString(numberconst.Value)
		if variable.Number() != number.Zero {
			conditionIsAlwaysTrue = true
		}
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
					c.gotoForLabel(endLabel),
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
				c.gotoForLabel(startLabel),
			},
		},
	})

	repl = append(repl, &nast.StatementLine{
		Label: endLabel,
		Line: ast.Line{
			Position:   loop.Position,
			Statements: []ast.Statement{},
		},
	})

	return ast.NewNodeReplacementSkip(repl...)
}

func (c *Converter) convertWhileLoopInline(loop *nast.WhileLoop, looplabel string) (ast.Node, error) {
	mergedIfElements, _ := c.mergeNololNestableElements(loop.Block.Elements)

	if len(mergedIfElements) > 1 || (len(mergedIfElements) > 0 && mergedIfElements[0].(*nast.StatementLine).Label != "") {
		return nil, errInlineIfImpossible
	}

	statements := []ast.Statement{}
	if len(mergedIfElements) > 0 {
		statements = mergedIfElements[0].(*nast.StatementLine).Line.Statements
	}
	statements = append(statements, c.gotoForLabel(looplabel))

	repl := &nast.StatementLine{
		Label: looplabel,
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

	if c.getLengthOfLine(&repl.Line) > c.maxLineLength() {
		return nil, errInlineIfImpossible
	}

	return repl, nil
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
	endLabel := fmt.Sprintf("endwhile%d", c.getCurrentLoopNumber())
	return ast.NewNodeReplacementSkip(c.gotoForLabelPos(endLabel, brk.Position))
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
	startLabel := fmt.Sprintf("while%d", c.getCurrentLoopNumber())
	return ast.NewNodeReplacementSkip(c.gotoForLabelPos(startLabel, cnt.Position))
}
