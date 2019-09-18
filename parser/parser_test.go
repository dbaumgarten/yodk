package parser_test

import (
	"testing"

	"github.com/dbaumgarten/yodk/ast"
	"github.com/dbaumgarten/yodk/parser"
	"github.com/dbaumgarten/yodk/testdata"
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
