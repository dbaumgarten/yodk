package nolol

import (
	"path/filepath"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/optimizers"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Converter can convert a nolol-ast to a yolol-ast
type Converter struct {
	prog          *nast.Program
	convertedProg *ast.Program
	files         FileSystem
	err           error
	// all fount line-labels
	lineLabels map[string]int
	// the names of definitions are case-insensitive. Keys are converted to lowercase before using them
	// all lookups MUST also use lowercased keys
	definitions      map[string]*nast.Definition
	usesTimeTracking bool
	iflabelcounter   int
	waitlabelcounter int
	loopcounter      int
	// keeps track of the current loop we are in while converting
	// the last element in the list is the current innermost loop
	loopLevel        []loopinfo
	sexpOptimizer    *optimizers.StaticExpressionOptimizer
	boolexpOptimizer *optimizers.ExpressionInversionOptimizer
	varnameOptimizer *optimizers.VariableNameOptimizer
	includecount     int
	// holds all found defined macros
	macros map[string]*nast.MacroDefinition
	// a stack of macro-scopes, used for renaming local vars
	macroLevel []string
	// a count of total insertions to detect loops
	macroInsertionCount       int
	macroCurrentStatementLine *nast.StatementLine
	macroCurrentStatement     int
	// if true, enable debug-logging
	debug bool
}

// NewConverter creates a new converter
func NewConverter() ConverterEmpty {
	return &Converter{
		lineLabels:       make(map[string]int),
		definitions:      make(map[string]*nast.Definition),
		macros:           make(map[string]*nast.MacroDefinition),
		macroLevel:       make([]string, 0),
		sexpOptimizer:    optimizers.NewStaticExpressionOptimizer(),
		boolexpOptimizer: &optimizers.ExpressionInversionOptimizer{},
		varnameOptimizer: optimizers.NewVariableNameOptimizer(),
		loopLevel:        make([]loopinfo, 0),
	}
}

func (c *Converter) Error() error {
	return c.err
}

// Convert converts the nolol-program to a yolol-program
// This is a shortcut to calling the ProcessXY-Methods in order
func (c *Converter) Convert() (*ast.Program, error) {
	return c.ProcessIncludes().
		ProcessCodeExpansion().
		ProcessNodes().
		ProcessLineNumbers().
		ProcessFinalize().
		Get()
}

// RunConversion jumps to the final phase of conversion
func (c *Converter) RunConversion() ConverterDone {
	return c.ProcessIncludes().
		ProcessCodeExpansion().
		ProcessNodes().
		ProcessLineNumbers().
		ProcessFinalize()
}

// SetDebug enables/disables debug logging
func (c *Converter) SetDebug(b bool) ConverterEmpty {
	c.debug = b
	return c
}

// LoadFile is a shortcut that loads a file to convert from the file-system
// mainfile is the path to the file on the disk.
// All included files are loaded relative to the mainfile.
func (c *Converter) LoadFile(mainfile string) ConverterIncludes {
	files := DiskFileSystem{
		Dir: filepath.Dir(mainfile),
	}
	return c.LoadFileEx(filepath.Base(mainfile), files)
}

// LoadFileEx acts like LoadFile, but allows the passing of a custom filesystem from which the source files
// are retrieved. This way, files that are not stored on disk can be converted
func (c *Converter) LoadFileEx(mainfile string, files FileSystem) ConverterIncludes {
	file, err := files.Get(mainfile)
	if err != nil {
		c.err = err
		return c
	}
	p := NewParser()
	p.Debug(c.debug)
	parsed, err := p.Parse(file)
	if err != nil {
		c.err = err
		return c
	}
	return c.Load(parsed, files)
}

// Load loades a nolol-ast to convert. Included files are retrieved using the given filesystem
func (c *Converter) Load(prog *nast.Program, files FileSystem) ConverterIncludes {
	c.prog = prog
	c.files = files
	return c
}

// GetIntermediateProgram returns the current intermediate nolol-ast.
// The state of it depends on what steps of the conversion have already been performed
func (c *Converter) GetIntermediateProgram() *nast.Program {
	return c.prog
}

// ProcessIncludes resolves all Include-Directives in the given nolol-code
func (c *Converter) ProcessIncludes() ConverterExpansions {
	if c.err != nil {
		return c
	}

	f := func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *nast.IncludeDirective:
			return c.convertInclude(n)
		}
		return nil
	}
	c.err = c.prog.Accept(ast.VisitorFunc(f))
	return c
}

