package vm

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// VariableFromString tries to create a variable of the correct type from the given string.
// If the string is enclosed in quotes, the string between the quotes is used as string-value for the variable.
// Else, it tries to parse the given string into a number. If that also fails, the plain given string is used as value.
func VariableFromString(str string) *Variable {
	var value interface{}
	if strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"") && len(str) >= 2 {
		value = str[1 : len(str)-1]
	} else {
		deci, err := decimal.NewFromString(str)
		if err == nil {
			value = deci
		} else {
			value = str
		}
	}
	return &Variable{
		Value: value,
	}
}

// VariableFromType creates a new variable from the given input. The type of the variable is decided by the input-type.
func VariableFromType(inp interface{}) (*Variable, error) {
	var value interface{}
	switch v := inp.(type) {
	case string:
		value = v
	case *string:
		value = *v
	case int:
		value = decimal.NewFromInt(int64(v))
	case int32:
		value = decimal.NewFromInt32(v)
	case int64:
		value = decimal.NewFromInt(v)
	case float32:
		value = decimal.NewFromFloat32(v)
	case float64:
		value = decimal.NewFromFloat(v)
	case decimal.Decimal:
		value = v
	default:
		return nil, fmt.Errorf("Can not convert type %T to variable", inp)
	}
	return &Variable{
		Value: value,
	}, nil
}

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

// TypeName returns the name of the type this variable has
func (v *Variable) TypeName() string {
	if v.IsString() {
		return "string"
	}
	return "number"
}

// Equals checks if this variable equals another variable
func (v *Variable) Equals(other *Variable) bool {
	if !v.SameType(other) {
		return false
	}
	if v.IsString() {
		return v.String() == other.String()
	}
	if v.IsNumber() {
		return v.Number().Equal(other.Number())
	}
	return false
}

func (v *Variable) String() string {
	if val, isString := v.Value.(string); isString {
		return val
	}
	return ""
}

// Repr returns the string-representation of the variable.
// If the variable is of type string, its value is enclosed in quotes.
func (v *Variable) Repr() string {
	if v.IsNumber() {
		return v.Itoa()
	}
	return "\"" + v.String() + "\""
}

// Itoa returns the string-representation of the number stored in the variable
func (v *Variable) Itoa() string {
	if val, isNum := v.Value.(decimal.Decimal); isNum {
		return val.String()
	}
	return ""
}

// Number returns the value of the variable as number
func (v *Variable) Number() decimal.Decimal {
	if val, isNum := v.Value.(decimal.Decimal); isNum {
		return val
	}
	return decimal.Zero
}
