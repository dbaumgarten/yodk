package vm

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Variable represents a yolol-variable during the execution
type Variable struct {
	Value interface{}
}

// IsNumber returns true if the variable represents a number
func (v *Variable) IsNumber() bool {
	_, isNum := v.Value.(decimal.Decimal)
	_, isNump := v.Value.(*decimal.Decimal)
	return isNum || isNump
}

// IsString returns true if the variable represents a string
func (v *Variable) IsString() bool {
	_, isStr := v.Value.(string)
	_, isStrp := v.Value.(string)
	return isStr || isStrp
}

// SameType returns true if the variable has the same type as the given variable
func (v *Variable) SameType(other *Variable) bool {
	return v.IsNumber() == other.IsNumber()
}

func (v *Variable) String() string {
	if val, isString := v.Value.(string); isString {
		return val
	}
	return ""
}

// Itoa returns the string-representation of the number stored in the variable
func (v *Variable) Itoa() string {
	if val, isNum := v.Value.(decimal.Decimal); isNum {
		return val.String()
	}
	if val, isInt := v.Value.(int); isInt {
		return fmt.Sprint(val)
	}
	return v.Value.(string)
}

// Number returns the value of the variable as number
func (v *Variable) Number() decimal.Decimal {
	if val, isNum := v.Value.(decimal.Decimal); isNum {
		return val
	}
	return decimal.Zero
}
