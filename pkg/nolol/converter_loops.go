package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// getCurrentLoopNumber returns the number of the current (innermost) loop that is beeing converted
func (c *Converter) getCurrentLoopNumber() int {
	return c.loopLevel[len(c.loopLevel)-1]
}

// convertWhileLoop converts while loops into yolol-code
func (c *Converter) convertWhileLoop(loop *nast.WhileLoop) error {
	loopnr := c.getCurrentLoopNumber()
	startLabel := fmt.Sprintf("while%d", loopnr)
	endLabel := fmt.Sprintf("endwhile%d", loopnr)

	condition := c.boolexpOptimizer.OptimizeExpression(&ast.UnaryOperation{
		Operator: "not",
		Exp:      loop.Condition,
		Position: loop.Condition.Start(),
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
		Position: loop.Block.End(),
		Line: ast.Line{
			Statements: []ast.Statement{
				&nast.GoToLabelStatement{
					Position: loop.Block.End(),
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

	return ast.NewNodeReplacementSkip(repl...)

}

// convertBreakStatement converts the rbeak keyword
func (c *Converter) convertBreakStatement(brk *nast.BreakStatement) error {
	if len(c.loopLevel) == 0 {
		return &parser.Error{
			Message:       "The break keyword can only be used inside loops",
			StartPosition: brk.Start(),
			EndPosition:   brk.End(),
		}
	}
	endLabel := fmt.Sprintf("endwhile%d", c.getCurrentLoopNumber())
	return ast.NewNodeReplacementSkip(&nast.GoToLabelStatement{
		Position: brk.Position,
		Label:    endLabel,
	})
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
	return ast.NewNodeReplacementSkip(&nast.GoToLabelStatement{
		Position: cnt.Position,
		Label:    startLabel,
	})
}
