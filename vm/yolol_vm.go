package vm

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/shopspring/decimal"

	"github.com/dbaumgarten/yodk/ast"
	"github.com/dbaumgarten/yodk/parser"
)

var errAbortLine = fmt.Errorf("")

type variableType int

const (
	TypeString variableType = iota
	TypeNumber variableType = iota
)

const (
	StateIdle    = iota
	StateRunning = iota
	StatePaused  = iota
	StateDone    = iota
)

type variable struct {
	Type  variableType
	Value interface{}
}

type BreakpointFunc func(vm *YololVM) bool

type ErrorHandlerFunc func(vm *YololVM, err error) bool

type FinishHandlerFunc func(vm *YololVM)

type YololVM struct {
	variables         map[string]*variable
	loop              bool
	breakpointHandler BreakpointFunc
	errorHandler      ErrorHandlerFunc
	finishHandler     FinishHandlerFunc

	parser *parser.Parser
	// currentLine is 1-indexed
	currentLine int
	state       int
	breakpoints map[int]bool
	program     *ast.Programm
	lock        *sync.Mutex
	skipBp      bool
}

func NewYololVM() *YololVM {
	decimal.DivisionPrecision = 3
	return &YololVM{
		variables:   make(map[string]*variable),
		state:       StateIdle,
		loop:        false,
		breakpoints: make(map[int]bool),
		parser:      parser.NewParser(),
		lock:        &sync.Mutex{},
		currentLine: 1,
	}
}

func (v *YololVM) AddBreakpoint(line int) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.breakpoints[line] = true
}

func (v *YololVM) RemoveBreakpoint(line int) {
	v.lock.Lock()
	defer v.lock.Unlock()
	delete(v.breakpoints, line)
}

func (v *YololVM) PrintVariables() string {
	v.lock.Lock()
	defer v.lock.Unlock()
	txt := ""
	for key, value := range v.variables {
		if value.Type == TypeString {
			txt += fmt.Sprintf("%s: '%s'\n", key, value.Value.(string))
		} else if value.Type == TypeNumber {
			txt += fmt.Sprintf("%s: %s\n", key, value.Value.(decimal.Decimal).String())
		} else {
			txt += fmt.Sprintf("%s: Unknown type: %T\n", key, value.Value)
		}
	}
	return txt
}

func (v *YololVM) Run(prog string) error {
	ast, err := v.parser.Parse(prog)
	if err != nil {
		if v.errorHandler != nil {
			v.errorHandler(v, err)
		}
		return fmt.Errorf("Error when parsing file: %s", err.Error())
	}
	v.lock.Lock()
	v.currentLine = 1
	v.program = ast
	v.skipBp = false
	v.state = StateRunning
	v.variables = make(map[string]*variable)
	v.lock.Unlock()
	return v.run()
}

func (v *YololVM) Resume() error {
	v.lock.Lock()
	if v.program == nil {
		v.lock.Unlock()
		err := fmt.Errorf("can not resume. Execution has not been started")
		if v.errorHandler != nil {
			v.errorHandler(v, err)
		}
		return err
	}
	v.state = StateRunning
	v.lock.Unlock()
	return v.run()
}

func (v *YololVM) Pause() {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.state = StatePaused
}

func (v *YololVM) State() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.state
}

func (v *YololVM) CurrentLine() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.currentLine
}

func (v *YololVM) SetBreakpointHandler(f BreakpointFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.breakpointHandler = f
}

func (v *YololVM) SetErrorHandler(f ErrorHandlerFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.errorHandler = f
}

func (v *YololVM) SetFinishHandler(f FinishHandlerFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.finishHandler = f
}

func (v *YololVM) EnableLoop(b bool) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.loop = b
}

func (v *YololVM) ListBreakpoints() []int {
	v.lock.Lock()
	defer v.lock.Unlock()
	li := make([]int, 0, len(v.breakpoints))
	for k := range v.breakpoints {
		li = append(li, k)
	}
	return li
}

func (v *YololVM) GetVariable(name string) (interface{}, bool) {
	v.lock.Lock()
	defer v.lock.Unlock()
	val, exists := v.variables[name]
	if exists {
		return val.Value, exists
	}
	return nil, false
}

func (v *YololVM) GetVariables() map[string]interface{} {
	v.lock.Lock()
	defer v.lock.Unlock()
	varlist := make(map[string]interface{})
	for key, value := range v.variables {
		varlist[key] = value.Value
	}
	return varlist
}

