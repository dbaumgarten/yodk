package testdata

import (
	"fmt"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/vm"
	"github.com/shopspring/decimal"
)

var TestProgram = `:testsum = 1 + 2 == 3
:testsub = 3 - 1 == 2
:testmul = 2*5 == 10
:testdiv = 20 / 10 == 2
:testmod = 11 % 10 == 1
counter=0
counter++
if counter < 20 then goto 7 end
:testgoto = counter == 20
// comment1
:testexp = 10^2 == 100
:testeq = 42 == 42 and not (41 == 24)
:testneq = 1 != 42 and not (1!=1)
:testgt = 2 > 1 and not (1>2) and 5 > -5
:testgte = 2 >= 1 and not (1 >= 2) and 2 >= 2
:done = 1
`

func ExecuteTestProgram(prog string) error {
	var err error

	v, _ := vm.CreateFromSource(prog)
	v.SetErrorHandler(vm.ErrorHandlerFunc(func(v *vm.VM, e error) bool {
		err = e
		return true
	}))

	v.SetVariableChangedHandler(vm.TerminateOnDoneVar)
	v.Resume()
	v.WaitForTermination()

	if err != nil {
		return err
	}

	if len(v.GetVariables()) == 0 {
		return fmt.Errorf("Program not executed")
	}

	for name, value := range v.GetVariables() {
		if strings.HasPrefix(name, ":test") {
			if !value.IsNumber() {
				return fmt.Errorf("Operator-test %s returend string '%s' instead of 1", name, value.String())
			}
			if !value.Number().Equal(decimal.NewFromFloat(1)) {
				return fmt.Errorf("Operator-test %s failed", name)
			}
		}
	}
	return nil
}
