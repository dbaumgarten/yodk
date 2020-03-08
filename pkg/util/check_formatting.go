package util

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

var positionRegex = regexp.MustCompile("\"Position\":{\"Line\":\\d+,\"Coloumn\":\\d+}")
var errmsg = "Formatting failed because of a bug (%s). Please open an issue containing the yolol-code that caused this at https://github.com/dbaumgarten/yodk/issues\n\nResult of formatting:\n\n%s\n"

func getErrmsgCompile(formatted string) error {
	return fmt.Errorf(errmsg, "Formatted output is no valid program", formatted)
}

func getErrmsgEqual(formatted string) error {
	return fmt.Errorf(errmsg, "Formatted output is not semantically equivalent", formatted)
}

// CheckForFormattingErrorYolol checks if the ast before and after formatting are identical
func CheckForFormattingErrorYolol(unformatted *ast.Program, formatted string) error {
	parser := parser.NewParser()
	formattedProg, err := parser.Parse(formatted)
	if err != nil {
		return getErrmsgCompile(formatted)
	}
	checkstring1 := ComputeASTCheckstring(unformatted)
	checkstring2 := ComputeASTCheckstring(formattedProg)

	if checkstring1 != checkstring2 {
		return getErrmsgEqual(formatted)
	}

	return nil
}

// CheckForFormattingErrorNolol checks if the ast before and after formatting are identical
func CheckForFormattingErrorNolol(unformatted *nast.Program, formatted string) error {
	parser := nolol.NewParser()
	formattedProg, err := parser.Parse(formatted)
	if err != nil {
		return getErrmsgCompile(formatted)
	}
	checkstring1 := ComputeASTCheckstring(unformatted)
	checkstring2 := ComputeASTCheckstring(formattedProg)

	if checkstring1 != checkstring2 {
		return getErrmsgEqual(formatted)
	}

	return nil
}

// ComputeASTCheckstring computes a string-representation of an ast.
// Two asts are identical, if their checkstrings are identical
func ComputeASTCheckstring(prog ast.Node) string {
	chkstr, _ := json.Marshal(prog)
	// remove position dependent substrings
	chkstr = positionRegex.ReplaceAll(chkstr, []byte(""))
	return string(chkstr)
}
