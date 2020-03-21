package optimizers

import (
	"strings"
	"testing"

	"github.com/dbaumgarten/yodk/pkg/parser"
)

func TestInversionOptimization(t *testing.T) {
	cases := map[string]string{
		"a=123":                                 "a=123",
		"a=not true":                            "a= not true",
		"x=a or b":                              "x=a or b",
		"x=not (a or b)":                        "x= not (a or b)",
		"x=not a and not b":                     "x= not (a or b)",
		"x=not (a==b)":                          "x=a!=b",
		"x=not (a > b)":                         "x=a<=b",
		"x=not (a >= b)":                        "x=a<b",
		"x=not not y":                           "x=y",
		"x=not not y<3":                         "x=y<3",
		"x=not a and not b or not c and not d":  "x= not (a or b and c or d)",
		"x=not (not a and not b)":               "x=a or b",
		"x=not (not a and b)":                   "x=a or not b",
		"x=not a and not b and not c and not d": "x= not (a or b or c or d)",
	}
	p := parser.NewParser()
	o := ExpressionInversionOptimizer{}
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
