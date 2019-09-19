package vm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/shopspring/decimal"

	"github.com/dbaumgarten/yodk/parser"
)

var errAbortLine = fmt.Errorf("")

const (
	StateIdle    = iota
	StateRunning = iota
	StatePaused  = iota
	StateDone    = iota
	StateStep    = iota
)

type BreakpointFunc func(vm *YololVM) bool

type ErrorHandlerFunc func(vm *YololVM, err error) bool

type FinishHandlerFunc func(vm *YololVM)

type RuntimeError struct {
	Base error
	Node parser.Node
}

func (e RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error at %s (up to %s): %s", e.Node.Start(), e.Node.End(), e.Base.Error())
}

type YololVM struct {
	variables         map[string]*Variable
	loop              bool
	breakpointHandler BreakpointFunc
	errorHandler      ErrorHandlerFunc
	finishHandler     FinishHandlerFunc

	parser *parser.Parser
	// currentLine is 1-indexed
	currentLine int
	state       int
	breakpoints map[int]bool
	program     *parser.Programm
	lock        *sync.Mutex
	skipBp      bool
}

func NewYololVM() *YololVM {
	decimal.DivisionPrecision = 3
	return &YololVM{
		variables:   make(map[string]*Variable),
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
		if value.IsString() {
			txt += fmt.Sprintf("%s: '%s'\n", key, value.String())
		} else if value.IsNumber() {
			txt += fmt.Sprintf("%s: %s\n", key, value.Number())
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
	v.variables = make(map[string]*Variable)
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

func (v *YololVM) Step() error {
	v.lock.Lock()
	if v.program == nil {
		v.lock.Unlock()
		err := fmt.Errorf("can not resume. Execution has not been started")
		if v.errorHandler != nil {
			v.errorHandler(v, err)
		}
		return err
	}
	v.state = StateStep
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

func (v *YololVM) GetVariable(name string) (*Variable, bool) {
	v.lock.Lock()
	defer v.lock.Unlock()
	val, exists := v.variables[name]
	if exists {
		return &Variable{
			Value: val.Value,
		}, true
	}
	return nil, false
}

func (v *YololVM) GetVariables() map[string]Variable {
	v.lock.Lock()
	defer v.lock.Unlock()
	varlist := make(map[string]Variable)
	for key, value := range v.variables {
		varlist[key] = Variable{
			Value: value.Value,
		}
	}
	return varlist
}

func (v *YololVM) SetVariable(name string, value interface{}) error {
	v.lock.Lock()
	defer v.lock.Unlock()
	val := Variable{
		value,
	}
	v.variables[name] = &val
	return nil
}

func (v *YololVM) run() error {
	v.lock.Lock()
	defer v.lock.Unlock()
	for {

		// give other goroutines a chance to aquire the lock
		v.lock.Unlock()
		v.lock.Lock()

		if v.state != StateRunning && v.state != StateStep {
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
			cont := v.loop
			if v.errorHandler != nil {
				v.lock.Unlock()
				cont = v.errorHandler(v, err)
				v.lock.Lock()
			}
			if !cont {
				return err
			}
		}
		v.currentLine++
		if v.state == StateStep {
			v.state = StatePaused
			return nil
		}
	}
}

func (v *YololVM) runLine(line *parser.Line) error {
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

func (v *YololVM) runStmt(stmt parser.Statement) error {
	switch e := stmt.(type) {
	case *parser.Assignment:
		return v.runAssignment(e)
	case *parser.IfStatement:
		conditionResult, err := v.runExpr(e.Condition)
		if err != nil {
			return err
		}
		if !conditionResult.IsNumber() {
			return RuntimeError{fmt.Errorf("If-condition can not be a string"), stmt}
		}
		if !conditionResult.Number().Equal(decimal.Zero) {
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
	case *parser.GoToStatement:
		v.currentLine = e.Line - 1
		return errAbortLine
	case *parser.Dereference:
		_, err := v.runDeref(e)
		return err
	default:
		return RuntimeError{fmt.Errorf("UNKNWON-STATEMENT:%T", e), stmt}
	}
}

func (v *YololVM) runAssignment(as *parser.Assignment) error {
	var newValue *Variable
	var err error
	if as.Operator != "=" {
		binop := parser.BinaryOperation{
			Exp1: &parser.Dereference{
				Variable: as.Variable,
				Position: as.Start(),
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

func (v *YololVM) runExpr(expr parser.Expression) (*Variable, error) {
	switch e := expr.(type) {
	case *parser.StringConstant:
		return &Variable{Value: e.Value}, nil
	case *parser.NumberConstant:
		num, err := decimal.NewFromString(e.Value)
		if err != nil {
			return nil, err
		}
		return &Variable{Value: num}, nil
	case *parser.BinaryOperation:
		return v.runBinOp(e)
	case *parser.UnaryOperation:
		return v.runUnaryOp(e)
	case *parser.Dereference:
		return v.runDeref(e)
	case *parser.FuncCall:
		return v.runFuncCall(e)
	default:
		return nil, RuntimeError{fmt.Errorf("UNKNWON-EXPRESSION:%T", e), expr}
	}
}

func (v *YololVM) runFuncCall(d *parser.FuncCall) (*Variable, error) {
	arg, err := v.runExpr(d.Argument)
	if err != nil {
		return nil, err
	}
	return RunFunction(arg, d.Function)
}

func (v *YololVM) runDeref(d *parser.Dereference) (*Variable, error) {
	oldval, exists := v.variables[d.Variable]
	if !exists {
		return nil, RuntimeError{fmt.Errorf("Variable %s used before assignment", d.Variable), d}
	}
	var newval Variable
	if oldval.IsNumber() {
		switch d.Operator {
		case "":
			return oldval, nil
		case "++":
			newval.Value = oldval.Number().Add(decimal.NewFromFloat(1))
			break
		case "--":
			newval.Value = oldval.Number().Sub(decimal.NewFromFloat(1))
			break
		default:
			return nil, RuntimeError{fmt.Errorf("Unknown operator '%s'", d.Operator), d}
		}
		v.variables[d.Variable] = &newval
	}
	if oldval.IsString() {
		switch d.Operator {
		case "":
			return oldval, nil
		case "++":
			newval.Value = oldval.String() + " "
			break
		case "--":
			if len(oldval.String()) == 0 {
				return nil, RuntimeError{fmt.Errorf("String in variable '%s' is already empty", d.Variable), d}
			}
			newval.Value = string([]rune(oldval.String())[:len(oldval.String())-1])
			break
		default:
			return nil, RuntimeError{fmt.Errorf("Unknown operator '%s'", d.Operator), d}
		}
		v.variables[d.Variable] = &newval
	}
	if d.PrePost == "Pre" {
		return &newval, nil
	}
	return oldval, nil
}

func (v *YololVM) runBinOp(op *parser.BinaryOperation) (*Variable, error) {
	arg1, err1 := v.runExpr(op.Exp1)
	if err1 != nil {
		return nil, err1
	}
	arg2, err2 := v.runExpr(op.Exp2)
	if err2 != nil {
		return nil, err2
	}
	result, err := RunBinaryOperation(arg1, arg2, op.Operator)
	if err != nil {
		return nil, RuntimeError{err, op}
	}
	return result, err
}

func (v *YololVM) runUnaryOp(op *parser.UnaryOperation) (*Variable, error) {
	arg, err := v.runExpr(op.Exp)
	if err != nil {
		return nil, err
	}

	result, err := RunUnaryOperation(arg, op.Operator)
	if err != nil {
		return nil, RuntimeError{err, op}
	}
	return result, err
}
