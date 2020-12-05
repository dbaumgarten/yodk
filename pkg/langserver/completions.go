package langserver

import (
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/lsp"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

func (s *LangServer) GetCompletions(params *lsp.CompletionParams) (*lsp.CompletionList, error) {
	var items []lsp.CompletionItem

	if strings.HasSuffix(string(params.TextDocument.URI), ".yolol") {
		items = make([]lsp.CompletionItem, 0, len(DefaultYololCompletions)+15)
		items = append(items, DefaultYololCompletions...)
		items = append(items, s.getVariableCompletions(params)...)
	} else if strings.HasSuffix(string(params.TextDocument.URI), ".nolol") {
		items = make([]lsp.CompletionItem, 0, len(DefaultNololCompletions)+30)
		items = append(items, DefaultNololCompletions...)
		items = append(items, s.getNololCompletions(params)...)
	} else {
		return nil, nil
	}

	return &lsp.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

func (s *LangServer) getNololCompletions(params *lsp.CompletionParams) []lsp.CompletionItem {
	diags, err := s.cache.GetDiagnostics(params.TextDocument.URI)
	if err != nil {
		return []lsp.CompletionItem{}
	}

	analysis := diags.AnalysisReport
	if analysis == nil {
		return []lsp.CompletionItem{}
	}

	items := make([]lsp.CompletionItem, len(analysis.Variables)+len(analysis.Definitions)+len(analysis.Macros))

	for _, v := range analysis.GetVarsAtLine(int(params.Position.Line) + 1) {
		items = append(items, lsp.CompletionItem{
			Label: v,
			Kind:  6,
		})
	}

	for _, m := range analysis.Macros {
		item := lsp.CompletionItem{
			Detail:           m.Name + "(" + strings.Join(m.Arguments, ",") + ")",
			Label:            m.Name,
			Kind:             15,
			InsertText:       m.Name + argsToSnippet(m.Arguments),
			InsertTextFormat: 2,
		}
		if doc, exists := analysis.Docstrings[m.Name]; exists {
			item.Documentation = doc
		}
		items = append(items, item)
	}

	for _, d := range analysis.Definitions {
		kind := 21.0
		insert := ""
		detail := ""
		if len(d.Placeholders) > 0 {
			kind = 3.0
			detail += d.Name + "(" + strings.Join(d.Placeholders, ",") + ")"
			insert = d.Name + argsToSnippet(d.Placeholders)
		}
		item := lsp.CompletionItem{
			Label:            d.Name,
			Kind:             kind,
			InsertText:       insert,
			Detail:           detail,
			InsertTextFormat: 2,
		}
		if doc, exists := analysis.Docstrings[d.Name]; exists {
			item.Documentation = doc
		}
		items = append(items, item)
	}

	for _, l := range analysis.Labels {
		items = append(items, lsp.CompletionItem{
			Label: l,
			Kind:  13,
		})
	}

	return items
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

func argsToSnippet(args []string) string {
	snip := "("
	for i, arg := range args {
		snip += "${" + strconv.Itoa(i+1) + ":" + arg + "}"
		if i != len(args)-1 {
			snip += ","
		}
	}
	snip += ")$0"
	return snip
}
