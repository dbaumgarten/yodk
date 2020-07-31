package langserver

import (
	"context"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/lsp"
	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/optimizers"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
	"github.com/dbaumgarten/yodk/pkg/validators"
)

// fs is a special filesystem that retrieves the main file from the cache and all
// other files from the filesystem. It is used when compiling a nolol file, as nolol files may
// depend on files from the file-system using includes
type fs struct {
	*LangServer
	Dir      string
	Mainfile string
}

func getFilePath(u lsp.DocumentURI) string {
	ur, _ := url.Parse(string(u))
	s := filepath.FromSlash(ur.Path)

	if !strings.HasSuffix(s, "\\\\") {
		s = strings.TrimPrefix(s, "\\")
	}
	return s
}

func newfs(ls *LangServer, mainfile lsp.DocumentURI) *fs {
	return &fs{
		LangServer: ls,
		Dir:        filepath.Dir(getFilePath(mainfile)),
		Mainfile:   string(mainfile),
	}
}

func (f fs) Get(name string) (string, error) {
	if name == f.Mainfile {
		return f.cache.Get(lsp.DocumentURI(name))
	}
	path := filepath.Join(f.Dir, name)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), err
}

func (s *LangServer) Diagnose(ctx context.Context, uri lsp.DocumentURI) {

	go func() {
		var errs error
		var parsed *ast.Program
		text, _ := s.cache.Get(uri)

		if strings.HasSuffix(string(uri), ".yolol") {
			p := parser.NewParser()
			parsed, errs = p.Parse(text)
		} else if strings.HasSuffix(string(uri), ".nolol") {
			conv := nolol.NewConverter()
			mainfile := string(uri)
			_, errs = conv.ConvertFileEx(mainfile, newfs(s, uri))
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
			log.Printf("Unknown error type: %T\n (%s)", errs, errs.Error())
			return
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

		// check if the code-length of yolol-code is OK
		if len(diags) == 0 && s.settings.Yolol.LengthChecking.Mode != LengthCheckModeOff && strings.HasSuffix(string(uri), ".yolol") {
			lengtherror := validators.ValidateCodeLength(text)

			// check if the code is small enough after optimizing it
			if lengtherror != nil && s.settings.Yolol.LengthChecking.Mode == LengthCheckModeOptimize && parsed != nil {

				opt := optimizers.NewCompoundOptimizer()
				err := opt.Optimize(parsed)
				if err == nil {
					printer := parser.Printer{
						Mode: parser.PrintermodeCompact,
					}
					optimized, err := printer.Print(parsed)
					if err == nil {
						lengtherror = validators.ValidateCodeLength(optimized)
					}
				}
			}

			if lengtherror != nil {
				err := lengtherror.(*parser.Error)
				diag := lsp.Diagnostic{
					Source:   "validator",
					Message:  err.Message,
					Severity: lsp.SeverityWarning,
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

		}

		s.client.PublishDiagnostics(ctx, &lsp.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diags,
		})

	}()
}
