package nolol

import (
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Parser parses a nolol-program
type Parser struct {
	*parser.Parser
}

// PublicParser is an interface that hides all overridable-methods from normal users
type PublicParser interface {
	Parse(prog string) (*nast.Program, error)
	Debug(b bool)
}

// Constant definitions for parser.Error.Code
const (
	ErrExpectedStringConstant = "ErrExpectedStringConstant"
	ErrExpectedIdentifier     = "ErrExpectedIdentifier"
	ErrExpectedExistingMacro  = "ErrExpectedExistingMacro"
	ErrExpectedJumplabel      = "ErrExpectedJumplabel"
	ErrExpectedMacroType      = "ErrExpectedMacroType"
)

// NewParser creates and returns a nolol parser
func NewParser() PublicParser {
	ep := &Parser{
		Parser: parser.NewParser().(*parser.Parser),
	}
	ep.This = ep
	ep.Tokenizer = nast.NewNololTokenizer()
	return ep
}

// Debug enables/disables debug logging
func (p *Parser) Debug(b bool) {
	p.DebugLog = b
}

// SetFilename sets the filename that is included in the position of every returned ast.node
// Necessary when parsing an included file to differenciate between positions in different files
func (p *Parser) SetFilename(name string) {
	p.Tokenizer.SetFilename(name)
}

// Parse is the entry point for parsing
func (p *Parser) Parse(prog string) (*nast.Program, error) {
	p.Reset()
	p.Tokenizer.Load(prog)
	// Advance to fill CurrentToken
	p.Advance()
	parsed := p.ParseProgram()

	if len(p.Errors) != 0 {
		return nil, p.Errors
	}

	validationErrors := parser.Validate(parsed, parser.ValidateAll^parser.ValidateLocalVars)
	p.Errors = append(p.Errors, validationErrors...)
	if len(p.Errors) != 0 {
		return nil, p.Errors
	}

	parser.RemoveParenthesis(parsed)

	return parsed, nil
}

// ParseProgram parses the program
func (p *Parser) ParseProgram() *nast.Program {
	p.Log()
	ret := nast.Program{
		Elements: make([]nast.Element, 0),
	}
	for p.HasNext() {
		ret.Elements = append(ret.Elements, p.ParseElement())
	}
	return &ret
}

// ParseNestableElement parses a NOLOL-Element which can appear inside a blocl
func (p *Parser) ParseNestableElement() nast.NestableElement {
	p.Log()

	ifline := p.ParseMultilineIf()
	if ifline != nil {
		return ifline
	}

	whileline := p.ParseWhile()
	if whileline != nil {
		return whileline
	}

	fcall := p.ParseNestableElementFuncCall()
	if fcall != nil {
		return fcall
	}

	return p.ParseStatementLine()
}

// ParseElement parses an element
func (p *Parser) ParseElement() nast.Element {
	p.Log()

	include := p.ParseInclude()
	if include != nil {
		return include
	}

	constDecl := p.ParseDefinition()
	if constDecl != nil {
		return constDecl
	}

	mDef := p.ParseMacroDefinition()
	if mDef != nil {
		return mDef
	}

	// NestableElements are also elements
	return p.ParseNestableElement()
}

// ParseInclude parses an include directive
func (p *Parser) ParseInclude() *nast.IncludeDirective {
	p.Log()

	if !p.IsCurrent(ast.TypeKeyword, "include") {
		return nil
	}
	incl := &nast.IncludeDirective{
		Position: p.CurrentToken.Position,
	}
	p.Advance()
	if !p.IsCurrentType(ast.TypeString) {
		p.ErrorString("Expected a string-constant after include", ErrExpectedStringConstant)
		return incl
	}
	incl.File = p.CurrentToken.Value
	p.Advance()
	if !p.IsCurrentType(ast.TypeEOF) {
		p.Expect(ast.TypeNewline, "")
	}
	return incl
}

// ParseMacroDefinition parses the definition of a macro
func (p *Parser) ParseMacroDefinition() *nast.MacroDefinition {
	if !p.IsCurrent(ast.TypeKeyword, "macro") {
		return nil
	}
	p.Advance()
	mdef := &nast.MacroDefinition{
		Position:  p.CurrentToken.Position,
		Arguments: []string{},
		Externals: []string{},
	}
	if !p.IsCurrentType(ast.TypeID) {
		p.ErrorString("Expected an identifier after the macro keyword", ErrExpectedIdentifier)
		return mdef
	}
	mdef.Name = p.CurrentToken.Value
	p.Advance()

	p.Expect(ast.TypeSymbol, "(")
	for !p.IsCurrent(ast.TypeSymbol, ")") {
		if !p.IsCurrentType(ast.TypeID) {
			p.ErrorString("Only comma separated identifiers are allowed as arguments in a macro definition", ErrExpectedIdentifier)
			break
		}
		mdef.Arguments = append(mdef.Arguments, p.CurrentToken.Value)
		p.Advance()
		if p.IsCurrent(ast.TypeSymbol, ",") {
			p.Advance()
			continue
		}
		break
	}
	p.Expect(ast.TypeSymbol, ")")

	if p.IsCurrent(ast.TypeSymbol, "<") {
		p.Advance()
		for !p.IsCurrent(ast.TypeSymbol, ">") {
			if !p.IsCurrentType(ast.TypeID) {
				p.ErrorString("Only comma separated identifiers are allowed as globals in a macro definition", ErrExpectedIdentifier)
				break
			}
			mdef.Externals = append(mdef.Externals, p.CurrentToken.Value)
			p.Advance()
			if p.IsCurrent(ast.TypeSymbol, ",") {
				p.Advance()
				continue
			}
			break
		}
		p.Expect(ast.TypeSymbol, ">")
	}

	if !p.IsCurrentType(ast.TypeKeyword) || !p.IsCurrentValueIn([]string{nast.MacroTypeBlock, nast.MacroTypeLine, nast.MacroTypeExpr}) {
		p.ErrorString("Expected macro-type definiton ('block', 'line' or 'expr')", ErrExpectedMacroType)
		return nil
	}

	mdef.Type = p.CurrentToken.Value
	p.Advance()

	p.Expect(ast.TypeNewline, "")

	if mdef.Type != nast.MacroTypeBlock {
		mdef.PreComments = make([]string, 0)
		for {
			if p.IsCurrentType(ast.TypeComment) {
				mdef.PreComments = append(mdef.PreComments, p.CurrentToken.Value)
				p.Advance()
				if p.IsCurrentType(ast.TypeNewline) {
					p.Advance()
				}
			} else if p.IsCurrentType(ast.TypeNewline) {
				mdef.PreComments = append(mdef.PreComments, "")
				p.Advance()
			} else {
				break
			}
		}
	}

	if mdef.Type == nast.MacroTypeBlock {
		mdef.Code = p.ParseBlock(func() bool {
			return p.IsCurrent(ast.TypeKeyword, "end")
		})
	} else if mdef.Type == nast.MacroTypeLine {
		mdef.Code = p.ParseStatementLine()
		if mdef.Code == nil {
			p.ErrorString("Expected a (single) line of statements", "")
		}
	} else if mdef.Type == nast.MacroTypeExpr {
		mdef.Code = p.ParseExpression()
		if mdef.Code == nil {
			p.ErrorExpectedExpression("inside macro of type expr")
		}
		p.Expect(ast.TypeNewline, "")
	}

	if mdef.Type != nast.MacroTypeBlock {
		mdef.PostComments = make([]string, 0)
		for {
			if p.IsCurrentType(ast.TypeComment) {
				mdef.PostComments = append(mdef.PostComments, p.CurrentToken.Value)
				p.Advance()
				if p.IsCurrentType(ast.TypeNewline) {
					p.Advance()
				}
			} else if p.IsCurrentType(ast.TypeNewline) {
				mdef.PostComments = append(mdef.PostComments, "")
				p.Advance()
			} else {
				break
			}
		}
	}

	p.Expect(ast.TypeKeyword, "end")

	return mdef
}

// ParseStatementLine parses a statement line
func (p *Parser) ParseStatementLine() *nast.StatementLine {
	p.Log()
	ret := nast.StatementLine{
		Line: ast.Line{
			Position:   p.CurrentToken.Position,
			Statements: make([]ast.Statement, 0, 1),
		},
	}

	// get line-label if it exists
	nextToken := p.Tokenizer.Peek()
	if p.IsCurrentType(ast.TypeID) && (nextToken.Type == ast.TypeSymbol && nextToken.Value == ">") {
		ret.Label = strings.ToLower(p.CurrentToken.Value)
		p.Advance()
		p.Advance()
	}

	// this line has no statements, only a comment
	if p.IsCurrentType(ast.TypeComment) {
		ret.Comment = p.CurrentToken.Value
		p.Advance()
	}

	if p.IsCurrent(ast.TypeSymbol, "$") {
		ret.HasBOL = true
		p.Advance()
	}

	// the line has no statements
	if p.IsCurrentType(ast.TypeEOF) || p.IsCurrentType(ast.TypeNewline) || p.IsCurrentType(ast.TypeComment) {
		if p.IsCurrentType(ast.TypeComment) {
			ret.Comment = p.CurrentToken.Value
		}
		p.Advance()
		// if a line has no statements, its BOL is also its EOL
		ret.HasEOL = ret.HasBOL
		return &ret
	}

	stmt := p.This.ParseStatement()
	// at this point, the line must at least have one statement
	if stmt != nil {
		ret.Statements = append(ret.Statements, stmt)
	} else {
		p.ErrorExpectedStatement("")
		p.Advance()
		return &ret
	}

	for p.IsCurrent(ast.TypeSymbol, ";") {
		p.Advance()
		stmt = p.This.ParseStatement()
		if stmt != nil {
			ret.Statements = append(ret.Statements, stmt)
		} else {
			p.ErrorExpectedStatement(("after ';'"))
		}
	}

	if p.IsCurrent(ast.TypeSymbol, "$") {
		ret.HasEOL = true
		p.Advance()
	}

	// This line has statements and a comment at the end
	if p.IsCurrentType(ast.TypeComment) {
		ret.Comment = p.CurrentToken.Value
		p.Advance()
	}

	if !p.IsCurrentType(ast.TypeEOF) {
		p.Expect(ast.TypeNewline, "")
	}

	return &ret
}

// ParseDefinition parses a constant declaration
func (p *Parser) ParseDefinition() *nast.Definition {
	p.Log()
	if !p.IsCurrent(ast.TypeKeyword, "define") {
		return nil
	}
	startpos := p.CurrentToken.Position
	p.Advance()
	if !p.IsCurrentType(ast.TypeID) {
		p.ErrorString("const keyword must be followed by an identifier", ErrExpectedIdentifier)
	}
	decl := &nast.Definition{
		Name:     p.CurrentToken.Value,
		Position: startpos,
	}
	p.Advance()

	p.Expect(ast.TypeSymbol, "=")
	value := p.ParseExpression()
	if value == nil {
		p.ErrorExpectedExpression("after the '=' of a definition")
	}
	decl.Value = value
	if !p.IsCurrentType(ast.TypeEOF) {
		p.Expect(ast.TypeNewline, "")
	}
	return decl
}

// ParseMultilineIf parses a nolol-style multiline if
func (p *Parser) ParseMultilineIf() nast.NestableElement {
	p.Log()

	// We can not be absolutely sure that this is really a multiline if
	// Backup the parser-state, just in case
	savedToken := p.CurrentToken
	tokenizerCheckpoint := p.Tokenizer.Checkpoint()

	mlif := nast.MultilineIf{
		Positions:  make([]ast.Position, 1),
		Conditions: make([]ast.Expression, 0),
		Blocks:     make([]*nast.Block, 0),
	}
	mlif.Positions[0] = p.CurrentToken.Position
	if !p.IsCurrent(ast.TypeKeyword, "if") {
		return nil
	}
	p.Advance()

	for {
		condition := p.This.ParseExpression()
		if condition == nil {
			p.ErrorExpectedExpression("as if-condition")
			p.Advance()
		}

		p.Expect(ast.TypeKeyword, "then")

		if p.IsCurrentType(ast.TypeNewline) {
			p.Advance()
		} else {
			// We fucked up, this is not a multiline if. Restore saved state and return
			p.CurrentToken = savedToken
			p.Tokenizer.Restore(tokenizerCheckpoint)
			return nil
		}

		block := p.ParseBlock(func() bool {
			return p.IsCurrentType(ast.TypeKeyword) && (p.IsCurrentValue("end") || p.IsCurrentValue("else"))
		})
		mlif.Conditions = append(mlif.Conditions, condition)
		mlif.Blocks = append(mlif.Blocks, block)

		if p.IsCurrent(ast.TypeKeyword, "end") {
			break
		}

		if p.IsCurrent(ast.TypeKeyword, "else") {
			p.Advance()
		}

		if p.IsCurrent(ast.TypeKeyword, "if") {
			mlif.Positions = append(mlif.Positions, p.CurrentToken.Position)
			p.Advance()
			continue
		} else {
			p.Expect(ast.TypeNewline, "")
			mlif.ElseBlock = p.ParseBlock(func() bool {
				return p.IsCurrent(ast.TypeKeyword, "end")
			})
			break
		}
	}

	p.Expect(ast.TypeKeyword, "end")

	if !p.IsCurrentType(ast.TypeEOF) {
		p.Expect(ast.TypeNewline, "")
	}

	return &mlif
}

// ParseWhile pasres a nolol while
func (p *Parser) ParseWhile() nast.NestableElement {
	p.Log()
	loop := nast.WhileLoop{
		Position: p.CurrentToken.Position,
	}
	if !p.IsCurrent(ast.TypeKeyword, "while") {
		return nil
	}
	p.Advance()

	loop.Condition = p.This.ParseExpression()
	if loop.Condition == nil {
		p.ErrorExpectedExpression("as loop-condition")
	}

	p.Expect(ast.TypeKeyword, "do")
	p.Expect(ast.TypeNewline, "")

	loop.Block = p.ParseBlock(func() bool {
		return p.IsCurrent(ast.TypeKeyword, "end")
	})

	p.Expect(ast.TypeKeyword, "end")

	if !p.IsCurrentType(ast.TypeEOF) {
		p.Expect(ast.TypeNewline, "")
	}

	return &loop
}

// ParseBlock parse lines until stop() returns true
func (p *Parser) ParseBlock(stop func() bool) *nast.Block {
	p.Log()
	elements := make([]nast.NestableElement, 0)
	for p.HasNext() && !stop() {
		element := p.ParseNestableElement()
		if elements == nil {
			break
		}
		elements = append(elements, element)
	}
	return &nast.Block{
		Elements: elements,
	}
}

// ParseFuncCall parse a function call
func (p *Parser) ParseFuncCall() *nast.FuncCall {
	p.Log()
	nextToken := p.Tokenizer.Peek()
	if !p.IsCurrentType(ast.TypeID) || nextToken.Type != ast.TypeSymbol || nextToken.Value != "(" {
		return nil
	}
	fc := &nast.FuncCall{
		Position:  p.CurrentToken.Position,
		Function:  p.CurrentToken.Value,
		Arguments: make([]ast.Expression, 0),
	}
	p.Advance()
	p.Advance()

	for !p.IsCurrent(ast.TypeSymbol, ")") {
		exp := p.ParseExpression()
		if exp == nil {
			p.ErrorExpectedExpression("as arguments(s)")
			break
		}
		fc.Arguments = append(fc.Arguments, exp)
		if p.IsCurrent(ast.TypeSymbol, ",") {
			p.Advance()
			continue
		}
		break
	}

	p.Expect(ast.TypeSymbol, ")")

	return fc
}

// ParseNestableElementFuncCall parses a funccall that is the only element on the line
func (p *Parser) ParseNestableElementFuncCall() *nast.FuncCall {
	p.Log()
	savedToken := p.CurrentToken
	tokenizerCheckpoint := p.Tokenizer.Checkpoint()
	fc := p.ParseFuncCall()
	if fc != nil {
		if !p.IsCurrentType(ast.TypeNewline) && !p.IsCurrentType(ast.TypeEOF) {
			// This is not a single funccall. Its probably a StatementLine. Reset parser to checkpoint
			p.CurrentToken = savedToken
			p.Tokenizer.Restore(tokenizerCheckpoint)
			return nil
		}
		fc.Type = nast.MacroTypeBlock
	}
	return fc
}

// ParseSingleExpression wraps the method of the yolol-parser and adds parsing of func-calls
func (p *Parser) ParseSingleExpression() ast.Expression {
	p.Log()
	funccall := p.ParseFuncCall()
	if funccall != nil {
		funccall.Type = nast.MacroTypeExpr
		return funccall
	}
	return p.Parser.ParseSingleExpression()
}

// ParseStatement wraps the method of the yolol-parser to add new statement-types
func (p *Parser) ParseStatement() ast.Statement {
	p.Log()
	breakstmt := p.ParseBreak()
	if breakstmt != nil {
		return breakstmt
	}
	continuestmt := p.ParseContinue()
	if continuestmt != nil {
		return continuestmt
	}
	funccall := p.ParseFuncCall()
	if funccall != nil {
		funccall.Type = nast.MacroTypeLine
		return funccall
	}
	return p.Parser.ParseStatement()
}

// ParseBreak parses the break keyword
func (p *Parser) ParseBreak() ast.Statement {
	p.Log()
	if p.IsCurrent(ast.TypeKeyword, "break") {
		rval := &nast.BreakStatement{
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		return rval
	}
	return nil
}

// ParseContinue parses the continue keyword
func (p *Parser) ParseContinue() ast.Statement {
	p.Log()
	if p.IsCurrent(ast.TypeKeyword, "continue") {
		rval := &nast.ContinueStatement{
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		return rval
	}
	return nil
}

// ParseIf parses an if-node. Copied nearly 1to1 from yolol-parser, but requires ; instead of " " between statements
func (p *Parser) ParseIf() ast.Statement {
	p.Log()
	ret := ast.IfStatement{
		Position: p.CurrentToken.Position,
	}
	if !p.IsCurrent(ast.TypeKeyword, "if") {
		return nil
	}
	p.Advance()

	ret.Condition = p.This.ParseExpression()
	if ret.Condition == nil {
		p.ErrorExpectedExpression("as if-condition")
	}

	p.Expect(ast.TypeKeyword, "then")
	ret.IfBlock = make([]ast.Statement, 0, 1)

	for p.HasNext() {
		stmt := p.This.ParseStatement()
		if stmt == nil {
			break
		}
		ret.IfBlock = append(ret.IfBlock, stmt)
		if !p.IsCurrent(ast.TypeSymbol, ";") {
			break
		}
		p.Advance()
	}

	if p.IsCurrent(ast.TypeKeyword, "else") {
		p.Advance()

		ret.ElseBlock = make([]ast.Statement, 0, 1)

		for p.HasNext() {
			stmt := p.This.ParseStatement()
			if stmt == nil {
				break
			}
			ret.ElseBlock = append(ret.ElseBlock, stmt)
			if !p.IsCurrent(ast.TypeSymbol, ";") {
				break
			}
			p.Advance()
		}
	}

	p.Expect(ast.TypeKeyword, "end")

	return &ret
}
