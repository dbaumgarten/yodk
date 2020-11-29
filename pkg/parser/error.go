package parser

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Error represents an error encountered during parsing
type Error struct {
	// The human-readable error-message
	Message string
	// Where the error started
	StartPosition ast.Position
	// Where the error ends
	EndPosition ast.Position
	// Machine-Readable error-code
	Code string
	// A token that was expected here (optional)
	ExpectedToken *ast.Token
}

// Predefined constants for Error.Code
const (
	ErrExpectedExpression = "ErrExpectedExpression"
	ErrExpectedStatement  = "ErrExpectedStatement"
	ErrExpectedToken      = "ErrExpectedToken"
	ErrExpectedAssignop   = "ErrExpectedAssignop"
)

func (e Error) Error() string {
	if e.StartPosition != e.EndPosition {
		return fmt.Sprintf("Parser error at %s (up to %s): %s", e.StartPosition.String(), e.EndPosition.String(), e.Message)
	}
	return fmt.Sprintf("Parser error at %s: %s", e.StartPosition.String(), e.Message)
}

// Errors represents multiple Errors
type Errors []*Error

func (e Errors) Error() string {
	str := ""
	for _, err := range e {
		str += err.Error() + "\n\n"
	}
	return str
}
