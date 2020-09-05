package vm

import (
	"fmt"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/number"
)

// RunUnaryOperation executes the given operation with the given argument and returns the result
func RunUnaryOperation(arg *Variable, operator string) (*Variable, error) {
	if !arg.IsNumber() {
		return nil, fmt.Errorf("Unary operator '%s' is only available for numbers", operator)
	}
	var result Variable
	switch strings.ToLower(operator) {
	case "-":
		result.Value = arg.Number().Mul(number.FromInt(-1))
		break
	case "not":
		if arg.Number() == number.Zero {
			result.Value = number.One
		} else {
			result.Value = number.Zero
		}
		break
	case "abs":
		result.Value = arg.Number().Abs()
		break
	case "sqrt":
		result.Value = arg.Number().Sqrt()
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
		result.Value = arg.Number().Asin()
		break
	case "acos":
		result.Value = arg.Number().Acos()
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
			var err error
			endResult.Value, err = arg1.Number().Div(arg2.Number())
			if err != nil {
				return nil, err
			}
			break
		case "%":
			var err error
			endResult.Value, err = arg1.Number().Mod(arg2.Number())
			if err != nil {
				return nil, err
			}
			break
		case "^":
			endResult.Value = arg1.Number().Pow(arg2.Number())
			break
		case "==":
			if arg1.Number() == arg2.Number() {
				endResult.Value = number.One
			} else {
				endResult.Value = number.Zero
			}
			break
		case "!=":
			if arg1.Number() != arg2.Number() {
				endResult.Value = number.One
			} else {
				endResult.Value = number.Zero
			}
			break
		case ">=":
			if arg1.Number() >= arg2.Number() {
				endResult.Value = number.One
			} else {
				endResult.Value = number.Zero
			}
			break
		case "<=":
			if arg1.Number() <= arg2.Number() {
				endResult.Value = number.One
			} else {
				endResult.Value = number.Zero
			}
			break
		case ">":
			if arg1.Number() > arg2.Number() {
				endResult.Value = number.One
			} else {
				endResult.Value = number.Zero
			}
			break
		case "<":
			if arg1.Number() < arg2.Number() {
				endResult.Value = number.One
			} else {
				endResult.Value = number.Zero
			}
			break
		case "and":
			if arg1.Number() != number.Zero && arg2.Number() != number.Zero {
				endResult.Value = number.One
			} else {
				endResult.Value = number.Zero
			}
			break
		case "or":
			if arg1.Number() != number.Zero || arg2.Number() != number.Zero {
				endResult.Value = number.One
			} else {
				endResult.Value = number.Zero
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
				endResult.Value = number.One
			} else {
				endResult.Value = number.Zero
			}
			break
		case "!=":
			if arg1.String() != arg2.String() {
				endResult.Value = number.One
			} else {
				endResult.Value = number.Zero
			}
			break
		default:
			return nil, fmt.Errorf("Unknown binary operator for strings '%s'", operator)
		}
	}
	return &endResult, nil

}
