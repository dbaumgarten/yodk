package vm_test

import (
	"testing"

	"github.com/dbaumgarten/yodk/pkg/vm"
)

func TestCoordinatedExecution(t *testing.T) {
	prog1 := `
	:result = ""
	:result += "b"
	:result += "d"
	:result += "f"
	:result += "h"
	`
	prog2 := `
	:result += "a"
	:result += "c"
	:result += "e"
	:result += "g"
	`
	coord := vm.NewCoordinator()
	vm1 := vm.NewYololVMCoordinated(coord)
	vm2 := vm.NewYololVMCoordinated(coord)

	vm1.RunSource(prog1)
	vm2.RunSource(prog2)

	coord.Run()

	vm1.WaitForTermination()
	vm2.WaitForTermination()

	r1, _ := vm1.GetVariable(":result")
	r2, _ := vm1.GetVariable(":result")

	result1 := r1.String()
	result2 := r2.String()

	if result1 != result2 {
		t.Fatal("Results are not identical", result1, result2)
	}

	if result1 != "abcdefgh" {
		t.Fatalf("Wrong result for computation, wanted %s but got %s", "abcdefgh", result1)
	}
}
