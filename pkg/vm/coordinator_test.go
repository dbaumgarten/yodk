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
	vm1, _ := vm.CreateFromSource(prog1)
	vm2, _ := vm.CreateFromSource(prog2)

	vm1.SetCoordinator(coord)
	vm2.SetCoordinator(coord)

	vm1.SetMaxExecutedLines(10)
	vm2.SetMaxExecutedLines(10)

	vm1.Resume()
	vm2.Resume()

	coord.Run()
	coord.WaitForTermination()

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
