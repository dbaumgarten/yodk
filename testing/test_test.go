package testing_test

import (
	"testing"

	thistesting "github.com/dbaumgarten/yodk/testing"
)

func TestTestcase(t *testing.T) {
	testcase := `scripts: 
  - name: fizbuzz.yolol
    iterations: 1
    maxlines: 10000
cases:
  - name: TestOutput
    inputs:
      number: 0
    outputs:
      out: "fizzbuzz fizz buzz fizz fizz buzz fizz fizzbuzz fizz buzz fizz fizz buzz fizz fizzbuzz fizz buzz fizz fizz buzz fizz fizzbuzz fizz buzz fizz fizz buzz fizz fizzbuzz fizz buzz fizz fizz buzz fizz fizzbuzz fizz buzz fizz fizz buzz fizz fizzbuzz fizz buzz fizz fizz buzz "
      number: 101
  - name: TestOutput2
    inputs:
      number: 99
    outputs:
       out: "fizz buzz "
`
	script := `if :out==0 then :out="" end
if not (:number<=100) then goto 7 end
if :number%3==0 and :number%5==0 then :out+="fizzbuzz " goto 6 end
if :number%3==0 then :out+="fizz " end
if :number%5==0 then :out+="buzz " end
:number++ goto 2
`
	test, err := thistesting.Parse([]byte(testcase), "")
	if err != nil {
		t.Error(err)
	}
	test.Scripts[0].Content = script
	fails := test.Run(nil)
	if len(fails) > 0 {
		t.Log("Testcase had errors but should not")
		for _, f := range fails {
			t.Log(f)
		}
		t.FailNow()
	}

	test.Cases[0].Outputs["number"] = 1337
	fails = test.Run(nil)
	if len(fails) != 1 {
		t.Fatalf("Testcase should have 1 error, but had: %d", len(fails))
	}
}

func TestTestcase2(t *testing.T) {
	testcase := `scripts: 
    - name: fizbuzz.yolol
      iterations: 1
      maxlines: 10000
    - name: fizbuzz.yolol
      iterations: 1
      maxlines: 10000
cases:
    - name: TestOutput
      inputs:
        number: 0
      outputs:
        out: "fizzbuzz fizzbuzz fizz fizz buzz buzz fizz fizz fizz fizz buzz buzz fizz fizz fizzbuzz fizzbuzz fizz fizz buzz buzz fizz fizz fizz fizz buzz buzz fizz fizz fizzbuzz fizzbuzz fizz fizz buzz buzz fizz fizz fizz fizz buzz buzz fizz fizz fizzbuzz fizzbuzz fizz fizz buzz buzz "
        number: 102
    - name: TestOutput2
      inputs:
        number: 99
      outputs:
        out: "fizz fizz "
`
	script := `if :out==0 then :out="" end
if not (:number<=100) then goto 7 end
if :number%3==0 and :number%5==0 then :out+="fizzbuzz " goto 6 end
if :number%3==0 then :out+="fizz " end
if :number%5==0 then :out+="buzz " end
:number++ goto 2
`

	script2 := `if :out==0 then :out="" end
if not (:number<=100) then goto 7 end
if :number%3==0 and :number%5==0 then :out+="fizzbuzz " goto 6 end
if :number%3==0 then :out+="fizz " end
if :number%5==0 then :out+="buzz " end
:number++ goto 2
`
	test, err := thistesting.Parse([]byte(testcase), "")
	if err != nil {
		t.Error(err)
	}
	test.Scripts[0].Content = script
	test.Scripts[1].Content = script2
	fails := test.Run(nil)
	if len(fails) > 0 {
		t.Log("Testcase had errors but should not")
		for _, f := range fails {
			t.Log(f)
		}
		t.FailNow()
	}

	test.Cases[0].Outputs["number"] = 1337
	fails = test.Run(nil)
	if len(fails) != 1 {
		t.Fatalf("Testcase should have 1 error, but had: %d", len(fails))
	}
}
