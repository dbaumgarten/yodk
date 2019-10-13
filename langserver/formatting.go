package langserver

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/dbaumgarten/yodk/lsp"
	"github.com/dbaumgarten/yodk/nolol"
	"github.com/dbaumgarten/yodk/parser"
	"github.com/pmezard/go-difflib/difflib"
)

func format(params *lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	file := string(params.TextDocument.URI)
	file = strings.Replace(file, "file://", "", 1)
	unformattedraw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	unformatted := string(unformattedraw)
	var formatted string

	if strings.HasSuffix(file, ".yolol") {
		p := parser.NewParser()
		parsed, errs := p.Parse(unformatted)
		if errs != nil {
			return nil, errs
		}
		gen := parser.Printer{}
		formatted, err = gen.Print(parsed)
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(file, ".nolol") {
		p := nolol.NewParser()
		parsed, errs := p.Parse(unformatted)
		if errs != nil {
			return nil, errs
		}
		printer := nolol.NewPrinter()
		formatted, err = printer.Print(parsed)
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Unsupported file-type:", file)
	}

	return ComputeTextEdits(unformatted, formatted), nil
}

// ComputeTextEdits computes text edits that are required to
// change the `unformatted` to the `formatted` text.
// Blatantly stolen from https://github.com/sourcegraph/go-langserver/blob/master/langserver/format.go
func ComputeTextEdits(unformatted string, formatted string) []lsp.TextEdit {
	// LSP wants a list of TextEdits. We use difflib to compute a
	// non-naive TextEdit. Originally we returned an edit which deleted
	// everything followed by inserting everything. This leads to a poor
	// experience in vscode.
	unformattedLines := strings.Split(unformatted, "\n")
	formattedLines := strings.Split(formatted, "\n")
	m := difflib.NewMatcher(unformattedLines, formattedLines)
	var edits []lsp.TextEdit
	for _, op := range m.GetOpCodes() {
		switch op.Tag {
		case 'r': // 'r' (replace):  a[i1:i2] should be replaced by b[j1:j2]
			edits = append(edits, lsp.TextEdit{
				Range: lsp.Range{
					Start: lsp.Position{
						Line: float64(op.I1),
					},
					End: lsp.Position{
						Line: float64(op.I2),
					},
				},
				NewText: strings.Join(formattedLines[op.J1:op.J2], "\n") + "\n",
			})
		case 'd': // 'd' (delete):   a[i1:i2] should be deleted, j1==j2 in this case.
			edits = append(edits, lsp.TextEdit{
				Range: lsp.Range{
					Start: lsp.Position{
						Line: float64(op.I1),
					},
					End: lsp.Position{
						Line: float64(op.I2),
					},
				},
			})
		case 'i': // 'i' (insert):   b[j1:j2] should be inserted at a[i1:i1], i1==i2 in this case.
			edits = append(edits, lsp.TextEdit{
				Range: lsp.Range{
					Start: lsp.Position{
						Line: float64(op.I1),
					},
					End: lsp.Position{
						Line: float64(op.I1),
					},
				},
				NewText: strings.Join(formattedLines[op.J1:op.J2], "\n") + "\n",
			})
		}
	}

	return edits
}
