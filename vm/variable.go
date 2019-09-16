package vm

import "github.com/shopspring/decimal"

type Variable struct {
	Value interface{}
}

func (v *Variable) IsNumber() bool {
	_, isNum := v.Value.(decimal.Decimal)
	_, isNump := v.Value.(*decimal.Decimal)
	return isNum || isNump
}

func (v *Variable) IsString() bool {
	_, isStr := v.Value.(string)
	_, isStrp := v.Value.(string)
	return isStr || isStrp
}

func (v *Variable) SameType(other *Variable) bool {
	return v.IsNumber() == other.IsNumber()
}

func (v *Variable) String() string {
	if val, isString := v.Value.(string); isString {
		return val
	}
	return ""
}

func (v *Variable) Itoa() string {
	if val, isNum := v.Value.(decimal.Decimal); isNum {
		return val.String()
	}
	return v.Value.(string)
}

func (v *Variable) Number() decimal.Decimal {
	if val, isNum := v.Value.(decimal.Decimal); isNum {
		return val
	}
	return decimal.Zero
}
