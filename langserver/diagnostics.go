package langserver

import (
	"context"

	"github.com/dbaumgarten/yodk/lsp"
	"github.com/dbaumgarten/yodk/parser"
)

func (s *LangServer) Diagnose(ctx context.Context, uri lsp.DocumentURI, text string) {

	go func() {

		p := parser.NewParser()
		_, err := p.Parse(text)
		diags := make([]lsp.Diagnostic, 0)

		if err != nil {
			if pError, is := err.(*parser.ParserError); is {
				diag := lsp.Diagnostic{
					Source:   "parser",
					Message:  pError.Message,
					Severity: lsp.SeverityError,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      float64(pError.StartPosition.Line) - 1,
							Character: float64(pError.StartPosition.Coloumn),
						},
						End: lsp.Position{
							Line:      float64(pError.EndPosition.Line) - 1,
							Character: float64(pError.EndPosition.Coloumn),
						},
					},
				}
				diags = append(diags, diag)
			}
		}

		s.client.PublishDiagnostics(ctx, &lsp.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diags,
		})

	}()
}
