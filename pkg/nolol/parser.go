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

	block := p.ParseWaitStatement()
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

	constDecl := p.ParseConstantDeclaration()
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

	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "include" {
		return nil
	}
	incl := &nast.IncludeDirective{
		Position: p.CurrentToken.Position,
	}
	p.Advance()
	if p.CurrentToken.Type != ast.TypeString {
		p.ErrorCurrent("Expected a string-constant after include")
		return incl
	}
	incl.File = p.CurrentToken.Value
	p.Advance()
	if p.CurrentToken.Type != ast.TypeEOF {
		p.Expect(ast.TypeNewline, "")
	}
	return incl
}

// ParseMacroDefinition parses the definition of a macro
func (p *Parser) ParseMacroDefinition() *nast.MacroDefinition {
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "macro" {
		return nil
	}
	p.Advance()
	mdef := &nast.MacroDefinition{
		Position:  p.CurrentToken.Position,
		Arguments: []string{},
	}
	if p.CurrentToken.Type != ast.TypeID {
		p.ErrorCurrent("Expected an idantifier after the macro keyword")
		return mdef
	}
	mdef.Name = p.CurrentToken.Value
	p.Advance()

	p.Expect(ast.TypeSymbol, "(")

	for p.CurrentToken.Type != ast.TypeSymbol || p.CurrentToken.Value != ")" {
		if p.CurrentToken.Type != ast.TypeID {
			p.ErrorCurrent("Only comma separated identifiers are allowed as arguments in a macro definition")
			break
		}
		mdef.Arguments = append(mdef.Arguments, p.CurrentToken.Value)
		p.Advance()
		if p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == "," {
			p.Advance()
			continue
		}
		break
	}

	p.Expect(ast.TypeSymbol, ")")

	p.Expect(ast.TypeNewline, "")

	mdef.Block = p.ParseBlock(func() bool {
		return p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "end"
	})
	p.Expect(ast.TypeKeyword, "end")

	return mdef
}

