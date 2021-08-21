package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/number"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
	"github.com/dbaumgarten/yodk/pkg/vm"
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

func (c *Converter) eliminateDeadCode(n ast.Node) error {
	f := func(node ast.Node, visitType int) error {
		if visitType == ast.PostVisit {
			if mlif, is := node.(*nast.MultilineIf); is {
				return c.removeDeadBranches(mlif)
			}
		}
		return nil
	}
	return n.Accept(ast.VisitorFunc(f))
}

// Returns true if the given expression is a static value (and returns that value)
func (c Converter) isStaticValue(e ast.Expression) (bool, *vm.Variable) {
	condition := c.sexpOptimizer.OptimizeExpression(e)
	if numberconst, is := condition.(*ast.NumberConstant); is {
		return true, vm.VariableFromString(numberconst.Value)
	}
	return false, nil
}

func (c *Converter) removeDeadBranches(mlif *nast.MultilineIf) error {
	replacement := &nast.MultilineIf{
		Positions:  make([]ast.Position, 0, len(mlif.Positions)),
		Conditions: make([]ast.Expression, 0, len(mlif.Conditions)),
		Blocks:     make([]*nast.Block, 0, len(mlif.Blocks)),
		ElseBlock:  mlif.ElseBlock,
	}
	foundAlwaysTrue := false
	for i, cond := range mlif.Conditions {
		isstatic, value := c.isStaticValue(cond)
		if isstatic && value.IsNumber() {
			if value.Number() == number.Zero {
				// condition is always false. Do not copy this to replacement
			} else {
				// always true condition.
				foundAlwaysTrue = true
				if len(replacement.Conditions) == 0 {
					// this is the first (and now only) condition.
					replacement.Positions = append(replacement.Positions, mlif.Positions[i])
					replacement.Conditions = append(replacement.Conditions, mlif.Conditions[i])
					replacement.Blocks = append(replacement.Blocks, mlif.Blocks[i])
					replacement.ElseBlock = nil
				} else {
					// there is a condition before this
					replacement.ElseBlock = mlif.Blocks[i]
				}
				break
			}
		} else {
			// not a static condition. Copy over to replacement and continue
			replacement.Positions = append(replacement.Positions, mlif.Positions[i])
			replacement.Conditions = append(replacement.Conditions, mlif.Conditions[i])
			replacement.Blocks = append(replacement.Blocks, mlif.Blocks[i])
		}
	}

	if len(replacement.Conditions) == 0 {
		if replacement.ElseBlock == nil {
			return ast.NewNodeReplacementSkip()
		} else {
			return ast.NewNodeReplacementSkip(blockToNodelist(replacement.ElseBlock)...)
		}
	} else if len(replacement.Conditions) == 1 && foundAlwaysTrue && replacement.ElseBlock == nil {
		return ast.NewNodeReplacementSkip(blockToNodelist(replacement.Blocks[0])...)
	}

	return ast.NewNodeReplacementSkip(replacement)
}

func blockToNodelist(block *nast.Block) []ast.Node {
	nodes := make([]ast.Node, len(block.Elements))
	for i, el := range block.Elements {
		nodes[i] = el
	}
	return nodes
}
