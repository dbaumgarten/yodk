package nolol

import (
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// getDefinition is a case-insensitive getter for c.definitions
func (c *Converter) getDefinition(name string) (ast.Expression, bool) {
	name = strings.ToLower(name)
	val, exists := c.definitions[name]
	return val, exists
}

// setDefinition is a case-insensitive setter for c.definitions
func (c *Converter) setDefinition(name string, val ast.Expression) {
	name = strings.ToLower(name)
	c.definitions[name] = val
}

// convertDefinitions converts a definition to yolol by discarding it, but saving the defined value
func (c *Converter) convertDefinition(decl *nast.Definition) error {
	c.setDefinition(decl.Name, decl.Value)
	return ast.NewNodeReplacement()
}

// convertAssignment optimizes the variable name and the expression of an assignment
func (c *Converter) convertAssignment(ass *ast.Assignment) error {
	if replacement, exists := c.getDefinition(ass.Variable); exists {
		if replacementVariable, isvar := replacement.(*ast.Dereference); isvar && replacementVariable.Operator == "" {
			ass.Variable = replacementVariable.Variable
		} else {
			return &parser.Error{
				Message:       "Can not assign to a definition that is an expression (need a single variable name)",
				StartPosition: ass.Start(),
				EndPosition:   ass.End(),
			}
		}
	} else {
		ass.Variable = c.varnameOptimizer.OptimizeVarName(ass.Variable)
	}
	return nil
}

// convertDereference replaces mentionings of constants with the value of the constant
func (c *Converter) convertDereference(deref *ast.Dereference) error {
	if replacement, exists := c.getDefinition(deref.Variable); exists {
		replacement = nast.CopyAst(replacement)
		if replacementVariable, isvar := replacement.(*ast.Dereference); isvar {
			if deref.Operator != "" && replacementVariable.Operator != "" {
				return &parser.Error{
					Message:       "You can not use pre/post-operators on defitions that use the operator themselves",
					StartPosition: deref.Start(),
					EndPosition:   deref.End(),
				}
			}
			if deref.Operator != "" {
				replacementVariable.Operator = deref.Operator
				replacementVariable.PrePost = deref.PrePost
			}
		} else if deref.Operator != "" {
			return &parser.Error{
				Message:       "Can not use pre/port-operators on expressions",
				StartPosition: deref.Start(),
				EndPosition:   deref.End(),
			}
		}
		return ast.NewNodeReplacementSkip(replacement)
	}
	// we are dereferencing a variable
	deref.Variable = c.varnameOptimizer.OptimizeVarName(deref.Variable)
	return nil
}
