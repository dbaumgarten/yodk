package parser

import "fmt"

// Error represents an error encountered during parsing
type Error struct {
	Message       string
	StartPosition Position
	EndPosition   Position
	ErrorStack    []error
	Fatal         bool
}

// Append adds a follow-up error to the error to make the displayed error-message more helpful
func (e *Error) Append(err error) *Error {
	if e.ErrorStack == nil {
		e.ErrorStack = make([]error, 0)
	}
	e.ErrorStack = append(e.ErrorStack, err)
	return e
}

func (e Error) Error() string {
	txt := fmt.Sprintf("Parser error at %s (up to %s): %s", e.StartPosition.String(), e.EndPosition.String(), e.Message)
	if e.ErrorStack != nil {
		txt += "\n" + "Following errors:\n"
		for _, err := range e.ErrorStack {
			txt += "    " + err.Error() + "\n"
		}
	}
	return txt
}

// Errors represents multiple Errors
type Errors []*Error

func (e Errors) Error() string {
	str := ""
	for _, err := range e {
		str += err.Error() + "\n"
	}
	return str
}