func (v *YololVM) SetVariable(name string, value interface{}) error {
	v.lock.Lock()
	defer v.lock.Unlock()
	val := &variable{
		Value: value,
	}
	switch value.(type) {
	case string:
		val.Type = TypeString
		break
	case decimal.Decimal:
		val.Type = TypeNumber
	default:
		return fmt.Errorf("Unsupported variable type: %T", value)
	}
	v.variables[name] = val
	return nil
}

func (v *YololVM) run() error {
	v.lock.Lock()
	defer v.lock.Unlock()
	for {

		// give other goroutines a chance to aquire the lock
		v.lock.Unlock()
		v.lock.Lock()

		if v.state != StateRunning {
			return nil
		}

		if _, exists := v.breakpoints[v.currentLine]; exists && !v.skipBp {

			if v.breakpointHandler != nil {
				v.lock.Unlock()
				continueExecution := v.breakpointHandler(v)
				v.lock.Lock()
				if !continueExecution {
					v.state = StatePaused
					// on next resume, start on this line, but ignore the breakpoint
					v.skipBp = true
					return nil
				}
			}
		}

		v.skipBp = false

		if v.currentLine > len(v.program.Lines) {
			v.currentLine = 1
			if v.loop {
				continue
			} else {
				v.state = StateDone
				if v.finishHandler != nil {
					v.lock.Unlock()
					v.finishHandler(v)
					v.lock.Lock()
				}
				return nil
			}
		}

		// lines are counted from 1. Compensate this
		line := v.program.Lines[v.currentLine-1]
		err := v.runLine(line)
		if err != nil {
			errReport := fmt.Errorf("Runtime-Error on line %d: %s", v.currentLine, err.Error())
			cont := v.loop
			if v.errorHandler != nil {
				v.lock.Unlock()
				cont = v.errorHandler(v, errReport)
				v.lock.Lock()
			}
			if !cont {
				return errReport
			}
		}
		v.currentLine++
	}
}

func (v *YololVM) runLine(line *ast.Line) error {
	for _, stmt := range line.Statements {
		err := v.runStmt(stmt)
		if err != nil {
			//errAbortLine is returned when the line is aborted due to an if. It is not really an 'error'
			if err != errAbortLine {
				return err
			}
			return nil
		}
	}
	return nil
}

func (v *YololVM) runStmt(stmt ast.Statement) error {
	switch e := stmt.(type) {
	case *ast.Assignment:
		return v.runAssignment(e)
	case *ast.IfStatement:
		conditionResult, err := v.runExpr(e.Condition)
		if err != nil {
			return err
		}
		if conditionResult.Type != TypeNumber {
			return fmt.Errorf("If-condition can not be a string")
		}
		if !conditionResult.Value.(decimal.Decimal).Equal(decimal.Zero) {
			for _, st := range e.IfBlock {
				err := v.runStmt(st)
				if err != nil {
					return err
				}
			}
		} else if e.ElseBlock != nil {
			for _, st := range e.ElseBlock {
				err := v.runStmt(st)
				if err != nil {
					return err
				}
			}
		}
		return nil
	case *ast.GoToStatement:
		v.currentLine = e.Line - 1
		return errAbortLine
	case *ast.Dereference:
		_, err := v.runDeref(e)
		return err
	default:
		return fmt.Errorf("UNKNWON-STATEMENT:%T", e)
	}
}

func (v *YololVM) runAssignment(as *ast.Assignment) error {
	var newValue *variable
	var err error
	if as.Operator != "=" {
		binop := ast.BinaryOperation{
			Exp1: &ast.Dereference{
				Variable: as.Variable,
			},
			Exp2:     as.Value,
			Operator: strings.Replace(as.Operator, "=", "", -1),
		}
		newValue, err = v.runBinOp(&binop)
	} else {
		newValue, err = v.runExpr(as.Value)
	}
	if err != nil {
		return err
	}

	v.variables[as.Variable] = newValue
	return nil
}

