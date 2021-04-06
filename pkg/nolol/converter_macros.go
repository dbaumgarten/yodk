package nolol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// MaxExpandedMacros is the maximum number of macros to expand, before aborting due to a loop
const MaxExpandedMacros = 50

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

// InsertedMacro is a pseudo-ast node that is used during the insertion of a macro
// IT bundles information about an (already expanded) macro and it's insetion
type InsertedMacro struct {
	// Code to be inserted
	Code ast.Node
	// Type of the original macro
	MacroType string
	// The inseriton of the macro
	FuncCall *nast.FuncCall
	// The StatementLine the macro is inserted into (if any)
	ParentStatementLine *nast.StatementLine
	// Position inside ParentStatementLine
	ParentStatementLineIndex int
}

// Start is needed to implement ast.Node
func (e *InsertedMacro) Start() ast.Position {
	return e.Code.Start()
}

// End is needed to implement ast.Node
func (e *InsertedMacro) End() ast.Position {
	return e.Code.End()
}

// Accept is used to implement Acceptor
func (e *InsertedMacro) Accept(v ast.Visitor) error {
	err := v.Visit(e, ast.PreVisit)
	if err != nil {
		return err
	}

	e.Code, err = ast.AcceptChild(v, e.Code)
	if err != nil {
		return err
	}

	err = v.Visit(e, ast.PostVisit)
	if err != nil {
		return err
	}
	return nil
}

// El implements the type-marker method
func (e *InsertedMacro) El() {}

// NestEl implements the type-marker method
func (e *InsertedMacro) NestEl() {}

// Stmt implements the type-marker method
func (e *InsertedMacro) Stmt() {}

// Expr implements the type-marker method
func (e *InsertedMacro) Expr() {}

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

// convert a macro insetion, by replacing it with the corresponting ExpandedMacro
func (c *Converter) convertMacroInsertion(ins *nast.FuncCall, visitType int) error {
	if visitType == ast.PreVisit {
		m, defined := c.getMacro(ins.Function)
		if !defined {
			// This is not a defined macro? We will check later if it is a known function
			return nil
		}

		c.macroLevel = append(c.macroLevel, ins.Function+":"+strconv.Itoa(ins.Start().Line))
		c.macroInsertionCount++

		if c.macroInsertionCount > MaxExpandedMacros {
			return &parser.Error{
				Message:       "Error when processing macros: Macro-loop detected",
				StartPosition: ast.NewPosition("", 1, 1),
				EndPosition:   ast.NewPosition("", 20, 70),
			}
		}

		if len(m.Arguments) != len(ins.Arguments) {
			return &parser.Error{
				Message:       fmt.Sprintf("Wrong number of arguments for %s, got %d but want %d", ins.Function, len(ins.Arguments), len(m.Arguments)),
				StartPosition: ins.Start(),
				EndPosition:   ins.End(),
			}
		}

		if ins.Type == nast.MacroTypeExpr && m.Type != nast.MacroTypeExpr {
			return &parser.Error{
				Message:       fmt.Sprintf("Macro %s has type '%s', but type 'expr' would be required here", m.Name, m.Type),
				StartPosition: ins.Start(),
				EndPosition:   ins.End(),
			}
		}

		if ins.Type == nast.MacroTypeLine && m.Type != nast.MacroTypeLine {
			return &parser.Error{
				Message:       fmt.Sprintf("Macro %s has type '%s', but type 'line' would be required here", m.Name, m.Type),
				StartPosition: ins.Start(),
				EndPosition:   ins.End(),
			}
		}

		if ins.Type != nast.MacroTypeExpr && m.Type == nast.MacroTypeExpr {
			return &parser.Error{
				Message:       fmt.Sprintf("Macro %s has type '%s', but type 'line' or 'block' would be required here", m.Name, m.Type),
				StartPosition: ins.Start(),
				EndPosition:   ins.End(),
			}
		}

		result, err := c.expandMacro(ins, m)
		if err != nil {
			return err
		}

		// Replace the funccall with an InsertedMacro, which will later be replaced with the actual code
		// We do not directly insert the actual code, because we need to get the chance to have a PostVisit on the inserted code
		return ast.NewNodeReplacement(&InsertedMacro{
			Code:                     result,
			MacroType:                m.Type,
			FuncCall:                 ins,
			ParentStatementLine:      c.macroCurrentStatementLine,
			ParentStatementLineIndex: c.macroCurrentStatement,
		})
	}
	return nil
}

