package langserver

import (
	"context"
	"log"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol"

	"github.com/dbaumgarten/yodk/pkg/lsp"
	"github.com/dbaumgarten/yodk/pkg/parser"
)

func (s *LangServer) Diagnose(ctx context.Context, uri lsp.DocumentURI) {

	go func() {

		var errs error
		text, _ := s.cache.Get(uri)
		if strings.HasSuffix(string(uri), ".yolol") {
			p := parser.NewParser()
			_, errs = p.Parse(text)
		} else if strings.HasSuffix(string(uri), ".nolol") {
			conv := nolol.NewConverter()
			_, errs = conv.ConvertFromSource(text)
		} else {
			return
		}

		diags := make([]lsp.Diagnostic, 0)

		if errs == nil {
			errs = make(parser.Errors, 0)
		}
		switch e := errs.(type) {
		case parser.Errors:
			break
		case *parser.Error:
			// if it is a single error, convert it to a one-element list
			errlist := make(parser.Errors, 1)
			errlist[0] = e
			errs = errlist
		default:
			log.Printf("Unknown error type: %T\n", errs)
		}
		for _, err := range errs.(parser.Errors) {
			diag := lsp.Diagnostic{
				Source:   "parser",
				Message:  err.Message,
				Severity: lsp.SeverityError,
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      float64(err.StartPosition.Line) - 1,
						Character: float64(err.StartPosition.Coloumn) - 1,
					},
					End: lsp.Position{
						Line:      float64(err.EndPosition.Line) - 1,
						Character: float64(err.EndPosition.Coloumn) - 1,
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