func (v *YololVM) runExpr(expr ast.Expression) (*variable, error) {
	switch e := expr.(type) {
	case *ast.StringConstant:
		return &variable{Type: TypeString, Value: e.Value}, nil
	case *ast.NumberConstant:
		num, err := decimal.NewFromString(e.Value)
		if err != nil {
			return nil, err
		}
		return &variable{Type: TypeNumber, Value: num}, nil
	case *ast.BinaryOperation:
		return v.runBinOp(e)
	case *ast.UnaryOperation:
		return v.runUnaryOp(e)
	case *ast.Dereference:
		return v.runDeref(e)
	case *ast.FuncCall:
		return v.runFuncCall(e)
	default:
		return nil, fmt.Errorf("UNKNWON-EXPRESSION:%T", e)
	}
}

func (v *YololVM) runFuncCall(d *ast.FuncCall) (*variable, error) {
	arg, err := v.runExpr(d.Argument)
	if err != nil {
		return nil, err
	}
	if arg.Type != TypeNumber {
		return nil, fmt.Errorf("Function %s expects a number as argument", d.Function)
	}
	result := &variable{
		Type: TypeNumber,
	}
	switch d.Function {
	case "abs":
		result.Value = arg.Value.(decimal.Decimal).Abs()
		break
	case "sqrt":
		v, _ := arg.Value.(decimal.Decimal).Float64()
		result.Value = decimal.NewFromFloat(math.Sqrt(v))
		break
	case "sin":
		result.Value = arg.Value.(decimal.Decimal).Sin()
		break
	case "cos":
		result.Value = arg.Value.(decimal.Decimal).Cos()
		break
	case "tan":
		result.Value = arg.Value.(decimal.Decimal).Tan()
		break
	case "asin":
		v, _ := arg.Value.(decimal.Decimal).Float64()
		result.Value = decimal.NewFromFloat(math.Asin(v))
		break
	case "acos":
		v, _ := arg.Value.(decimal.Decimal).Float64()
		result.Value = decimal.NewFromFloat(math.Acos(v))
		break
	case "atan":
		result.Value = arg.Value.(decimal.Decimal).Atan()
		break
	default:
		return nil, fmt.Errorf("Unknown function: %s", d.Function)
	}
	result.Value = result.Value.(decimal.Decimal).Truncate(int32(decimal.DivisionPrecision))
	return result, nil
}

func (v *YololVM) runDeref(d *ast.Dereference) (*variable, error) {
	oldval, exists := v.variables[d.Variable]
	if !exists {
		return nil, fmt.Errorf("Variable %s used before assignment", d.Variable)
	}
	var newval variable
	if oldval.Type == TypeNumber {
		switch d.Operator {
		case "":
			return oldval, nil
		case "++":
			newval.Value = oldval.Value.(decimal.Decimal).Add(decimal.NewFromFloat(1))
			newval.Type = TypeNumber
			break
		case "--":
			newval.Value = oldval.Value.(decimal.Decimal).Sub(decimal.NewFromFloat(1))
			newval.Type = TypeNumber
			break
		default:
			return nil, fmt.Errorf("Unknown operator '%s'", d.Operator)
		}
		v.variables[d.Variable] = &newval
	}
	if oldval.Type == TypeString {
		switch d.Operator {
		case "":
			return oldval, nil
		case "++":
			newval.Value = oldval.Value.(string) + " "
			newval.Type = TypeString
			break
		case "--":
			if len(oldval.Value.(string)) == 0 {
				return nil, fmt.Errorf("String in variable '%s' is already empty", d.Variable)
			}
			newval.Value = string([]rune(oldval.Value.(string))[:len(oldval.Value.(string))-1])
			newval.Type = TypeString
			break
		default:
			return nil, fmt.Errorf("Unknown operator '%s'", d.Operator)
		}
		v.variables[d.Variable] = &newval
	}
	if d.PrePost == "Pre" {
		return &newval, nil
	}
	return oldval, nil
}

