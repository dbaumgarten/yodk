package parser_test

import (
	"testing"

	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
	"github.com/dbaumgarten/yodk/pkg/testdata"
)

func TestParser(t *testing.T) {

	p := parser.NewParser()

	result, err := p.Parse(testdata.TestProgram)

	if err != nil {
		t.Fatal(err)
	}

	if len(result.Lines) == 0 {
		t.Fatal("Parsed programm is empty")
	}
}

func TestParserMultipleErrors(t *testing.T) {

	prog := `
	a = b + c c++ x=sin(x)
	a = b++c c-- b+-
	x = y + z
	y = if x then y=1 else z=1 end
	if x then y=1 else z=1 end
	if x then y=1 else z=1
	`

	p := parser.NewParser()

	result, errs := p.Parse(prog)

	if errs != nil && len(errs.(parser.Errors)) != 3 {
		for _, err := range errs.(parser.Errors) {
			t.Log(err)
		}
		t.Fatalf("Found %d errors instead of %d", len(errs.(parser.Errors)), 3)
	}

	if result != nil && len(result.Lines) == 0 {
		t.Fatal("Parsed programm is empty")
	}
}

type nodePositionTester struct {
	*testing.T
}

func (o *nodePositionTester) Visit(node ast.Node, visitType int) error {
	if visitType == ast.PreVisit || visitType == ast.SingleVisit {
		startPos := node.Start()
		if startPos.Line == 0 && startPos.Coloumn == 0 {
			o.Fatalf("Empty position for %T", node)
		}
	}
	return nil
}

func TestNodePositions(t *testing.T) {
	tester := nodePositionTester{t}

	p := parser.NewParser()

	result, err := p.Parse(testdata.TestProgram)
	if err != nil {
		t.Fatal(err)
	}

	result.Accept(&tester)
}
