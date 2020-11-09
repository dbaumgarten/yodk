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

// NewParser creates and returns a nolol parser
func NewParser() *Parser {
	ep := &Parser{
		Parser: parser.NewParser(),
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
	// Advance twice to fill CurrentToken and NextToken
	p.Advance()
	p.Advance()
	parsed := p.ParseProgram()
	if len(p.Errors) == 0 {
		return parsed, nil
	}
	return nil, p.Errors
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

	block := p.ParseWaitDirective()
	if block != nil {
		return block
	}

	mIns := p.ParseMacroInsertion()
	if mIns != nil {
		return mIns
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
		p.ErrorCurrent("Expected a string-constant after include")
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
		p.ErrorCurrent("Expected an idantifier after the macro keyword")
		return mdef
	}
	mdef.Name = p.CurrentToken.Value
	p.Advance()

	p.Expect(ast.TypeSymbol, "(")
	for !p.IsCurrent(ast.TypeSymbol, ")") {
		if !p.IsCurrentType(ast.TypeID) {
			p.ErrorCurrent("Only comma separated identifiers are allowed as arguments in a macro definition")
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

	if p.IsCurrent(ast.TypeSymbol, "(") {
		p.Advance()
		for !p.IsCurrent(ast.TypeSymbol, ")") {
			if !p.IsCurrentType(ast.TypeID) {
				p.ErrorCurrent("Only comma separated identifiers are allowed as globals in a macro definition")
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
		p.Expect(ast.TypeSymbol, ")")
	}

	p.Expect(ast.TypeNewline, "")

	mdef.Block = p.ParseBlock(func() bool {
		return p.IsCurrent(ast.TypeKeyword, "end")
	})
	p.Expect(ast.TypeKeyword, "end")

	return mdef
}

// ParseMacroInsertion parses a macro insertion
func (p *Parser) ParseMacroInsertion() *nast.MacroInsetion {
	p.Log()
	if !p.IsCurrent(ast.TypeKeyword, "insert") {
		return nil
	}
	p.Advance()
	mins := &nast.MacroInsetion{
		Position: p.CurrentToken.Position,
	}

	mins.FuncCall = p.ParseFuncCall()

	if mins.FuncCall == nil {
		p.ErrorCurrent("Expected a macro instanziation after the insert keyword")
		return mins
	}

	if !p.IsCurrentType(ast.TypeEOF) {
		p.Expect(ast.TypeNewline, "")
	}

	return mins
}

// ParseStatementLine parses a statement line
func (p *Parser) ParseStatementLine() *nast.StatementLine {
	p.Log()
	ret := nast.StatementLine{
		Line: ast.Line{
			Statements: make([]ast.Statement, 0, 1),
		},
		Position: p.CurrentToken.Position,
	}

	// get line-label if it exists
	if p.IsCurrentType(ast.TypeID) && (p.NextToken.Type == ast.TypeSymbol && p.NextToken.Value == ">") {
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
		p.ErrorCurrent("Expected a statement")
		p.Advance()
		return &ret
	}

	for p.IsCurrent(ast.TypeSymbol, ";") {
		p.Advance()
		stmt = p.This.ParseStatement()
		if stmt != nil {
			ret.Statements = append(ret.Statements, stmt)
		} else {
			p.ErrorCurrent(("Expected a statement after ';'"))
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

// ParseWaitDirective parses a NOLOL wait-statement
func (p *Parser) ParseWaitDirective() *nast.WaitDirective {
	p.Log()
	if !p.IsCurrent(ast.TypeKeyword, "wait") {
		return nil
	}
	p.Advance()
	st := &nast.WaitDirective{
		Position: p.CurrentToken.Position,
	}

	st.Condition = p.This.ParseExpression()
	if st.Condition == nil {
		p.ErrorCurrent("Expected an expression after 'block'")
	}

	if p.IsCurrent(ast.TypeKeyword, "then") {
		p.Advance()

		st.Statements = make([]ast.Statement, 0)
		stmt := p.This.ParseStatement()
		// at this point, the line must at least have one statement
		if stmt != nil {
			st.Statements = append(st.Statements, stmt)
		} else {
			p.ErrorCurrent("Expected a statement")
			p.Advance()
			return st
		}

		for p.IsCurrent(ast.TypeSymbol, ";") {
			p.Advance()
			stmt = p.This.ParseStatement()
			if stmt != nil {
				st.Statements = append(st.Statements, stmt)
			} else {
				p.ErrorCurrent(("Expected a statement after ';'"))
			}
		}

		p.Expect(ast.TypeKeyword, "end")
	}

	return st
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
		p.ErrorCurrent("const keyword must be followed by an identifier")
	}
	decl := &nast.Definition{
		Name:         p.CurrentToken.Value,
		Position:     startpos,
		Placeholders: make([]string, 0),
	}
	p.Advance()

	if p.IsCurrent(ast.TypeSymbol, "(") {
		starttoken := p.Advance()
		for !p.IsCurrent(ast.TypeSymbol, ")") {
			if !p.IsCurrentType(ast.TypeID) {
				p.ErrorCurrent("Only comma separated identifiers are allowed as arguments in a definition")
				break
			}
			decl.Placeholders = append(decl.Placeholders, p.CurrentToken.Value)
			p.Advance()
			if p.IsCurrent(ast.TypeSymbol, ",") {
				p.Advance()
				continue
			}
			break
		}
		endpos := p.Expect(ast.TypeSymbol, ")")
		if len(decl.Placeholders) == 0 {
			p.Error("Definitions with placeholder-parenthesis need at least one placeholder", starttoken.Position, endpos)
		}
	}

	p.Expect(ast.TypeSymbol, "=")
	value := p.ParseExpression()
	if value == nil {
		p.ErrorCurrent("The = of a const declaration must be followed by an expression")
	}
	decl.Value = value
	if !p.IsCurrentType(ast.TypeEOF) {
		p.Expect(ast.TypeNewline, "")
	}
	return decl
}

// ParseMultilineIf parses a nolol-style multiline if
func (p *Parser) ParseMultilineIf() nast.Element {
	p.Log()
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
			p.ErrorCurrent("No expression found as if-condition")
			p.Advance()
		}

		p.Expect(ast.TypeKeyword, "then")
		p.Expect(ast.TypeNewline, "")

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
func (p *Parser) ParseWhile() nast.Element {
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
		p.ErrorCurrent("No expression found as loop-condition")
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

// ParseIf is copied nearly exactly from the yolol-parser, but the start token changed from "if" to "_if"
func (p *Parser) ParseIf() ast.Statement {
	p.Log()
	ret := ast.IfStatement{
		Position: p.CurrentToken.Position,
	}
	if !p.IsCurrent(ast.TypeKeyword, "_if") {
		return nil
	}
	p.Advance()

	ret.Condition = p.This.ParseExpression()
	if ret.Condition == nil {
		p.ErrorCurrent("No expression found as inline-if-condition")
	}

	p.Expect(ast.TypeKeyword, "then")

	stmt := p.This.ParseStatement()
	if stmt == nil {
		p.ErrorCurrent("If-block needs at least one statement")
	}
	ret.IfBlock = make([]ast.Statement, 0, 1)
	ret.IfBlock = append(ret.IfBlock, stmt)

	for {
		stmt2 := p.This.ParseStatement()
		if stmt2 == nil {
			break
		}
		ret.IfBlock = append(ret.IfBlock, stmt2)
	}

	if p.IsCurrent(ast.TypeKeyword, "else") {
		p.Advance()
		stmt := p.This.ParseStatement()
		if stmt == nil {
			p.ErrorCurrent("Else-block needs at least one statement")
		}
		ret.ElseBlock = make([]ast.Statement, 0, 1)
		ret.ElseBlock = append(ret.ElseBlock, stmt)

		for {
			stmt2 := p.This.ParseStatement()
			if stmt2 == nil {
				break
			}
			ret.ElseBlock = append(ret.IfBlock, stmt2)
		}
	}

	p.Expect(ast.TypeKeyword, "end")

	return &ret
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

// ParseGoto allows labeled-gotos and forbids line-based gotos
func (p *Parser) ParseGoto() ast.Statement {
	p.Log()
	// this is a nolol label-goto
	if p.IsCurrent(ast.TypeKeyword, "goto") {
		p.Advance()

		stmt := &nast.GoToLabelStatement{
			Position: p.CurrentToken.Position,
			Label:    strings.ToLower(p.CurrentToken.Value),
		}

		if !p.IsCurrentType(ast.TypeID) {
			p.ErrorCurrent("Goto must be followed by an identifier")
		} else {
			p.Advance()
		}

		return stmt
	}

	// this is a yolol-style expression-goto
	if p.IsCurrent(ast.TypeKeyword, "_goto") {
		stmt := ast.GoToStatement{
			Position: p.CurrentToken.Position,
		}
		p.Advance()
		stmt.Line = p.This.ParseExpression()
		if stmt.Line == nil {
			p.Error("Goto must be followed by an expression", stmt.Start(), stmt.Start())
		}
		if _, is := stmt.Line.(*ast.StringConstant); is {
			p.Error("Can not go to a string", stmt.Start(), stmt.Start())
		}
		return &stmt
	}
	return nil
}

// ParseFuncCall parse a function call
func (p *Parser) ParseFuncCall() *nast.FuncCall {
	p.Log()
	if !p.IsCurrentType(ast.TypeID) || p.NextToken.Type != ast.TypeSymbol || p.NextToken.Value != "(" {
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
			p.ErrorCurrent("Expected expression(s) as arguments(s)")
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

// ParseSingleExpression wraps the method of the yolol-parser and adds parsing of func-calls
func (p *Parser) ParseSingleExpression() ast.Expression {
	p.Log()
	funccall := p.ParseFuncCall()
	if funccall != nil {
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