func (v *YololVM) runBinOp(op *ast.BinaryOperation) (*variable, error) {
	arg1, err1 := v.runExpr(op.Exp1)
	if err1 != nil {
		return nil, err1
	}
	arg2, err2 := v.runExpr(op.Exp2)
	if err2 != nil {
		return nil, err2
	}
	// automatic type casting
	if arg1.Type != arg2.Type {
		// do NOT modify the existing variable. Create a temporary new one
		if arg1.Type != TypeString {
			arg1 = &variable{
				Type:  TypeString,
				Value: arg1.Value.(decimal.Decimal).String(),
			}
		}
		if arg2.Type != TypeString {
			arg2 = &variable{
				Type:  TypeString,
				Value: arg2.Value.(decimal.Decimal).String(),
			}
		}
	}
	endResult := variable{
		Type: arg1.Type,
	}

	one := decimal.NewFromFloat(1)

	if arg1.Type == TypeNumber {
		switch op.Operator {
		case "+":
			endResult.Value = arg1.Value.(decimal.Decimal).Add(arg2.Value.(decimal.Decimal))
			break
		case "-":
			endResult.Value = arg1.Value.(decimal.Decimal).Sub(arg2.Value.(decimal.Decimal))
			break
		case "*":
			endResult.Value = arg1.Value.(decimal.Decimal).Mul(arg2.Value.(decimal.Decimal))
			break
		case "/":
			endResult.Value = arg1.Value.(decimal.Decimal).Div(arg2.Value.(decimal.Decimal))
			break
		case "%":
			endResult.Value = arg1.Value.(decimal.Decimal).Mod(arg2.Value.(decimal.Decimal))
			break
		case "^":
			endResult.Value = arg1.Value.(decimal.Decimal).Pow(arg2.Value.(decimal.Decimal))
			break
		case "==":
			if arg1.Value.(decimal.Decimal).Equal(arg2.Value.(decimal.Decimal)) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "!=":
			if !arg1.Value.(decimal.Decimal).Equal(arg2.Value.(decimal.Decimal)) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case ">=":
			if arg1.Value.(decimal.Decimal).GreaterThanOrEqual(arg2.Value.(decimal.Decimal)) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "<=":
			if arg1.Value.(decimal.Decimal).LessThanOrEqual(arg2.Value.(decimal.Decimal)) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case ">":
			if arg1.Value.(decimal.Decimal).GreaterThan(arg2.Value.(decimal.Decimal)) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "<":
			if arg1.Value.(decimal.Decimal).LessThan(arg2.Value.(decimal.Decimal)) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "and":
			if !arg1.Value.(decimal.Decimal).Equal(decimal.Zero) && !arg2.Value.(decimal.Decimal).Equal(decimal.Zero) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		case "or":
			if !arg1.Value.(decimal.Decimal).Equal(decimal.Zero) || !arg2.Value.(decimal.Decimal).Equal(decimal.Zero) {
				endResult.Value = one
			} else {
				endResult.Value = decimal.Zero
			}
			break
		default:
			return nil, fmt.Errorf("Unknown binary operator for numbers '%s'", op.Operator)
		}
	}

	if arg1.Type == TypeString {
		switch op.Operator {
		case "+":
			endResult.Value = arg1.Value.(string) + arg2.Value.(string)
			break
		case "-":
			lastIndex := strings.LastIndex(arg1.Value.(string), arg2.Value.(string))
			if lastIndex >= 0 {
				endResult.Value = string([]rune(arg1.Value.(string))[:lastIndex]) + string([]rune(arg1.Value.(string))[lastIndex+len(arg2.Value.(string)):])
			} else {
				endResult.Value = arg1.Value.(string)
			}
			break
		case "==":
			if arg1.Value.(string) == arg2.Value.(string) {
				endResult.Value = decimal.NewFromFloat(1)
			} else {
				endResult.Value = decimal.Zero
			}
			endResult.Type = TypeNumber
			break
		case "!=":
			if arg1.Value.(string) != arg2.Value.(string) {
				endResult.Value = decimal.NewFromFloat(1)
			} else {
				endResult.Value = decimal.Zero
			}
			endResult.Type = TypeNumber
			break
		default:
			return nil, fmt.Errorf("Unknown binary operator for strings '%s'", op.Operator)
		}
	}
	return &endResult, nil
}

func (v *YololVM) runUnaryOp(op *ast.UnaryOperation) (*variable, error) {
	arg, err := v.runExpr(op.Exp)
	if err != nil {
		return nil, err
	}
	if arg.Type != TypeNumber {
		return nil, fmt.Errorf("Unary operator '%s' is only available for numbers", op.Operator)
	}
	res := &variable{
		Type: TypeNumber,
	}
	switch op.Operator {
	case "-":
		res.Value = arg.Value.(decimal.Decimal).Mul(decimal.NewFromFloat(-1))
		break
	case "not":
		if arg.Value.(decimal.Decimal) == decimal.Zero {
			res.Value = decimal.NewFromFloat(1)
		} else {
			res.Value = decimal.Zero
		}
		break
	default:
		return nil, fmt.Errorf("Unknown unary operator for numbers '%s'", op.Operator)
	}
	return res, nil
}
