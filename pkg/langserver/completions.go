package langserver

import (
	"strings"

	"github.com/dbaumgarten/yodk/pkg/lsp"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

var completionItemGroups = map[string][]string{
	"unaryOps":  {"abs", "sqrt", "sin", "cos", "tan", "asin", "acos", "atan", "not"},
	"keywords":  {"if", "then", "else", "end", "goto"},
	"binaryOps": {"and", "or"},
}

// see here: https://microsoft.github.io/language-server-protocol/specifications/specification-current/
var completionItemTypes = map[string]float64{
	"unaryOps":  24,
	"keywords":  14,
	"binaryOps": 24,
}

var completionItemDocs = map[string]string{
	"abs":  "abs X: Returns the absolute value of X",
	"sqrt": "sqrt X: Returns the square-root of X",
	"sin":  "sin X: Return the sine (degree) of X",
	"cos":  "cos X: Return the cosine (degree) of X",
	"tan":  "sin X: Return the tangent (degree) of X",
	"asin": "asin X: Return the inverse sine (degree) of X",
	"acos": "asin X: Return the inverse cosine (degree) of X",
	"atan": "asin X: Return the inverse tanget (degree) of X",
	"not":  "not X: Returns 1 if X is 0, otherwise it returns 0",
	"and":  "X and Y: Returns true if X and Y are true",
	"or":   "X or Y: Returns true if X or Y are true",
}

func buildDefaultCompletionItems() []lsp.CompletionItem {
	items := make([]lsp.CompletionItem, 0, 50)
	for k, v := range completionItemGroups {
		kind := completionItemTypes[k]
		for _, str := range v {
			docs, hasDocs := completionItemDocs[str]
			item := lsp.CompletionItem{
				Label: str,
				Kind:  kind,
			}
			if hasDocs {
				item.Detail = docs
			}
			items = append(items, item)
		}
	}
	return items
}

var defaultCompletionItems []lsp.CompletionItem

func init() {
	defaultCompletionItems = buildDefaultCompletionItems()
}

func (s *LangServer) GetCompletions(params *lsp.CompletionParams) (*lsp.CompletionList, error) {
	items := make([]lsp.CompletionItem, 0, len(defaultCompletionItems)+15)

	items = append(items, defaultCompletionItems...)
	items = append(items, s.getVariableCompletions(params)...)

	return &lsp.CompletionList{
		IsIncomplete: true,
		Items:        items,
	}, nil
}

func (s *LangServer) getVariableCompletions(params *lsp.CompletionParams) []lsp.CompletionItem {
	diags, err := s.cache.GetDiagnostics(params.TextDocument.URI)
	if err != nil {
		return []lsp.CompletionItem{}
	}
	if diags.Variables == nil {
		return []lsp.CompletionItem{}
	}
	items := make([]lsp.CompletionItem, len(diags.Variables))

	for i, v := range diags.Variables {
		item := lsp.CompletionItem{
			Label: v,
			Kind:  6,
		}
		if strings.HasPrefix(v, ":") {
			item.Kind = 5
		}
		items[i] = item
	}
	return items
}

// Find all variable-names that are used inside a program
func findUsedVariables(prog *ast.Program) []string {
	variables := make(map[string]bool)
	f := func(node ast.Node, visitType int) error {
		if assign, is := node.(*ast.Assignment); visitType == ast.PreVisit && is {
			variables[assign.Variable] = true
		}
		if deref, is := node.(*ast.Dereference); visitType == ast.SingleVisit && is {
			variables[deref.Variable] = true
		}
		return nil
	}
	prog.Accept(ast.VisitorFunc(f))

	vars := make([]string, 0, len(variables))
	for k := range variables {
		vars = append(vars, k)
	}

	return vars
}
