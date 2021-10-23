package optimizers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/testdata"
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

		vno.variableMappings[original] = actual
	}
}

func TestOptName(t *testing.T) {
	vno := NewVariableNameOptimizer()

	vno.SetBlacklist([]string{"c"})

	if vno.OptimizeVarName(":extvar") != ":extvar" {
		t.Fatal("Replaced external var")
	}

	if vno.OptimizeVarName("abc") != "a" {
		t.Fatal("Wrong replacement for first variable")
	}

	if vno.OptimizeVarName("aBc") != "a" {
		t.Fatal("Wrong replacement for other cased variable")
	}

	if vno.OptimizeVarName("abcd") != "b" {
		t.Fatal("Wrong replacement for second variable")
	}

	if vno.OptimizeVarName("abc") != "a" {
		t.Fatal("Did not remember first variable")
	}

	if vno.OptimizeVarName("abcde") != "d" {
		t.Fatal("Did not honor blacklisting")
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

	gen := parser.Printer{}
	generated, err := gen.Print(parsed)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(generated, " pi") || strings.Contains(generated, " hw") {
		t.Fatal("Variables have not been replaced", generated)
	}

	err = testdata.ExecuteTestProgram(generated)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitializeByFrequency(t *testing.T) {
	testprog := `
foo=123
bar=456
bar=555
baz=aaa+aaa+aaa
ignoreme=1
other=2
	`
	p := parser.NewParser()
	parsed, err := p.Parse(testprog)
	if err != nil {
		t.Fatal(err)
	}
	opt := NewVariableNameOptimizer()

	opt.InitializeByFrequency(parsed, []string{"ignoreme"})

	if opt.OptimizeVarName("aaa") != "a" {
		t.Fatal("Wrong replacement for aaa variable")
	}

	if opt.OptimizeVarName("bar") != "b" {
		t.Fatal("Wrong replacement for bar variable")
	}

	if opt.OptimizeVarName("foo") != "c" {
		t.Fatal("Wrong replacement for foo variable")
	}

	if opt.OptimizeVarName("baz") != "d" {
		t.Fatal("Wrong replacement for baz variable")
	}

	if opt.OptimizeVarName("other") != "e" {
		t.Fatal("Wrong replacement for baz variable")
	}

}