// ParseMacroInsertion parses a macro insertion
func (p *Parser) ParseMacroInsertion() *nast.MacroInsetion {
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "insert" {
		return nil
	}
	p.Advance()
	mins := &nast.MacroInsetion{
		Position:  p.CurrentToken.Position,
		Arguments: []ast.Expression{},
	}
	if p.CurrentToken.Type != ast.TypeID {
		p.ErrorCurrent("Expected an idantifier after the macro keyword")
		return mins
	}
	mins.Name = p.CurrentToken.Value
	p.Advance()

	p.Expect(ast.TypeSymbol, "(")

	for p.CurrentToken.Type != ast.TypeSymbol || p.CurrentToken.Value != ")" {
		exp := p.ParseExpression()
		if exp == nil {
			p.ErrorCurrent("Expected expression(s) as argument to the macro")
			break
		}
		mins.Arguments = append(mins.Arguments, exp)
		if p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == "," {
			p.Advance()
			continue
		}
		break
	}

	p.Expect(ast.TypeSymbol, ")")

	if p.CurrentToken.Type != ast.TypeEOF {
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
	if p.CurrentToken.Type == ast.TypeID && (p.NextToken.Type == ast.TypeSymbol && p.NextToken.Value == ">") {
		ret.Label = strings.ToLower(p.CurrentToken.Value)
		p.Advance()
		p.Advance()
	}

	// this line has no statements, only a comment
	if p.CurrentToken.Type == ast.TypeComment {
		ret.Comment = p.CurrentToken.Value
		p.Advance()
	}

	if p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == "$" {
		ret.HasBOL = true
		p.Advance()
	}

	// the line has no statements
	if p.CurrentToken.Type == ast.TypeEOF || p.CurrentToken.Type == ast.TypeNewline || p.CurrentToken.Type == ast.TypeComment {
		if p.CurrentToken.Type == ast.TypeComment {
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

	for p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == ";" {
		p.Advance()
		stmt = p.This.ParseStatement()
		if stmt != nil {
			ret.Statements = append(ret.Statements, stmt)
		} else {
			p.ErrorCurrent(("Expected a statement after ';'"))
		}
	}

	if p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == "$" {
		ret.HasEOL = true
		p.Advance()
	}

	// This line has statements and a comment at the end
	if p.CurrentToken.Type == ast.TypeComment {
		ret.Comment = p.CurrentToken.Value
		p.Advance()
	}

	if p.CurrentToken.Type != ast.TypeEOF {
		p.Expect(ast.TypeNewline, "")
	}

	return &ret
}

// ParseWaitStatement parses a NOLOL wait-statement
func (p *Parser) ParseWaitStatement() *nast.WaitDirective {
	p.Log()
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "wait" {
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

	return st
}

// ParseConstantDeclaration parses a constant declaration
func (p *Parser) ParseConstantDeclaration() *nast.ConstDeclaration {
	p.Log()
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "const" {
		return nil
	}
	startpos := p.CurrentToken.Position
	p.Advance()
	if p.CurrentToken.Type != ast.TypeID {
		p.ErrorCurrent("const keyword must be followed by an identifier")
	}
	decl := &nast.ConstDeclaration{
		Name:        strings.ToLower(p.CurrentToken.Value),
		DisplayName: p.CurrentToken.Value,
		Position:    startpos,
	}
	p.Advance()
	p.Expect(ast.TypeSymbol, "=")
	value := p.ParseSingleExpression()
	if value == nil {
		p.ErrorCurrent("The = of a const declaration must be followed by an expression")
	}
	decl.Value = value
	if p.CurrentToken.Type != ast.TypeEOF {
		p.Expect(ast.TypeNewline, "")
	}
	return decl
}

// ParseMultilineIf parses a nolol-style multiline if
func (p *Parser) ParseMultilineIf() nast.Element {
	p.Log()
	mlif := nast.MultilineIf{
		Position:   p.CurrentToken.Position,
		Conditions: make([]ast.Expression, 0),
		Blocks:     make([]*nast.Block, 0),
	}
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "if" {
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
			return p.CurrentToken.Type == ast.TypeKeyword && (p.CurrentToken.Value == "end" || p.CurrentToken.Value == "else")
		})
		mlif.Conditions = append(mlif.Conditions, condition)
		mlif.Blocks = append(mlif.Blocks, block)

		if p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "end" {
			break
		}

		if p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "else" {
			p.Advance()
		}

		if p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "if" {
			p.Advance()
			continue
		} else {
			p.Expect(ast.TypeNewline, "")
			mlif.ElseBlock = p.ParseBlock(func() bool {
				return p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "end"
			})
			break
		}
	}

	p.Expect(ast.TypeKeyword, "end")

	if p.CurrentToken.Type != ast.TypeEOF {
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
	if p.CurrentToken.Type != ast.TypeKeyword || p.CurrentToken.Value != "while" {
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
		return p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "end"
	})

	p.Expect(ast.TypeKeyword, "end")

	if p.CurrentToken.Type != ast.TypeEOF {
		p.Expect(ast.TypeNewline, "")
	}

	return &loop
}

// ParseIf overrides and disables the old yolol-style inline ifs
func (p *Parser) ParseIf() ast.Statement {
	p.Log()
	//Inline if is not supported by nolol. Always return nil
	return nil
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
	if p.CurrentToken.Type == ast.TypeKeyword && p.CurrentToken.Value == "goto" {
		p.Advance()

		stmt := &nast.GoToLabelStatement{
			Position: p.CurrentToken.Position,
			Label:    strings.ToLower(p.CurrentToken.Value),
		}

		if p.CurrentToken.Type != ast.TypeID {
			p.ErrorCurrent("Goto must be followed by an identifier")
		} else {
			p.Advance()
		}

		return stmt
	}
	return nil
}

// ParseFuncCall parse a function call
func (p *Parser) ParseFuncCall() ast.Expression {
	p.Log()
	if p.CurrentToken.Type != ast.TypeID || p.NextToken.Type != ast.TypeSymbol || p.NextToken.Value != "(" {
		return nil
	}
	fc := &ast.FuncCall{
		Position: p.CurrentToken.Position,
		Function: strings.ToLower(p.CurrentToken.Value),
	}
	p.Advance()
	p.Advance()

	if p.CurrentToken.Type == ast.TypeSymbol && p.CurrentToken.Value == ")" {
		p.Advance()
		return fc
	}

	arg := p.This.ParseExpression()
	fc.Argument = arg
	if arg == nil {
		p.ErrorCurrent("Expected a function argument or ')'")
	}

	p.Expect(ast.TypeSymbol, ")")

	return fc
}
