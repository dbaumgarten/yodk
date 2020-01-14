package parser

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Error represents an error encountered during parsing
type Error struct {
	Message       string
	StartPosition ast.Position
	EndPosition   ast.Position
}

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
