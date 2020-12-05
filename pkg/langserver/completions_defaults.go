package langserver

import "github.com/dbaumgarten/yodk/pkg/lsp"

// DefaultYololCompletions contains completion-items for yolol-builtins
var DefaultYololCompletions = []lsp.CompletionItem{
	{
		Label: "if",
		Kind:  14,
	},
	{
		Label: "then",
		Kind:  14,
	},
	{
		Label: "else",
		Kind:  14,
	},
	{
		Label: "end",
		Kind:  14,
	},
	{
		Label: "goto",
		Kind:  14,
	},
	{
		Label:         "not",
		Detail:        "not X",
		Kind:          24,
		Documentation: "Returns 1 if X is 0, otherwise it returns 0",
	},
	{
		Label:         "and",
		Detail:        "X and Y",
		Kind:          24,
		Documentation: "Returns true if X and Y are true",
	},
	{
		Label:         "or",
		Detail:        "X or Y",
		Kind:          24,
		Documentation: "Returns true if X or Y are true",
	},

	{
		Label:         "abs",
		Detail:        "abs X",
		Kind:          3,
		Documentation: "Returns the absolute value of X",
	},
	{
		Label:         "sqrt",
		Detail:        "sqrt X",
		Kind:          3,
		Documentation: "Returns the square-root of X",
	},
	{
		Label:         "sin",
		Detail:        "sin X",
		Kind:          3,
		Documentation: "Returns the sine (degree) of X",
	},
	{
		Label:         "cos",
		Detail:        "cos X",
		Kind:          3,
		Documentation: "Returns the cosine (degree) of X",
	},
	{
		Label:         "tan",
		Detail:        "tan X",
		Kind:          3,
		Documentation: "Returns the tangent (degree) of X",
	},
	{
		Label:         "asin",
		Detail:        "asin X",
		Kind:          3,
		Documentation: "Returns the inverse sine (degree) of X",
	},
	{
		Label:         "acos",
		Detail:        "acos X",
		Kind:          3,
		Documentation: "Returns the inverse cosine (degree) of X",
	},
	{
		Label:         "atan",
		Detail:        "atan X",
		Kind:          3,
		Documentation: "Returns the inverse tangent (degree) of X",
	},
}

// DefaultNololCompletions contains completion-items for nolol-builtins
var DefaultNololCompletions = []lsp.CompletionItem{
	{
		Label: "if",
		Kind:  14,
	},
	{
		Label: "then",
		Kind:  14,
	},
	{
		Label: "else",
		Kind:  14,
	},
	{
		Label: "end",
		Kind:  14,
	},
	{
		Label: "goto",
		Kind:  14,
	},
	{
		Label: "while",
		Kind:  14,
	},
	{
		Label: "do",
		Kind:  14,
	},
	{
		Label: "continue",
		Kind:  14,
	},
	{
		Label: "break",
		Kind:  14,
	},
	{
		Label: "wait",
		Kind:  14,
	},
	{
		Label: "define",
		Kind:  14,
	},
	{
		Label: "include",
		Kind:  14,
	},
	{
		Label: "insert",
		Kind:  14,
	},
	{
		Label: "macro",
		Kind:  14,
	},
	{
		Label: "_if",
		Kind:  14,
	},
	{
		Label: "_goto",
		Kind:  14,
	},
	{
		Label:         "not",
		Detail:        "not X",
		Kind:          24,
		Documentation: "Returns 1 if X is 0, otherwise it returns 0",
	},
	{
		Label:         "and",
		Detail:        "X and Y",
		Kind:          24,
		Documentation: "Returns true if X and Y are true",
	},
	{
		Label:         "or",
		Detail:        "X or Y",
		Kind:          24,
		Documentation: "Returns true if X or Y are true",
	},

	{
		Label:            "abs",
		Detail:           "abs(X)",
		InsertText:       "abs(${1:x})$0",
		InsertTextFormat: 2,
		Kind:             3,
		Documentation:    "Returns the absolute value of X",
	},
	{
		Label:            "sqrt",
		Detail:           "sqrt(X)",
		InsertText:       "sqrt(${1:x})$0",
		InsertTextFormat: 2,
		Kind:             3,
		Documentation:    "Returns the square-root of X",
	},
	{
		Label:            "sin",
		Detail:           "sin(X)",
		InsertText:       "sin(${1:x})$0",
		InsertTextFormat: 2,
		Kind:             3,
		Documentation:    "Returns the sine (degree) of X",
	},
	{
		Label:            "cos",
		Detail:           "cos(X)",
		InsertText:       "cos(${1:x})$0",
		InsertTextFormat: 2,
		Kind:             3,
		Documentation:    "Returns the cosine (degree) of X",
	},
	{
		Label:            "tan",
		Detail:           "tan(X)",
		InsertText:       "tan(${1:x})$0",
		InsertTextFormat: 2,
		Kind:             3,
		Documentation:    "Returns the tangent (degree) of X",
	},
	{
		Label:            "asin",
		Detail:           "asin(X)",
		InsertText:       "asin(${1:x})$0",
		InsertTextFormat: 2,
		Kind:             3,
		Documentation:    "Returns the inverse sine (degree) of X",
	},
	{
		Label:            "acos",
		Detail:           "acos(X)",
		InsertText:       "acos(${1:x})$0",
		InsertTextFormat: 2,
		Kind:             3,
		Documentation:    "Returns the inverse cosine (degree) of X",
	},
	{
		Label:            "atan",
		Detail:           "atan(X)",
		InsertText:       "atan(${1:x})$0",
		InsertTextFormat: 2,
		Kind:             3,
		Documentation:    "Returns the inverse tangent (degree) of X",
	},
}
