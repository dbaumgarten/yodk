package nolol_test

import (
	"fmt"
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
:done = 1
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

var testProg4 = `
define ev = :mm
define cv = @cc
cv++
ev=cv
`

var testfs = nolol.MemoryFileSystem{
	"testProg.nolol":  testProg,
	"testProg2.nolol": testProg2,
	"testProg3.nolol": testProg3,
	"testProg4.nolol": testProg4,
}

func TestNolol(t *testing.T) {
	conv := nolol.NewConverter()
	prog, err := conv.LoadFileEx("testProg.nolol", testfs).Convert()
	if err != nil {
		t.Fatal(err)
	}

	gen := parser.Printer{}
	code, err := gen.Print(prog)
	if err != nil {
		t.Error(err)
	}

	v, err := vm.CreateFromSource(code)
	if err != nil {
		t.Error(err)
	}
	v.SetLineExecutedHandler(vm.TerminateOnDoneVar)
	v.Resume()
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
	prog, _ := conv.LoadFileEx("testProg.nolol", testfs).Convert()
	prog2, _ := conv.LoadFileEx("testProg3.nolol", testfs).Convert()
	printer := &parser.Printer{}
	printed, _ := printer.Print(prog)
	printed2, _ := printer.Print(prog2)
	if printed != printed2 {
		fmt.Println(printed)
		fmt.Println("------------")
		fmt.Println(printed2)
		t.Fatal("Include does not match original")
	}
}

func TestLineHandling(t *testing.T) {
	conv := nolol.NewConverter()
	prog, err := conv.LoadFileEx("testProg2.nolol", testfs).Convert()
	if err != nil {
		t.Error(err)
	}

	lines := len(prog.Lines)

	if lines != 8 {
		t.Fatal("Wrong amount of lines after merging. Expected 8, but got: ", lines)
	}
}

func TestVariableNames(t *testing.T) {
	conv := nolol.NewConverter()
	file := conv.LoadFileEx("testProg4.nolol", testfs)
	prog, err := file.Convert()

	printer := &parser.Printer{}
	actual, _ := printer.Print(prog)

	var expected = "cc++ :mm=cc goto1"
	if actual != expected {
		t.Fatal("Output is wrong:", actual)
	}
	if err != nil {
		t.Error(err)
	}
}
