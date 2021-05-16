package nolol

import (
	"fmt"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

func (c *Converter) optimizeExpression(exp ast.Expression) ast.Node {
	repl := c.boolexpOptimizer.OptimizeExpression(exp)
	if repl != nil {
		exp = repl
	}
	repl = c.sexpOptimizer.OptimizeExpressionNonRecursive(exp)
	if repl != nil {
		exp = repl
	}
	return exp
}

func (c *Converter) convertAssignment(n *ast.Assignment, visitType int) error {
	if visitType == ast.PreVisit {
		if _, isLineLabel := c.getLineLabel(n.Variable); isLineLabel {
			return &parser.Error{
				Message:       "Can not assign to a line-label",
				StartPosition: n.Start(),
				EndPosition:   n.End(),
			}
		}
		n.Variable = c.varnameOptimizer.OptimizeVarName(n.Variable)
	}
	return nil
}

func (c *Converter) convertDereference(n *ast.Dereference) error {
	if _, isLineLabel := c.getLineLabel(n.Variable); isLineLabel {
		// dereference of line-label
		if n.Operator != "" {
			return &parser.Error{
				Message:       "Can not Pre/Post-Operate on line-label",
				StartPosition: n.Start(),
				EndPosition:   n.End(),
			}
		}
		return nil
	}
	n.Variable = c.varnameOptimizer.OptimizeVarName(n.Variable)
	return nil
}

// mergeNololNestableElements is a type-wrapper for mergeStatementElements
func (c *Converter) mergeNololNestableElements(lines []nast.NestableElement) ([]nast.NestableElement, error) {
	inp := make([]*nast.StatementLine, len(lines))
	for i, elem := range lines {
		line, isline := elem.(*nast.StatementLine)
		if !isline {
			return nil, parser.Error{
				Message: fmt.Sprintf("Err: Found unconverted nolol-element: %T", elem),
			}
		}
		inp[i] = line
	}
	interm, err := c.mergeStatementElements(inp)
	if err != nil {
		return nil, err
	}
	outp := make([]nast.NestableElement, len(interm))
	for i, elem := range interm {
		outp[i] = elem
	}
	return outp, nil
}

// mergeNololElements is a type-wrapper for mergeStatementElements
func (c *Converter) mergeNololElements(lines []nast.Element) ([]nast.Element, error) {
	inp := make([]*nast.StatementLine, len(lines))
	for i, elem := range lines {
		line, isline := elem.(*nast.StatementLine)
		if !isline {
			return nil, parser.Error{
				Message: fmt.Sprintf("Err: Found unconverted nolol-element: %T", elem),
			}
		}
		inp[i] = line
	}
	interm, err := c.mergeStatementElements(inp)
	if err != nil {
		return nil, err
	}
	outp := make([]nast.Element, len(interm))
	for i, elem := range interm {
		outp[i] = elem
	}
	return outp, nil
}

// mergeStatementElements merges consectuive statementlines into as few lines as possible
func (c *Converter) mergeStatementElements(lines []*nast.StatementLine) ([]*nast.StatementLine, error) {
	maxlen := c.maxLineLength()
	newElements := make([]*nast.StatementLine, 0, len(lines))
	i := 0
	for i < len(lines) {
		current := &nast.StatementLine{
			Line: ast.Line{
				Statements: []ast.Statement{},
				Position:   lines[i].Position,
			},
			Label:  lines[i].Label,
			HasEOL: lines[i].HasEOL,
		}
		current.Statements = append(current.Statements, lines[i].Statements...)
		newElements = append(newElements, current)

		if current.HasEOL {
			// no lines MUST be appended to a line having EOL
			i++
			continue
		}

		for i+1 < len(lines) {
			currlen := c.getLengthOfLine(&current.Line)

			if currlen > maxlen {
				return newElements, &parser.Error{
					Message:       "The line is too long (>70 characters) to be converted to yolol, even after optimization.",
					StartPosition: current.Start(),
					EndPosition:   current.End(),
				}
			}

			nextline := lines[i+1]

			if nextline.Label == "" && !nextline.HasBOL {
				prev := current.Statements
				current.Statements = make([]ast.Statement, 0, len(current.Statements)+len(nextline.Statements))
				current.Statements = append(current.Statements, prev...)
				current.Statements = append(current.Statements, nextline.Statements...)

				newlen := c.getLengthOfLine(&current.Line)
				if newlen > maxlen {
					// the newly created line is longer then allowed. roll back.
					current.Statements = prev
					break
				}

				i++
				if nextline.HasEOL {
					break
				}
			} else {
				break
			}
		}
		i++
	}
	return newElements, nil
}

//getLengthOfLine returns the amount of characters needed to represent the given line as yolol-code
func (c *Converter) getLengthOfLine(line ast.Node) int {
	ygen := parser.Printer{}
	ygen.Mode = parser.PrintermodeCompact

	silenceGotoExpression := false
	ygen.PrinterExtensionFunc = func(node ast.Node, visitType int, p *parser.Printer) (bool, error) {
		if gotostmt, is := node.(*ast.GoToStatement); is {
			if c.getGotoDestinationLabel(gotostmt) != "" {
				if visitType == ast.PreVisit {
					silenceGotoExpression = true
					p.Write("gotoXX")
				}
				if visitType == ast.PostVisit {
					silenceGotoExpression = false
				}
				return true, nil
			}
		}
		if silenceGotoExpression {
			if _, is := node.(ast.Expression); is {
				// The current expression is inside a goto.
				// DO NOT PRINT IT
				return true, nil
			}
		}
		return false, nil
	}
	generated, err := ygen.Print(line)
	if err != nil {
		panic(err)
	}

	linelen := len(generated)
	if strings.HasSuffix(generated, "\n") {
		linelen--
	}

	return linelen
}

func (c *Converter) maxLineLength() int {
	if !c.usesTimeTracking {
		return 70
	}
	return 70 - 4
}
