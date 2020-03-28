package optimizers

import (
	"testing"

	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/testdata"
)

var staticCases = map[string]string{
	"a=123+100":               "a=223",
	"a=b+100":                 "a=b+100",
	"a=123+100+a":             "a=223+a",
	"a=a+(123+100)+b":         "a=a+223+b",
	"a=a+(123+100)+b*(10*10)": "a=a+223+b*100",
}

func TestStaticExpressions(t *testing.T) {
	p := parser.NewParser()
	parsed, err := p.Parse(testdata.TestProgram)
	if err != nil {
		t.Fatal(err)
	}
	opt := StaticExpressionOptimizer{}
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

func TestStaticExpressions2(t *testing.T) {
	optimizationTesting(t, &StaticExpressionOptimizer{}, staticCases)
}
