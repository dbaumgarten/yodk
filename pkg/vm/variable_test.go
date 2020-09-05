package vm

import (
	"math"
	"testing"
)

func TestVariableFromString(t *testing.T) {
	var1 := VariableFromString("abc")
	if !var1.IsString() {
		t.Fatal("var1 should be a string")
	}
	if var1.String() != "abc" {
		t.Fatal("var1 has wrong value")
	}

	var2 := VariableFromString("123")
	if !var2.IsNumber() {
		t.Fatal("var2 should be a number")
	}
	if var2.Number().Int() != 123 {
		t.Fatal("var2 has wrong value")
	}

	var3 := VariableFromString("123.5")
	if !var3.IsNumber() {
		t.Fatal("var2 should be a number")
	}
	floatval := var3.Number().Float64()
	if math.Abs(floatval-123.5) > 0.000001 {
		t.Fatal("var3 has wrong value")
	}

	var4 := VariableFromString("\"123\"")
	if !var4.IsString() {
		t.Fatal("var4 should be a string")
	}
	if var4.String() != "123" {
		t.Fatal("var4 has wrong value")
	}
}

func TestVariableFromType(t *testing.T) {
	var1, err := VariableFromType("abc")
	if err != nil {
		t.Fatal(err)
	}
	if !var1.IsString() {
		t.Fatal("var1 should be a string")
	}
	if var1.String() != "abc" {
		t.Fatal("var1 has wrong value")
	}

	var2, err := VariableFromType(123)
	if err != nil {
		t.Fatal(err)
	}
	if !var2.IsNumber() {
		t.Fatal("var2 should be a number")
	}
	if var2.Number().Int() != 123 {
		t.Fatal("var2 has wrong value")
	}

	var3, err := VariableFromType(123.5)
	if err != nil {
		t.Fatal(err)
	}
	if !var3.IsNumber() {
		t.Fatal("var2 should be a number")
	}
	floatval := var3.Number().Float64()
	if math.Abs(floatval-123.5) > 0.000001 {
		t.Fatal("var3 has wrong value")
	}

	_, err = VariableFromType([]string{})
	if err == nil {
		t.Fatal("No error for invalid type")
	}
}

func TestVariableTypeHandling(t *testing.T) {
	var1 := VariableFromString("abc")
	var2 := VariableFromString("123")
	var3 := VariableFromString("123.5")
	var4 := VariableFromString("\"123\"")

	if var2.Itoa() != "123" {
		t.Fatal("Itoa not working")
	}

	if !var1.SameType(var4) {
		t.Fatal("SameType not working")
	}

	if !var2.SameType(var3) {
		t.Fatal("SameType not working")
	}
}

func TestEquals(t *testing.T) {
	if !VariableFromString("abc").Equals(VariableFromString("abc")) {
		t.Fatal("Equals failed")
	}
	if !VariableFromString("123.5").Equals(VariableFromString("123.5")) {
		t.Fatal("Equals failed")
	}
	if VariableFromString("abc").Equals(VariableFromString("abcd")) {
		t.Fatal("Equals failed")
	}
	if VariableFromString("123").Equals(VariableFromString("\"123\"")) {
		t.Fatal("Equals failed")
	}
}

func TestOutputs(t *testing.T) {
	var1 := VariableFromString("abc")
	if var1.String() != "abc" || var1.Repr() != "\"abc\"" || var1.Itoa() != "" {
		t.Fatal("String output not working")
	}
	var2 := VariableFromString("123")
	if var2.String() != "" || var2.Repr() != "123" || var2.Itoa() != "123" {
		t.Fatal("Number output not working")
	}
}
