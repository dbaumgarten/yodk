package vm_test

import (
	"testing"

	"github.com/dbaumgarten/yodk/testdata"
)

func TestOperators(t *testing.T) {
	err := testdata.ExecuteTestProgram(testdata.TestProgram)
	if err != nil {
		t.Fatal(err)
	}
}
