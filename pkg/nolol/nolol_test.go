package nolol_test

import (
	"testing"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/vm"
)

var testProg = `
define fizz = "fizz"
define buzz = "buzz"
define sep = " "
define upto = 100
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

var testProg2 = `
zzz=7
a=1;b=2;c=3 $
x="blabla"
y="hey"
$
$
foo="bar"
$ what="ever"
$ z = 0 $
x = 99
`

var testProg3 = `
include "testProg"
`

var testfs = nolol.MemoryFileSystem{
	"testProg":  testProg,
	"testProg2": testProg2,
	"testProg3": testProg3,
}

func TestNolol(t *testing.T) {
	conv := nolol.NewConverter()
	prog, err := conv.ConvertFileEx("testProg", testfs)
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
		t.Fatal("Output is wrong:", result.String())
	}
}

func TestInclude(t *testing.T) {
	conv := nolol.NewConverter()
	prog, _ := conv.ConvertFileEx("testProg", testfs)
	prog2, _ := conv.ConvertFileEx("testProg3", testfs)
	printer := &parser.Printer{}
	printed, _ := printer.Print(prog)
	printed2, _ := printer.Print(prog2)
	if printed != printed2 {
		t.Fatal("Include does not match original")
	}
}

func TestLineHandling(t *testing.T) {
	conv := nolol.NewConverter()
	prog, err := conv.ConvertFileEx("testProg2", testfs)
	if err != nil {
		t.Error(err)
	}

	lines := len(prog.Lines)

	if lines != 8 {
		t.Fatal("Wrong amount of lines after merging. Expected 8, but got: ", lines)
	}
}
