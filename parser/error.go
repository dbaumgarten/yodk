package parser

import "fmt"

// Error represents an error encountered during parsing
type Error struct {
	Message       string
	StartPosition Position
	EndPosition   Position
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
