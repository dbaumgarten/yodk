package langserver

import (
	"context"
	"log"
	"strings"

	"github.com/dbaumgarten/yodk/nolol"

	"github.com/dbaumgarten/yodk/lsp"
	"github.com/dbaumgarten/yodk/parser"
)

func (s *LangServer) Diagnose(ctx context.Context, uri lsp.DocumentURI, text string) {

	go func() {

		var errs error
		if strings.HasSuffix(string(uri), ".yolol") {
			p := parser.NewParser()
			_, errs = p.Parse(text)
		} else if strings.HasSuffix(string(uri), ".nolol") {
			conv := nolol.NewNololConverter()
			_, errs = conv.ConvertFromSource(text)
		} else {
			return
		}

		diags := make([]lsp.Diagnostic, 0)

		if errs == nil {
			errs = make(parser.ParserErrors, 0)
		}
		switch e := errs.(type) {
		case parser.ParserErrors:
			break
		case *parser.ParserError:
			// if it is a single error, convert it to a one-element list
			errlist := make(parser.ParserErrors, 1)
			errlist[0] = e
			errs = errlist
		default:
			log.Printf("Unknown error type: %T\n", errs)
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
