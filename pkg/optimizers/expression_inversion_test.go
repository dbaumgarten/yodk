package optimizers

import (
	"testing"
)

var inversionCases = map[string]string{
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
	"x=:number%3==0 and :number%5==0":       "x=:number%3==0 and :number%5==0",
	"x=not(:number%3==0 and :number%5==0)":  "x=:number%3!=0 or :number%5!=0",
}

func TestInversionOptimization(t *testing.T) {
	o := &ExpressionInversionOptimizer{}
	optimizationTesting(t, o, inversionCases)
}

func TestInversionOptimizationExp(t *testing.T) {
	o := &ExpressionInversionOptimizer{}
	expressionOptimizationTesting(t, o, inversionCases)
}
