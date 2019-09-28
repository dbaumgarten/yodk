package vm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/shopspring/decimal"

	"github.com/dbaumgarten/yodk/parser"
)

var errAbortLine = fmt.Errorf("")

// The current state of the VM
const (
	StateIdle    = iota
	StateRunning = iota
	StatePaused  = iota
	StateDone    = iota
	StateStep    = iota
)

// BreakpointFunc is a function that is called when a breakpoint is encountered.
// If true is returned the execution is resumed. Otherwise the vm remains paused
type BreakpointFunc func(vm *YololVM) bool

// ErrorHandlerFunc is a function that is called when a runtime-error is encountered.
// If true is returned the execution is resumed. Otherwise the vm remains paused
type ErrorHandlerFunc func(vm *YololVM, err error) bool

// FinishHandlerFunc is a function that is called when the programm finished execution (looping is disabled)
type FinishHandlerFunc func(vm *YololVM)

// RuntimeError represents an error encountered during execution
type RuntimeError struct {
	Base error
	// The Node that caused the error
	Node parser.Node
}

func (e RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error at %s (up to %s): %s", e.Node.Start(), e.Node.End(), e.Base.Error())
}

// YololVM is a virtual machine to execute YOLOL-Code
type YololVM struct {
	// the current variables of the programm
	variables map[string]*Variable
	// if true, restart programm after executing line 20
	loop              bool
	breakpointHandler BreakpointFunc
	errorHandler      ErrorHandlerFunc
	finishHandler     FinishHandlerFunc

	// the parser to use
	parser *parser.Parser
	// currentLine is 1-indexed
	currentLine int
	// current state of the vm
	state int
	// list of active breakpoints
	breakpoints map[int]bool
	// the parsed program
	program *parser.Programm
	// a lock to synchronize acces to the vms state
	lock *sync.Mutex
	// needed to resume after hitting a breakpoint
	skipBp bool
}

// NewYololVM creates a new VM
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

// AddBreakpoint adds a breakpoint at the line
func (v *YololVM) AddBreakpoint(line int) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.breakpoints[line] = true
}

// RemoveBreakpoint removes the breakpoint at the line
func (v *YololVM) RemoveBreakpoint(line int) {
	v.lock.Lock()
	defer v.lock.Unlock()
	delete(v.breakpoints, line)
}

// PrintVariables gets an overview over the current variable state
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

// Run compiles and runs the given programm code
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

// Resume resumes execution after a breakpoint or pause()
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

// Step executes the next line and stops the execution again
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

// Pause pauses the execution
func (v *YololVM) Pause() {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.state = StatePaused
}

// State returns the current vm state
func (v *YololVM) State() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.state
}

// CurrentLine returns the current (=next to be executed) line of the program
func (v *YololVM) CurrentLine() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.currentLine
}

// SetBreakpointHandler sets the function to be called when hitting a breakpoint
func (v *YololVM) SetBreakpointHandler(f BreakpointFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.breakpointHandler = f
}

// SetErrorHandler sets the function to be called when encountering an error
func (v *YololVM) SetErrorHandler(f ErrorHandlerFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.errorHandler = f
}

// SetFinishHandler sets the function to be called when execution finishes
func (v *YololVM) SetFinishHandler(f FinishHandlerFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.finishHandler = f
}

// EnableLoop wheter or not to loop from line 20 back to 1
func (v *YololVM) EnableLoop(b bool) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.loop = b
}

// ListBreakpoints returns the list of active breakpoints
func (v *YololVM) ListBreakpoints() []int {
	v.lock.Lock()
	defer v.lock.Unlock()
	li := make([]int, 0, len(v.breakpoints))
	for k := range v.breakpoints {
		li = append(li, k)
	}
	return li
}

// GetVariable gets the current state of a variable
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

// GetVariables gets the current state of all variables
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

// SetVariable sets the current state of a variable
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

		// the vm should pause now. Stop the loop
		if v.state != StateRunning && v.state != StateStep {
			return nil
		}

		// did we hit a breakpoint?
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
