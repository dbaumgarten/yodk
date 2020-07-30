package vm

import (
	"fmt"
	"math"
	"strings"

	"github.com/shopspring/decimal"
)

// RunUnaryOperation executes the given operation with the given argument and returns the result
func RunUnaryOperation(arg *Variable, operator string) (*Variable, error) {
	if !arg.IsNumber() {
		return nil, fmt.Errorf("Unary operator '%s' is only available for numbers", operator)
	}
	var result Variable
	switch strings.ToLower(operator) {
	case "-":
		result.Value = arg.Number().Mul(decimal.NewFromFloat(-1))
		break
	case "not":
		if arg.Number().Equal(decimal.Zero) {
			result.Value = decimal.NewFromFloat(1)
		} else {
			result.Value = decimal.Zero
		}
		break
	case "abs":
		result.Value = arg.Number().Abs()
		break
	case "sqrt":
		v, _ := arg.Number().Float64()
		result.Value = decimal.NewFromFloat(math.Sqrt(v))
		break
	case "sin":
		result.Value = arg.Number().Sin()
		break
	case "cos":
		result.Value = arg.Number().Cos()
		break
	case "tan":
		result.Value = arg.Number().Tan()
		break
	case "asin":
		v, _ := arg.Number().Float64()
		result.Value = decimal.NewFromFloat(math.Asin(v))
		break
	case "acos":
		v, _ := arg.Number().Float64()
		result.Value = decimal.NewFromFloat(math.Acos(v))
		break
	case "atan":
		result.Value = arg.Number().Atan()
		break
	default:
		return nil, fmt.Errorf("Unknown unary operator for numbers '%s'", operator)
	}
	return &result, nil
}

// RunBinaryOperation executes the given operation with the given arguments and returns the result
func RunBinaryOperation(arg1 *Variable, arg2 *Variable, operator string) (*Variable, error) {
	// automatic type casting
	if !arg1.SameType(arg2) {
		// do NOT modify the existing variable. Create a temporary new one
		if !arg1.IsString() {
			arg1 = &Variable{
				Value: arg1.Itoa(),
			}
		}
		if !arg2.IsString() {
			arg2 = &Variable{
				Value: arg2.Itoa(),
			}
		}
	}
	var endResult Variable

	one := decimal.NewFromFloat(1)

	if arg1.IsNumber() {
		switch operator {
		case "+":
			endResult.Value = arg1.Number().Add(arg2.Number())
			break
		case "-":
			endResult.Value = arg1.Number().Sub(arg2.Number())
			break
		case "*":
			endResult.Value = arg1.Number().Mul(arg2.Number())
			break
		case "/":
			if arg2.Number().IsZero() {
				return nil, fmt.Errorf("Can not divide by 0")
			}
			endResult.Value = arg1.Number().Div(arg2.Number())
			break
		case "%":
			endResult.Value = arg1.Number().Mod(arg2.Number())
			break
		case "^":
			endResult.Value = arg1.Number().Pow(arg2.Number())
			break
		case "==":
			if arg1.Number().Equal(arg2.Number()) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "!=":
			if !arg1.Number().Equal(arg2.Number()) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case ">=":
			if arg1.Number().GreaterThanOrEqual(arg2.Number()) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "<=":
			if arg1.Number().LessThanOrEqual(arg2.Number()) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case ">":
			if arg1.Number().GreaterThan(arg2.Number()) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "<":
			if arg1.Number().LessThan(arg2.Number()) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "and":
			if !arg1.Number().Equal(decimal.Zero) && !arg2.Number().Equal(decimal.Zero) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "or":
			if !arg1.Number().Equal(decimal.Zero) || !arg2.Number().Equal(decimal.Zero) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		default:
			return nil, fmt.Errorf("Unknown binary operator for numbers '%s'", operator)
		}

	}

	if arg1.IsString() {
		switch operator {
		case "+":
			endResult.Value = arg1.String() + arg2.String()
			break
		case "-":
			lastIndex := strings.LastIndex(arg1.String(), arg2.String())
			if lastIndex >= 0 {
				endResult.Value = string([]rune(arg1.String())[:lastIndex]) + string([]rune(arg1.String())[lastIndex+len(arg2.String()):])
			} else {
				endResult.Value = arg1.String()
			}
			break
		case "==":
			if arg1.String() == arg2.String() {
				endResult.Value = decimal.NewFromFloat(1)
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "!=":
			if arg1.String() != arg2.String() {
				endResult.Value = decimal.NewFromFloat(1)
			} else {
				endResult.Value = decimal.Zero
			}
			break
		default:
			return nil, fmt.Errorf("Unknown binary operator for strings '%s'", operator)
		}
	}
	return &endResult, nil

}
