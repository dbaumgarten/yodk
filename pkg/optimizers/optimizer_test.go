package optimizers

import (
	"testing"

	"github.com/dbaumgarten/yodk/pkg/parser"
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
