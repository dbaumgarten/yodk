package nolol

import (
	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// ConverterEmpty is part of the Sequenced-Builder-Pattern of the Converter
type ConverterEmpty interface {
	Load(prog *nast.Program, files FileSystem) ConverterIncludes
	LoadFile(mainfile string) ConverterIncludes
	LoadFileEx(mainfile string, files FileSystem) ConverterIncludes
	SetDebug(b bool) ConverterEmpty
}

// ConverterIncludes is part of the Sequenced-Builder-Pattern of the Converter
type ConverterIncludes interface {
	ProcessIncludes() ConverterExpansions
	Convert() (*ast.Program, error)
	RunConversion() ConverterDone
	Error() error
	GetIntermediateProgram() *nast.Program
}

// ConverterExpansions is part of the Sequenced-Builder-Pattern of the Converter
type ConverterExpansions interface {
	ProcessCodeExpansion() ConverterNodes
	Error() error
	GetIntermediateProgram() *nast.Program
}

// ConverterNodes is part of the Sequenced-Builder-Pattern of the Converter
type ConverterNodes interface {
	ProcessNodes() ConverterLines
	Error() error
	GetIntermediateProgram() *nast.Program
}

// ConverterLines is part of the Sequenced-Builder-Pattern of the Converter
type ConverterLines interface {
	ProcessLineNumbers() ConverterFinal
	Error() error
	GetIntermediateProgram() *nast.Program
}

// ConverterFinal is part of the Sequenced-Builder-Pattern of the Converter
type ConverterFinal interface {
	ProcessFinalize() ConverterDone
	Error() error
	GetIntermediateProgram() *nast.Program
}

// ConverterDone is part of the Sequenced-Builder-Pattern of the Converter
type ConverterDone interface {
	Get() (*ast.Program, error)
	GetVariableTranslations() map[string]string
	Error() error
	GetIntermediateProgram() *nast.Program
}
