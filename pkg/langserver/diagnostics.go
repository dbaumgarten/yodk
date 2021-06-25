package langserver

import (
	"context"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/lsp"
	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/optimizers"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
	"github.com/dbaumgarten/yodk/pkg/validators"
)

// fs is a special filesystem that retrieves the main file from the cache and all
// other files from the filesystem. It is used when compiling a nolol file, as nolol files may
// depend on files from the file-system using includes
type fs struct {
	*nolol.DiskFileSystem
	ls       *LangServer
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
		ls: ls,
		DiskFileSystem: &nolol.DiskFileSystem{
			Dir: filepath.Dir(getFilePath(mainfile)),
		},
		Mainfile: string(mainfile),
	}
}

func (f fs) Get(name string) (string, error) {
	if name == f.Mainfile {
		return f.ls.cache.Get(lsp.DocumentURI(name))
	}
	return f.DiskFileSystem.Get(name)
}

func convertToErrorlist(errs error) parser.Errors {
	if errs == nil {
		return make(parser.Errors, 0)
	}
	switch e := errs.(type) {
	case parser.Errors:
		return e
	case *parser.Error:
		// if it is a single error, convert it to a one-element list
		errlist := make(parser.Errors, 1)
		errlist[0] = e
		return errlist
	default:
		log.Printf("Unknown error type: %T\n (%s)", errs, errs.Error())
		return nil
	}
}

func convertErrorsToDiagnostics(errs parser.Errors, source string, severity lsp.DiagnosticSeverity) []lsp.Diagnostic {
	diags := make([]lsp.Diagnostic, 0)

	for _, err := range errs {
		diag := convertErrorToDiagnostic(err, source, severity)
		diags = append(diags, diag)
	}

	return diags
}

func convertErrorToDiagnostic(err *parser.Error, source string, severity lsp.DiagnosticSeverity) lsp.Diagnostic {
	return lsp.Diagnostic{
		Source:   source,
		Message:  err.Message,
		Severity: severity,
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
}

func (s *LangServer) validateCodeLength(uri lsp.DocumentURI, text string, parsed *ast.Program) []lsp.Diagnostic {
	// check if the code-length of yolol-code is OK
	if s.settings.Yolol.LengthChecking.Mode != LengthCheckModeOff {
		lengtherror := validators.ValidateCodeLength(text)

		// check if the code is small enough after optimizing it
		if lengtherror != nil && s.settings.Yolol.LengthChecking.Mode == LengthCheckModeOptimize && parsed != nil {

			opt := optimizers.NewCompoundOptimizer()
			err := opt.Optimize(parsed)
			if err == nil {
				printer := parser.Printer{}
				optimized, err := printer.Print(parsed)
				if err == nil {
					lengtherror = validators.ValidateCodeLength(optimized)
				}
			}
		}

		if lengtherror != nil {
			err := lengtherror.(*parser.Error)
			diag := convertErrorToDiagnostic(err, "validator", lsp.SeverityWarning)
			return []lsp.Diagnostic{diag}

		}
	}
	return []lsp.Diagnostic{}
}

func (s *LangServer) validateAvailableOperations(uri lsp.DocumentURI, parsed ast.Node) []lsp.Diagnostic {
	chipType, _ := validators.AutoChooseChipType(s.settings.Yolol.ChipType, string(uri))
	err := validators.ValidateAvailableOperations(parsed, chipType)

	if err != nil {
		errors := convertToErrorlist(err)
		return convertErrorsToDiagnostics(errors, "validator", lsp.SeverityError)
	}

	return []lsp.Diagnostic{}
}

func (s *LangServer) Diagnose(ctx context.Context, uri lsp.DocumentURI) {

	go func() {
		var parserError error
		var validationDiagnostics []lsp.Diagnostic
		var diagRes DiagnosticResults
		text, _ := s.cache.Get(uri)

		prevDiag, err := s.cache.GetDiagnostics(uri)
		if err == nil {
			diagRes = *prevDiag
		}

		if strings.HasSuffix(string(uri), ".yolol") {
			p := parser.NewParser()
			var parsed *ast.Program
			parsed, parserError = p.Parse(text)

			if parsed != nil {
				diagRes.Variables = findUsedVariables(parsed)
			}

			if parserError == nil {
				validationDiagnostics = s.validateAvailableOperations(uri, parsed)
				validationDiagnostics = append(validationDiagnostics, s.validateCodeLength(uri, text, parsed)...)
			}

		} else if strings.HasSuffix(string(uri), ".nolol") {
			mainfile := string(uri)
			converter := nolol.NewConverter().LoadFileEx(mainfile, newfs(s, uri)).ProcessIncludes()
			parserError = converter.Error()

			if parserError == nil {
				intermediate := converter.GetIntermediateProgram()
				// Analyze() will mutate the ast, so we create a copy of it
				analyse := nast.CopyAst(intermediate).(*nast.Program)
				analysis, err := nolol.Analyse(analyse)
				if err == nil {
					diagRes.AnalysisReport = analysis
				}

				validationDiagnostics = s.validateAvailableOperations(uri, intermediate)
				parserError = converter.ProcessCodeExpansion().ProcessNodes().ProcessLineNumbers().ProcessFinalize().Error()
			}
		} else {
			return
		}

		s.cache.SetDiagnostics(uri, diagRes)

		parserErrors := convertToErrorlist(parserError)
		if parserErrors == nil {
			return
		}

		diags := convertErrorsToDiagnostics(parserErrors, "parser", lsp.SeverityError)
		if validationDiagnostics != nil {
			diags = append(diags, validationDiagnostics...)
		}

		s.client.PublishDiagnostics(ctx, &lsp.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diags,
		})

	}()
}
