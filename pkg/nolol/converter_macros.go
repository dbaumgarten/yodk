package nolol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// getMacro is a case-insensitive getter for c.macros
func (c *Converter) getMacro(name string) (*nast.MacroDefinition, bool) {
	name = strings.ToLower(name)
	val, exists := c.macros[name]
	return val, exists
}

// setMacro is a case-insensitive setter for c.macros
func (c *Converter) setMacro(name string, val *nast.MacroDefinition) {
	name = strings.ToLower(name)
	c.macros[name] = val
}

// convertMacroDef takes a macro definition, stores it for later use and removes the definition from the code
func (c *Converter) convertMacroDef(def *nast.MacroDefinition, visitType int) error {
	// using pre-visit here is important
	// the definition must be resolved, BEFORE its contents are processed
	if visitType == ast.PreVisit {
		c.setMacro(def.Name, def)
		// remove the node from the output-code
		return ast.NewNodeReplacementSkip()
	}
	return nil
}

// convert a macro insetion, by inserting the code defined by the macro
func (c *Converter) convertMacroInsertion(ins *nast.MacroInsetion, visitType int) error {
	if visitType != ast.PreVisit {
		return nil
	}

	c.macroLevel = append(c.macroLevel, ins.Function+":"+strconv.Itoa(ins.Start().Line))

	c.macroInsertionCount++

	if c.macroInsertionCount > 20 {
		return &parser.Error{
			Message:       "Error when processing macros: Macro-loop detected",
			StartPosition: ast.NewPosition("", 1, 1),
			EndPosition:   ast.NewPosition("", 20, 70),
		}
	}

	m, defined := c.getMacro(ins.Function)
	if !defined {
		return &parser.Error{
			Message:       fmt.Sprintf("No macro named '%s' defined", ins.Function),
			StartPosition: ins.Start(),
			EndPosition:   ins.End(),
		}
	}

	if len(m.Arguments) != len(ins.Arguments) {
		return &parser.Error{
			Message:       fmt.Sprintf("Wrong number of arguments for %s, got %d but want %d", ins.Function, len(ins.Arguments), len(m.Arguments)),
			StartPosition: ins.Start(),
			EndPosition:   ins.End(),
		}
	}

	// gather replacements
	replacements := make(map[string]ast.Expression)
	for i := range ins.Arguments {
		lvarname := strings.ToLower(m.Arguments[i])
		replacements[lvarname] = ins.Arguments[i]
	}

	copy := nast.CopyAst(m).(*nast.MacroDefinition)

	err := c.replacePlaceholders(copy, replacements, m.Externals, true)
	if err != nil {
		return err
	}

	nodes := make([]ast.Node, len(copy.Block.Elements)+1)
	for i, el := range copy.Block.Elements {
		nodes[i] = el
	}
	nodes[len(nodes)-1] = &nast.Trigger{
		Kind: "macroleft",
	}

	// remove the node from the output-code
	return ast.NewNodeReplacement(nodes...)
}

// replacePlaceholders replaces all dereferences of and assignments to placeholders in a sub-ast with the given replacements
// if aliasRemaining is true, all not-replaced non-global vars, that are not listed in ignore, will be given new unique names (=are made local to the sub-ast)
func (c *Converter) replacePlaceholders(m ast.Node, replacements map[string]ast.Expression, ignore []string, aliasRemaining bool) error {
	f := func(node ast.Node, visitType int) error {
		// replace the variable name inside assignments
		if ass, is := node.(*ast.Assignment); is && visitType == ast.PreVisit {
			lvarname := strings.ToLower(ass.Variable)
			if replacement, exists := replacements[lvarname]; exists {
				if replacementVariable, isvar := replacement.(*ast.Dereference); isvar && replacementVariable.Operator == "" {
					ass.Variable = replacementVariable.Variable
				} else {
					return &parser.Error{
						Message:       "This argument must be a variable name (and not any other expression)",
						StartPosition: replacement.Start(),
						EndPosition:   replacement.End(),
					}
				}
			} else if aliasRemaining && !strings.HasPrefix(ass.Variable, ":") && !contains(ignore, ass.Variable) {
				if _, isDefinition := c.getDefinition(lvarname); !isDefinition {
					// replace local vars with a insertion-scoped version
					ass.Variable = strings.Join(c.macroLevel, "_") + "_" + ass.Variable
				}
			}
		}

		// replace the variable name of dereferences
		if deref, is := node.(*ast.Dereference); is && visitType == ast.SingleVisit {
			lvarname := strings.ToLower(deref.Variable)
			if replacement, exists := replacements[lvarname]; exists {
				replacement = nast.CopyAst(replacement).(ast.Expression)
				if replacementVariable, isvar := replacement.(*ast.Dereference); isvar {
					if deref.Operator != "" && replacementVariable.Operator != "" {
						return &parser.Error{
							Message:       "You can not use pre/post-operators for this particular argument",
							StartPosition: replacement.Start(),
							EndPosition:   replacement.End(),
						}
					}
					if deref.Operator != "" {
						replacementVariable.Operator = deref.Operator
						replacementVariable.PrePost = deref.PrePost
					}
				} else if deref.Operator != "" {
					return &parser.Error{
						Message:       "This argument must be a variable name (and not any other expression)",
						StartPosition: replacement.Start(),
						EndPosition:   replacement.End(),
					}
				}
				return ast.NewNodeReplacementSkip(replacement)
			} else if aliasRemaining && !strings.HasPrefix(deref.Variable, ":") && !contains(ignore, deref.Variable) {
				if _, isDefinition := c.getDefinition(lvarname); !isDefinition {
					// replace local vars with a insertion-scoped version
					deref.Variable = strings.Join(c.macroLevel, "_") + "_" + deref.Variable
				}
			}
		}
		return nil
	}

	return m.Accept(ast.VisitorFunc(f))
}

func contains(arr []string, s string) bool {
	if arr == nil {
		return false
	}
	for _, el := range arr {
		if el == s {
			return true
		}
	}
	return false
}