// getMacroCodeToInsert uses the macro-definition and the insertion to generate code that is to be inserted into the script
// arguments are replaced with the given values, local variables and labels are renamed and return-statements are processed
func (c *Converter) expandMacro(ins *nast.FuncCall, def *nast.MacroDefinition) (ast.Node, error) {

	copy := nast.CopyAst(def.Code)

	// crate a dict from argument-names to provided values
	arguments := make(map[string]ast.Expression)
	for i := range ins.Arguments {
		lvarname := strings.ToLower(def.Arguments[i])
		arguments[lvarname] = ins.Arguments[i]
	}

	f := func(node ast.Node, visitType int) error {
		// replace the variable name inside assignments
		if ass, is := node.(*ast.Assignment); is && visitType == ast.PreVisit {
			lvarname := strings.ToLower(ass.Variable)
			if replacement, exists := arguments[lvarname]; exists {
				if replacementVariable, isvar := replacement.(*ast.Dereference); isvar && replacementVariable.Operator == "" {
					ass.Variable = replacementVariable.Variable
				} else {
					return &parser.Error{
						Message:       "This argument must be a variable name (and not any other expression)",
						StartPosition: replacement.Start(),
						EndPosition:   replacement.End(),
					}
				}
			} else if !strings.HasPrefix(ass.Variable, ":") && !contains(def.Externals, ass.Variable) {
				if _, isDefinition := c.getDefinition(lvarname); !isDefinition {
					// replace local vars with a insertion-scoped version
					ass.Variable = strings.Join(c.macroLevel, "_") + "_" + strconv.Itoa(c.macroCurrentStatement) + "_" + ass.Variable
				}
			}
		}

		// replace the variable name of dereferences
		if deref, is := node.(*ast.Dereference); is && visitType == ast.SingleVisit {
			lvarname := strings.ToLower(deref.Variable)
			if replacement, exists := arguments[lvarname]; exists {
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
			} else if !strings.HasPrefix(deref.Variable, ":") && !contains(def.Externals, deref.Variable) {
				if _, isDefinition := c.getDefinition(lvarname); !isDefinition {
					// replace local vars with a insertion-scoped version
					deref.Variable = strings.Join(c.macroLevel, "_") + "_" + strconv.Itoa(c.macroCurrentStatement) + "_" + deref.Variable
				}
			}
		}

		// rename macro-internal jump-labels
		if stmtline, is := node.(*nast.StatementLine); is && visitType == ast.PreVisit {
			if stmtline.Label != "" {
				if !contains(def.Externals, stmtline.Label) {
					stmtline.Label = strings.Join(c.macroLevel, "_") + "_" + strconv.Itoa(c.macroCurrentStatement) + "_" + stmtline.Label
				}
				c.storeLineLabel(stmtline.Label, -1)
			}
		}

		return nil
	}

	return copy, copy.Accept(ast.VisitorFunc(f))
}

// convert a macro insetion, by replacing it with the corresponting ExpandedMacro
func (c *Converter) convertInsertedMacro(em *InsertedMacro, visitType int) error {
	if visitType == ast.PostVisit {

		// pop one level
		c.macroLevel = c.macroLevel[:len(c.macroLevel)-1]

		if em.FuncCall.Type == nast.MacroTypeBlock {
			if em.MacroType == nast.MacroTypeBlock {
				block := em.Code.(*nast.Block)
				nodes := make([]ast.Node, len(block.Elements))
				for i, el := range block.Elements {
					nodes[i] = el
				}
				return ast.NewNodeReplacement(nodes...)
			}
			if em.MacroType == nast.MacroTypeLine {
				ast.NewNodeReplacement(em.Code.(*nast.StatementLine))
			}
		}

		if em.FuncCall.Type == nast.MacroTypeLine {
			line := em.Code.(*nast.StatementLine)
			nodes := make([]ast.Node, len(line.Statements))
			for i, el := range line.Statements {
				nodes[i] = el
			}
			err := c.mergeInsertedLinemacro(em)
			if err != nil {
				return err
			}
			return ast.NewNodeReplacement(nodes...)
		}

		return ast.NewNodeReplacement(em.Code)
	}
	return nil
}

func (c *Converter) mergeInsertedLinemacro(em *InsertedMacro) error {
	inserted := em.Code.(*nast.StatementLine)
	existing := em.ParentStatementLine
	fc := em.ParentStatementLine
	index := em.ParentStatementLineIndex

	if inserted.Label != "" {
		if existing.Label != "" {
			return &parser.Error{
				Message:       fmt.Sprint("Cannot insert this macro here. Target-line already has a line-label"),
				StartPosition: fc.Start(),
				EndPosition:   fc.End(),
			}
		}
		if index != 0 {
			return &parser.Error{
				Message:       fmt.Sprint("Cannot insert this macro here. Inserted line has a linelabel, but is not inserted at the first position"),
				StartPosition: fc.Start(),
				EndPosition:   fc.End(),
			}
		}
		existing.Label = inserted.Label
	}
	if inserted.HasBOL {
		if index != 0 {
			return &parser.Error{
				Message:       fmt.Sprint("Cannot insert this macro here. Inserted line has BOL-Marker, but is not inserted at the first position"),
				StartPosition: fc.Start(),
				EndPosition:   fc.End(),
			}
		}
		existing.HasBOL = true
	}
	if inserted.HasEOL {
		if index != len(existing.Statements)-1 {
			return &parser.Error{
				Message:       fmt.Sprint("Cannot insert this macro here. Inserted line has EOL-Marker, but is not inserted at the last position"),
				StartPosition: fc.Start(),
				EndPosition:   fc.End(),
			}
		}
		existing.HasEOL = true
	}
	return nil
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
