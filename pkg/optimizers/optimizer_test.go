package optimizers

import (
	"strings"
	"testing"

	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
	"github.com/dbaumgarten/yodk/pkg/testdata"
)

func TestOptimizers(t *testing.T) {
	p := parser.NewParser()
	parsed, err := p.Parse(testdata.TestProgram)
	if err != nil {
		t.Fatal(err)
	}
	opt := NewCompoundOptimizer()
	err = opt.Optimize(parsed)
	if err != nil {
		t.Fatal(err)
	}

	gen := parser.Printer{}
	generated, err := gen.Print(parsed)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.ExecuteTestProgram(generated)
	if err != nil {
		t.Fatal(err)
	}
}

func optimizationTesting(t *testing.T, o Optimizer, cases map[string]string) {
	p := parser.NewParser()
	prin := parser.Printer{}
	for in, expected := range cases {
		parsed, err := p.Parse(in)
		if err != nil {
			t.Fatalf("Error when parsing test-case '%s':%s", in, err.Error())
		}
		err = o.Optimize(parsed)
		if err != nil {
			t.Fatalf("Error when optimizing test-case '%s':%s", in, err.Error())
		}
		optimized, err := prin.Print(parsed)
		if err != nil {
			t.Fatalf("Error when printing optimized code for '%s':%s", in, err.Error())
		}
		optimized = strings.Trim(optimized, " \n")
		if optimized != expected {
			t.Fatalf("Wrong optimized output for '%s'. Wanted '%s' but got '%s'", in, expected, optimized)
		}
	}
}

func expressionOptimizationTesting(t *testing.T, o ExpressionOptimizer, cases map[string]string) {
	p := parser.NewParser()
	prin := parser.Printer{}
	for in, expected := range cases {
		parsed, err := p.Parse(in)
		if err != nil {
			t.Fatalf("Error when parsing test-case '%s':%s", in, err.Error())
		}
		parsed.Lines[0].Statements[0].(*ast.Assignment).Value = o.OptimizeExpression(parsed.Lines[0].Statements[0].(*ast.Assignment).Value)

		optimized, err := prin.Print(parsed)
		if err != nil {
			t.Fatalf("Error when printing optimized code for '%s':%s", in, err.Error())
		}
		optimized = strings.Trim(optimized, " \n")
		if optimized != expected {
			t.Fatalf("Wrong optimized output for '%s'. Wanted '%s' but got '%s'", in, expected, optimized)
		}
	}
}
