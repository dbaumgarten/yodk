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
	StateStep    = iota
	StateKill    = iota
	StateDone    = iota
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

// errKillVM is a special error used to terminate the vm-goroutine using panic/recover
var errKillVM = fmt.Errorf("Kill this vm")

// YololVM is a virtual machine to execute YOLOL-Code
type YololVM struct {
	// the current variables of the programm
	variables map[string]*Variable
	// if true, restart programm after executing line 20
	loop              bool
	breakpointHandler BreakpointFunc
	errorHandler      ErrorHandlerFunc
	finishHandler     FinishHandlerFunc

	// current line in the ast is 1-indexed
	currentAstLine int
	// current line in the source code
	currentSourceLine int
	// current state of the vm
	state int
	// list of active breakpoints
	breakpoints map[int]bool
	// the parsed program
	program *parser.Program
	// a lock to synchronize acces to the vms state
	lock *sync.Mutex
	// line number of a breakpoint to skip
	skipBp int
	// condition to wait on while VM is paused
	waitCondition *sync.Cond
	// true while there is a gorouting executing for this vm
	running bool
	// the coordinator to use for coordinating execution with other VMs
	coordinator *Coordinator
}

// NewYololVM creates a new standalone VM
func NewYololVM() *YololVM {
	return NewYololVMCoordinated(nil)
}

// NewYololVMCoordinated creates a new VM that is coordinated with other VMs using the given coordinator
func NewYololVMCoordinated(coordinator *Coordinator) *YololVM {
	decimal.DivisionPrecision = 3
	vm := &YololVM{
		variables:         make(map[string]*Variable),
		state:             StateIdle,
		loop:              false,
		breakpoints:       make(map[int]bool),
		lock:              &sync.Mutex{},
		currentAstLine:    1,
		currentSourceLine: 1,
		skipBp:            -1,
		coordinator:       coordinator,
	}
	vm.waitCondition = sync.NewCond(vm.lock)
	if coordinator != nil {
		coordinator.registerVM(vm)
	}
	return vm
}

// Getters and Setters ---------------------------------

// AddBreakpoint adds a breakpoint at the line. Breakpoint-lines always refer to the position recorded in the
// ast nodes, not the position of the Line in the Line-Slice of parser.Program.
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

// Running returns true if there is a running goroutine for this vm
func (v *YololVM) Running() bool {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.running
}

// Resume resumes execution after a breakpoint or pause()
func (v *YololVM) Resume() error {
	v.lock.Lock()
	if !v.running {
		v.lock.Unlock()
		err := fmt.Errorf("can not resume. Execution has not been started")
		if v.errorHandler != nil {
			v.errorHandler(v, err)
		}
		return err
	}
	v.state = StateRunning
	v.waitCondition.Signal()
	v.lock.Unlock()
	return nil
}

// Step executes the next line and stops the execution again
func (v *YololVM) Step() error {
	v.lock.Lock()
	if !v.running {
		v.lock.Unlock()
		err := fmt.Errorf("can not resume. Execution has not been started")
		if v.errorHandler != nil {
			v.errorHandler(v, err)
		}
		return err
	}
	v.state = StateStep
	v.waitCondition.Signal()
	v.lock.Unlock()
	return nil
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

// CurrentSourceLine returns the current (=next to be executed) source line of the program
func (v *YololVM) CurrentSourceLine() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.currentSourceLine
}

// CurrentAstLine returns the current (=next to be executed) ast line of the program
func (v *YololVM) CurrentAstLine() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.currentAstLine
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
	if v.coordinator != nil {
		globals := v.coordinator.getVariables()
		for key, value := range globals {
			varlist[key] = Variable{
				Value: value.Value,
			}
		}
	}
	return varlist
}

// GetVariable gets the current state of a variable
func (v *YololVM) GetVariable(name string) (*Variable, bool) {
	v.lock.Lock()
	defer v.lock.Unlock()
	val, exists := v.getVariable(name)
	if exists {
		return &Variable{
			Value: val.Value,
		}, true
	}
	return nil, false
}

// SetVariable sets the current state of a variable
func (v *YololVM) SetVariable(name string, value interface{}) error {
	v.lock.Lock()
	defer v.lock.Unlock()
	// do return only a copy of the variable
	val := &Variable{
		value,
	}
	return v.setVariable(name, val)
}

// getVariable gets the current state of a variable.
// Does not use the lock. ONLY USE WHEN LOCK IS ALREADY HELD
func (v *YololVM) getVariable(name string) (*Variable, bool) {
	if v.coordinator != nil && strings.HasPrefix(name, ":") {
		return v.coordinator.getVariable(name)
	}
	val, exists := v.variables[name]
	return val, exists
}

