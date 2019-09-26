package optimizers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbaumgarten/yodk/parser"
	"github.com/dbaumgarten/yodk/testdata"
)

func TestGetNextVarName(t *testing.T) {

	vno := NewVariableNameOptimizer()

	for i := 0; i < 100; i++ {

		original := fmt.Sprintf("varn%d", i)

		actual := vno.getNextVarName()
		var expected string
		switch i {
		case 0:
			expected = "a"
		case 1:
			expected = "b"
		case 25:
			expected = "z"
		case 26:
			expected = "aa"
		case 27:
			expected = "ab"
		case 28:
			expected = "ac"
		}

		if expected != "" && expected != actual {
			t.Fatalf("Wrong var name for variable number %d. Expected '%s' but found '%s'.", i, expected, actual)
		}

		vno.variableNames[original] = actual
	}
}

func TestOptName(t *testing.T) {
	vno := NewVariableNameOptimizer()

	if vno.replaceVarName(":extvar") != ":extvar" {
		t.Fatal("Replaced external var")
	}

	if vno.replaceVarName("abc") != "a" {
		t.Fatal("Wrong replacement for first variable")
	}

	if vno.replaceVarName("abcd") != "b" {
		t.Fatal("Wrong replacement for second variable")
	}

	if vno.replaceVarName("abc") != "a" {
		t.Fatal("Did not remember first variable")
	}
}

func TestVarOpt(t *testing.T) {
	p := parser.NewParser()
	parsed, err := p.Parse(testdata.TestProgram)
	if err != nil {
		t.Fatal(err)
	}
	opt := NewVariableNameOptimizer()
	err = opt.Optimize(parsed)
	if err != nil {
		t.Fatal(err)
	}

	gen := parser.YololGenerator{}
	generated, err := gen.Generate(parsed)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(generated, "pi = ") || strings.Contains(generated, "hw = ") {
		t.Fatal("Variables have not been replaced")
	}

	err = testdata.ExecuteTestProgram(generated)
	if err != nil {
		t.Fatal(err)
	}
}
