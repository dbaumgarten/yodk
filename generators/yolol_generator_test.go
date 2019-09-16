package generators

import (
	"testing"

	"github.com/dbaumgarten/yodk/parser"
)

var testProgramm = `
testsum = 1 + 2 == 3
testsub = 3 - 1 == 2
testmul = 2*5 == 10
testdiv = 20 / 10 == 2
testmod = 11 % 10 == 1
testexp = 10^2 == 100
testeq = 42 == 42 and not (41 == 24)
testneq = 1 != 42 and not (1!=1)
testgt = 2 > 1 and not (1>2) and 5 > -5
testgte = 2 >= 1 and not (1 >= 2) and 2 >= 2
testlt = 1 < 2 and not (2<1) and -5 < 5
testlte = 1 <= 2 and not(2 <= 1) and 2 >= 2
i = 0 j = i++ k = ++i
testi = i == 2 and j == 0 and  k == 2
hw = "hello" hw++
hw += "world"
testhw = hw == "hello world"
abc = "abc"
testssub = hw - "world" == "hello " and abc-- == "abc" and abc == "ab" and --abc == "a" and abc == "a"
pi = 3.141
testsinpi = abs(sin(pi) - 0)<0.01
testcospi = abs(cos(pi) - -1)<0.01
testtanpi = abs(tan(pi) - 0)<0.01
testasin = abs(asin(sin(pi)) - 0) < 0.1
testacos = abs(acos(cos(pi)) - pi) < 0.1
testatan = abs(atan(tan(pi)) - 0) < 0.1
testsq = sqrt(16) == 4
testab = abs(-5) == 5 and abs(5) == 5
testz = (not 1 and not 10 and not not 0) == 0
testor = 20 or 0
testif = 0
if pi > 3 then testif=1 else testif = 0 end
counter=0
counter++
if counter < 20 then goto 35 end
testgoto = counter == 20
testnested = 3+(1+1)*5 == 13
k = 2
testnestedop = (k + 5)*k++ == 14
`

func TestGenerator(t *testing.T) {
	p := parser.NewParser()
	gen := YololGenerator{}
	parsed, err := p.Parse(testProgramm)
	if err != nil {
		t.Fatal(err)
	}
	generated := gen.Generate(parsed)
	newparsed, err := p.Parse(generated)
	if err != nil {
		t.Fatal(err)
	}
	if len(newparsed.Lines) == 0 {
		t.Fatal("Generated program has no lines")
	}
}