// ProcessCodeExpansion resolves all macro-definitions, macro-insertions and defines
func (c *Converter) ProcessCodeExpansion() ConverterNodes {
	if c.err != nil {
		return c
	}

	f := func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *nast.Definition:
			return c.convertDefinition(n, visitType)

		case *ast.Assignment:
			return c.convertDefinitionAssignment(n, visitType)

		case *ast.Dereference:
			return c.convertDefinitionDereference(n)

		case *nast.MacroDefinition:
			return c.convertMacroDef(n, visitType)

		case *nast.FuncCall:
			return c.convertMacroInsertion(n, visitType)

		case *InsertedMacro:
			return c.convertInsertedMacro(n, visitType)

		case *nast.StatementLine:
			if visitType == ast.PreVisit {
				c.macroCurrentStatementLine = n
			}
			if visitType >= 0 {
				c.macroCurrentStatement = visitType
			}
		}
		return nil
	}
	c.err = c.prog.Accept(ast.VisitorFunc(f))
	return c
}

// ProcessNodes converts most nolol ast-nodes to yolol ast.nodes
func (c *Converter) ProcessNodes() ConverterLines {
	if c.err != nil {
		return c
	}

	c.usesTimeTracking = usesTimeTracking(c.prog)
	// reserve a name for use in time-tracking
	c.varnameOptimizer.OptimizeVarName(reservedTimeVariable)

	// find all user-defined line-labels
	err := c.findLineLabels(c.prog, false)
	if err != nil {
		c.err = err
		return c
	}

	// convert the remaining nodes to yolol
	f := func(node ast.Node, visitType int) error {
		switch n := node.(type) {

		case *nast.FuncCall:
			return c.convertFuncCall(n, visitType)

		case *nast.MultilineIf:
			return c.convertIf(n, visitType)

		case *nast.WhileLoop:
			return c.convertWhileLoop(n, visitType)

		case *nast.BreakStatement:
			return c.convertBreakStatement(n)

		case *nast.ContinueStatement:
			return c.convertContinueStatement(n)

		case *ast.Assignment:
			return c.convertAssignment(n, visitType)

		case *ast.Dereference:
			return c.convertDereference(n)

		case *ast.UnaryOperation:
			if visitType == ast.PostVisit {
				return ast.NewNodeReplacementSkip(c.optimizeExpression(n))
			}
		case *ast.BinaryOperation:
			if visitType == ast.PostVisit {
				return ast.NewNodeReplacementSkip(c.optimizeExpression(n))
			}
		}

		return nil
	}
	c.err = c.prog.Accept(ast.VisitorFunc(f))
	return c
}

// ProcessLineNumbers handles gotos and line-labels
func (c *Converter) ProcessLineNumbers() ConverterFinal {
	if c.err != nil {
		return c
	}

	c.err = c.addFinalGoto(c.prog)
	if c.err != nil {
		return c
	}

	c.err = c.resolveGotoChains(c.prog)
	if c.err != nil {
		return c
	}

	c.err = c.removeUnusedLabels(c.prog)
	if c.err != nil {
		return c
	}

	// merge the statemens of the program as good as possible
	merged, err := c.mergeNololElements(c.prog.Elements)
	if err != nil {
		c.err = err
		return c
	}
	c.prog.Elements = merged

	c.err = c.removeDuplicateGotos(c.prog)
	if c.err != nil {
		return c
	}

	// find all line-labels (again). This time they have the correct lines.
	c.err = c.findLineLabels(c.prog, true)
	if c.err != nil {
		return c
	}

	// resolve line-labels
	c.err = c.replaceLineLabels(c.prog)
	if c.err != nil {
		return c
	}

	// replacing line-labels with actual line-numbers might have introduced un-optimized expression
	// re-run the static-expression optimizer
	c.err = c.sexpOptimizer.Optimize(c.prog)
	return c
}

// ProcessFinalize takes the final steps in converting the program
func (c *Converter) ProcessFinalize() ConverterDone {
	if c.err != nil {
		return c
	}

	if c.usesTimeTracking {
		c.insertLineCounter(c.prog)
	}

	// at this point the program consists entirely of statement-lines which contain pure yolol-code
	c.convertedProg = &ast.Program{
		Lines: make([]*ast.Line, len(c.prog.Elements)),
	}

	for i, element := range c.prog.Elements {
		line := element.(*nast.StatementLine)
		c.convertedProg.Lines[i] = &ast.Line{
			Position:   line.Position,
			Statements: line.Statements,
		}
	}

	c.removeFinalGotoIfNeeded(c.convertedProg)

	if len(c.convertedProg.Lines) > 20 {
		c.err = &parser.Error{
			Message: "Program is too large to be compiled into 20 lines of yolol.",
			StartPosition: ast.Position{
				Line:    1,
				Coloumn: 1,
			},
			EndPosition: ast.Position{
				Line:    30,
				Coloumn: 70,
			},
		}
	}

	return c
}

// Get returns the converted program and/or an error
func (c *Converter) Get() (*ast.Program, error) {
	return c.convertedProg, c.err
}

// GetVariableTranslations returns a table that can be used to find the original names
// of the variables whos names where shortened during conversion
func (c *Converter) GetVariableTranslations() map[string]string {
	return c.varnameOptimizer.GetReversalTable()
}