// setVariable sets the current state of a variable
// Does not use the lock. ONLY USE WHEN LOCK IS ALREADY HELD
func (v *YololVM) setVariable(name string, value *Variable) error {
	if v.coordinator != nil && strings.HasPrefix(name, ":") {
		return v.coordinator.setVariable(name, value)
	}
	v.variables[name] = value
	return nil
}

// Terminate the vm goroutine (if running)
func (v *YololVM) Terminate() {
	v.lock.Lock()
	if v.running {
		v.state = StateKill
		v.waitCondition.Signal()
		for v.state != StateDone {
			v.waitCondition.Wait()
		}
		v.running = false
	}
	v.lock.Unlock()
}

// WaitForTermination blocks until the vm finished running
func (v *YololVM) WaitForTermination() {
	v.lock.Lock()
	for v.state != StateDone {
		v.waitCondition.Wait()
	}
	v.lock.Unlock()
}

// Begin main section ------------------------------------------

// Run runs the compiled program prog in a new go-routine
func (v *YololVM) Run(prog *parser.Program) {
	v.Terminate()
	v.lock.Lock()
	v.running = true
	v.currentAstLine = 1
	v.currentSourceLine = 1
	v.program = prog
	v.skipBp = -1
	// do not reset to running so we can start paused
	//v.state = StateRunning
	v.variables = make(map[string]*Variable)
	v.lock.Unlock()
	go v.run()
}

// RunSource compiles and runs the given YOLOL code
func (v *YololVM) RunSource(prog string) {
	ast, err := parser.NewParser().Parse(prog)
	if err != nil {
		if v.errorHandler != nil {
			v.errorHandler(v, err)
		}
		return
	}

	v.Run(ast)
}

func (v *YololVM) wait() {
	if v.state == StateKill {
		panic(errKillVM)
	}
	v.state = StatePaused
	for v.state == StatePaused {
		v.waitCondition.Wait()
	}
	// vm should be terminated
	if v.state == StateKill {
		panic(errKillVM)
	}
}

func (v *YololVM) run() {
	defer func() {
		err := recover()
		if err != nil && err != errKillVM {
			panic(err)
		}
		v.state = StateDone
		v.running = false
		v.waitCondition.Signal()
	}()

	v.lock.Lock()
	defer v.lock.Unlock()
	for {

		// give other goroutines a chance to aquire the lock
		v.lock.Unlock()
		v.lock.Lock()

		// the vm should terminate
		if v.state == StateKill {
			panic(errKillVM)
		}

		// the vm should pause now
		if v.state == StatePaused {
			v.wait()
		}

		if v.currentAstLine > len(v.program.Lines) {
			v.currentAstLine = 1
			if v.loop {
				continue
			} else {
				if v.finishHandler != nil {
					v.lock.Unlock()
					v.finishHandler(v)
					v.lock.Lock()
				}
				panic(errKillVM)
			}
		}

		// lines are counted from 1. Compensate this
		line := v.program.Lines[v.currentAstLine-1]
		err := v.runLine(line)
		if err != nil {
			cont := v.loop
			if v.errorHandler != nil {
				v.lock.Unlock()
				cont = v.errorHandler(v, err)
				v.lock.Lock()
			}
			if !cont {
				v.wait()
			}
		}
		v.currentAstLine++
	}
}

func (v *YololVM) runLine(line *parser.Line) error {

	// wait until the coordinator allows this VM to run a line
	if v.coordinator != nil {
		// give up the lock while waiting for our turn to execute a line
		// to allow the retrieval of variables while waiting
		v.lock.Unlock()
		v.coordinator.waitForTurn(v)
		v.lock.Lock()
		defer v.coordinator.finishTurn()
	}

	for _, stmt := range line.Statements {
		statementLine := stmt.Start().Line
		v.currentSourceLine = statementLine
		if v.state == StateStep {
			v.wait()
		}
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

	// did we hit a breakpoint?
	if _, exists := v.breakpoints[v.currentSourceLine]; exists && v.skipBp != v.currentSourceLine {
		v.skipBp = v.currentSourceLine
		if v.breakpointHandler != nil {
			v.lock.Unlock()
			continueExecution := v.breakpointHandler(v)
			v.lock.Lock()
			if !continueExecution {
				v.wait()
			}
		}
	}

	// reached the end of the breakpoint-line. Reset skipBp.
	if v.currentSourceLine != v.skipBp {
		v.skipBp = -1
	}

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
		v.currentAstLine = e.Line - 1
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
	v.setVariable(as.Variable, newValue)
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
	oldval, exists := v.getVariable(d.Variable)
	if !exists {
		// uninitialized variables have a default value of 0
		oldval = &Variable{
			decimal.Zero,
		}
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
		v.setVariable(d.Variable, &newval)
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
		v.setVariable(d.Variable, &newval)
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
