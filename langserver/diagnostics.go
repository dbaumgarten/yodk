package langserver

import (
	"context"

	"github.com/dbaumgarten/yodk/lsp"
	"github.com/dbaumgarten/yodk/parser"
)

func (s *LangServer) Diagnose(ctx context.Context, uri lsp.DocumentURI, text string) {

	go func() {

		p := parser.NewParser()
		_, errs := p.Parse(text)
		diags := make([]lsp.Diagnostic, 0)

		if errs == nil {
			errs = make(parser.ParserErrors, 0)
		}
		for _, err := range errs.(parser.ParserErrors) {
			diag := lsp.Diagnostic{
				Source:   "parser",
				Message:  err.Message,
				Severity: lsp.SeverityError,
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      float64(err.StartPosition.Line) - 1,
						Character: float64(err.StartPosition.Coloumn),
					},
					End: lsp.Position{
						Line:      float64(err.EndPosition.Line) - 1,
						Character: float64(err.EndPosition.Coloumn),
					},
				},
			}
			diags = append(diags, diag)

		}

		s.client.PublishDiagnostics(ctx, &lsp.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diags,
		})

	}()
}
