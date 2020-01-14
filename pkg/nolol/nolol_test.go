package nolol_test

import (
	"testing"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/vm"
)

var testProg = `
const fizz = "fizz"
const buzz = "buzz"
const sep = " "
const upto = 100
:out = ""
number = 0
while number<=upto do
    if number%3==0 and number%5==0 then
        :out+=fizz+buzz+sep
        goto next
    end
    if number%3==0 then
        :out+=fizz+sep
        goto next
    end
    if number%5==0 then
        :out+=buzz+sep
        goto next
    end
    :out += number + sep
    next>
    number++
end
`

func TestNolol(t *testing.T) {
	conv := nolol.NewConverter()
	prog, err := conv.ConvertFromSource(testProg)
	if err != nil {
		t.Error(err)
	}

	gen := parser.Printer{}
	code, err := gen.Print(prog)
	if err != nil {
		t.Error(err)
	}

	v := vm.NewYololVM()
	go v.RunSource(code)
	v.WaitForTermination()

	expected := "fizzbuzz 1 2 fizz 4 buzz fizz 7 8 fizz buzz 11 fizz 13 14 fizzbuzz 16 17 fizz 19 buzz fizz 22 23 fizz buzz 26 fizz 28 29 fizzbuzz 31 32 fizz 34 buzz fizz 37 38 fizz buzz 41 fizz 43 44 fizzbuzz 46 47 fizz 49 buzz fizz 52 53 fizz buzz 56 fizz 58 59 fizzbuzz 61 62 fizz 64 buzz fizz 67 68 fizz buzz 71 fizz 73 74 fizzbuzz 76 77 fizz 79 buzz fizz 82 83 fizz buzz 86 fizz 88 89 fizzbuzz 91 92 fizz 94 buzz fizz 97 98 fizz buzz "

	result, exists := v.GetVariable(":out")
	if !exists {
		t.Fatal("Output variable does not exist")
	}

	if result.String() != expected {
		t.Fatal("Output is wrong")
	}
}
