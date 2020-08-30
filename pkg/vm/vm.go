package vm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/shopspring/decimal"

	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

var errAbortLine = fmt.Errorf("")

// The current state of the VM
const (
	StatePaused     = iota
	StateRunning    = iota
	StateStepping   = iota
	StateTerminated = iota
)

// BreakpointFunc is a function that is called when a breakpoint is encountered.
// If true is returned the execution is resumed. Otherwise the vm remains paused
type BreakpointFunc func(vm *VM) bool

// ErrorHandlerFunc is a function that is called when a runtime-error is encountered.
// If true is returned the execution is resumed. Otherwise the vm remains paused
type ErrorHandlerFunc func(vm *VM, err error) bool

// FinishHandlerFunc is a function that is called when the programm finished execution (looping is disabled)
type FinishHandlerFunc func(vm *VM)

// VarChangedHandlerFunc is the type for functions that react on variable changes
// If true is returned the execution is resumed. Otherwise the vm is paused
type VarChangedHandlerFunc func(vm *VM, name string, value *Variable) bool

// TerminateOnDoneVar is a predefined VarChangedHandlerFunc that can be used to terminate the VM once :done is set to 1
var TerminateOnDoneVar = func(vm *VM, name string, value *Variable) bool {
	if name == ":done" {
		go vm.Terminate()
		return false
	}
	return true
}

// RuntimeError represents an error encountered during execution
type RuntimeError struct {
	Base error
	// The Node that caused the error
	Node ast.Node
}

func (e RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error at %s (up to %s): %s", e.Node.Start(), e.Node.End(), e.Base.Error())
}

// errKillVM is a special error used to terminate the vm-goroutine using panic/recover
var errKillVM = fmt.Errorf("Kill this vm")

// VM is a virtual machine to execute YOLOL-Code
type VM struct {
	// the parsed program
	program *ast.Program
	// a lock to synchronize acces to variables
	lock *sync.Mutex
	// the current variables of the programm
	variables map[string]*Variable
	// event handlers
	breakpointHandler BreakpointFunc
	stepHandler       FinishHandlerFunc
	errorHandler      ErrorHandlerFunc
	finishHandler     FinishHandlerFunc
	varChangedHandler VarChangedHandlerFunc
	// current line in the ast is 1-indexed
	currentAstLine int
	// current line in the source code
	currentSourceLine int
	// currerent coloumn in the current source line
	currentSourceColoumn int
	// if true we arrived at the current line via a goto
	jumped bool
	// list of active breakpoints
	breakpoints map[int]bool
	// current state of the vm
	state int
	// this channel is used to comminucate state-change-requests
	stateRequests chan int
	// this channel is closed once the VM terminates
	terminationChannel chan interface{}
	// if set we are running in coordinated mode
	coordinator *Coordinator
	// this channel is obtained from the coordinator and queried for permission to run a line
	coordinatorPermission <-chan struct{}
	// this channel is used to signal to the coordinator that we finished running a line
	coordinatorDone chan<- struct{}
	// if != 0 and in the current iteration more the x lines are run, terminate VM
	maxExecutedLines int
	// number of lines executed in the current run
	executedLines int
}

// Create creates a new VM to run the given program in a seperate goroutine.
// The returned VM is paused. Configure it using the setters and then call Resume()
func Create(prog *ast.Program) *VM {
	vm := &VM{
		variables:          make(map[string]*Variable),
		state:              StatePaused,
		breakpoints:        make(map[int]bool),
		lock:               &sync.Mutex{},
		currentAstLine:     1,
		currentSourceLine:  1,
		stateRequests:      make(chan int),
		terminationChannel: make(chan interface{}),
		program:            prog,
	}
	go vm.run()
	return vm
}

// CreateFromSource creates a new VM to run the given program in a seperate goroutine.
// The returned VM is paused. Configure it using the setters and then call Resume()
func CreateFromSource(prog string) (*VM, error) {
	ast, err := parser.NewParser().Parse(prog)
	if err != nil {
		return nil, err
	}
	return Create(ast), nil
}

// Getters and Setters ---------------------------------

// SetMaxExecutedLines sets the maximum number of lines to run
// if the amount is reached the VM terminates
// Can be used to prevent blocking by endless loops
// <= 0 disables this. Default is 0
func (v *VM) SetMaxExecutedLines(lines int) {
	v.maxExecutedLines = lines
}

// GetExecutedLines returns the number of lines executed by this VM
func (v *VM) GetExecutedLines() int {
	return v.executedLines
}

// AddBreakpoint adds a breakpoint at the line. Breakpoint-lines always refer to the position recorded in the
// ast nodes, not the position of the Line in the Line-Slice of ast.Program.
func (v *VM) AddBreakpoint(line int) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.breakpoints[line] = true
}

// RemoveBreakpoint removes the breakpoint at the line
func (v *VM) RemoveBreakpoint(line int) {
	v.lock.Lock()
	defer v.lock.Unlock()
	delete(v.breakpoints, line)
}

// PrintVariables gets an overview over the current variable state
func (v *VM) PrintVariables() string {
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

// Resume resumes execution after a breakpoint or pause() or the initial stopped state
// Blocks until the VM reacts on the request
func (v *VM) Resume() {
	v.requestState(StateRunning)
}

// Step executes the next line and paused the execution
// Blocks until the VM reacts on the request
func (v *VM) Step() {
	v.requestState(StateStepping)
}

// Pause pauses the execution
// Blocks until the VM reacts on the request
func (v *VM) Pause() {
	v.requestState(StatePaused)
}

// State returns the current vm state
func (v *VM) State() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.state
}

// CurrentSourceLine returns the current (=next to be executed) source line of the program
func (v *VM) CurrentSourceLine() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.currentSourceLine
}

// CurrentSourceColoumn returns the current (=next to be executed) source column of the program
func (v *VM) CurrentSourceColoumn() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.currentSourceColoumn
}

// CurrentAstLine returns the current (=next to be executed) ast line of the program
func (v *VM) CurrentAstLine() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.currentAstLine
}

// SetBreakpointHandler sets the function to be called when hitting a breakpoint
func (v *VM) SetBreakpointHandler(f BreakpointFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.breakpointHandler = f
}

// SetStepHandler sets the function to be called when a step completes
func (v *VM) SetStepHandler(f FinishHandlerFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.stepHandler = f
}

// SetErrorHandler sets the function to be called when encountering an error
func (v *VM) SetErrorHandler(f ErrorHandlerFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.errorHandler = f
}

// SetFinishHandler sets the function to be called when execution finishes
func (v *VM) SetFinishHandler(f FinishHandlerFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.finishHandler = f
}

// SetVariableChangedHandler registers a callback that is executed when the value of a variable changes
func (v *VM) SetVariableChangedHandler(handler VarChangedHandlerFunc) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.varChangedHandler = handler
}

// SetCoordinator sets the coordinator that is used to coordinate execution with other vms
func (v *VM) SetCoordinator(c *Coordinator) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.coordinator = c
	v.coordinatorPermission, v.coordinatorDone = c.registerVM(v)
}

// ListBreakpoints returns the list of active breakpoints
func (v *VM) ListBreakpoints() []int {
	v.lock.Lock()
	defer v.lock.Unlock()
	li := make([]int, 0, len(v.breakpoints))
	for k := range v.breakpoints {
		li = append(li, k)
	}
	return li
}

// GetVariables gets the current state of all variables
func (v *VM) GetVariables() map[string]Variable {
	v.lock.Lock()
	defer v.lock.Unlock()
	varlist := make(map[string]Variable)
	for key, value := range v.variables {
		varlist[key] = Variable{
			Value: value.Value,
		}
	}
	if v.coordinator != nil {
		globals := v.coordinator.GetVariables()
		for key, value := range globals {
			varlist[key] = Variable{
				Value: value.Value,
			}
		}
	}
	return varlist
}

// GetVariable gets the current state of a variable
func (v *VM) GetVariable(name string) (*Variable, bool) {
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
func (v *VM) SetVariable(name string, value *Variable) error {
	v.lock.Lock()
	defer v.lock.Unlock()
	// do return only a copy of the variable
	val := &Variable{
		value.Value,
	}
	return v.setVariable(name, val)
}

// GetProgram returns the program that is run by the VM
func (v *VM) GetProgram() *ast.Program {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.program
}

// getVariable gets the current state of a variable.
// Does not use the lock. ONLY USE WHEN LOCK IS ALREADY HELD
// getting variables is case-insensitive
func (v *VM) getVariable(name string) (*Variable, bool) {
	name = strings.ToLower(name)
	if v.coordinator != nil && strings.HasPrefix(name, ":") {
		return v.coordinator.GetVariable(name)
	}
	val, exists := v.variables[name]
	return val, exists
}

// setVariable sets the current state of a variable
// Does not use the lock. ONLY USE WHEN LOCK IS ALREADY HELD
// setting variables is case-insensitive
func (v *VM) setVariable(name string, value *Variable) error {
	name = strings.ToLower(name)
	v.variables[name] = value
	if v.coordinator != nil && strings.HasPrefix(name, ":") {
		err := v.coordinator.SetVariable(name, value)
		if err != nil {
			return err
		}
	}

	if v.varChangedHandler != nil {
		v.lock.Unlock()
		cont := v.varChangedHandler(v, name, value)
		v.lock.Lock()
		if !cont {
			v.pause()
		}

	}

	return nil
}

// Terminate the vm goroutine (if running)
func (v *VM) Terminate() {
	v.requestState(StateTerminated)
}

// WaitForTermination blocks until the VM terminates
func (v *VM) WaitForTermination() {
	<-v.terminationChannel
}

// Begin main section ------------------------------------------

// this function request a state-change from the worker goroutine.
// blocks until the worker picks up the request
// called from outside go.routines
func (v *VM) requestState(state int) {
	v.stateRequests <- state
}

// called by the worker goroutine to receive state-change-requests
func (v *VM) receiveState() {
	select {
	case requested := <-v.stateRequests:
		v.changeState(requested)
	default:
		// no state-change-requests
	}
}

// set a new state for the VM. perform necessary actions or new state. called by worker
func (v *VM) changeState(requested int) {
	v.state = requested
	if requested == StateTerminated {
		panic(errKillVM)
	}
	if requested == StatePaused {
		v.pause()
	}
}

// sets the state to StatePaused and blocks until another state is requested
// unocks the lock while paused
func (v *VM) pause() {
	v.state = StatePaused
	for {
		v.lock.Unlock()
		newstate := <-v.stateRequests
		v.lock.Lock()
		if newstate != StatePaused {
			v.changeState(newstate)
			break
		}
	}
}

func (v *VM) run() {

	v.lock.Lock()
	defer v.lock.Unlock()

	defer func() {
		err := recover()
		if err != nil && err != errKillVM {
			panic(err)
		}
		v.state = StateTerminated
		close(v.stateRequests)
		close(v.terminationChannel)
		if v.coordinatorDone != nil {
			close(v.coordinatorDone)
		}
		if v.finishHandler != nil {
			v.finishHandler(v)
		}
	}()

	v.pause()

	// compensate the increment on the first iteration of the loop
	// necessary because of the pesky 1-indexing of lines
	v.currentAstLine--

	for {
		v.currentAstLine++

		// give other goroutines a chance to aquire the lock
		v.lock.Unlock()
		v.lock.Lock()

		// roll back to line 1
		if v.currentAstLine > 20 {
			v.currentAstLine = 1
		}

		if v.currentAstLine-1 < len(v.program.Lines) {
			// lines are counted from 1. Compensate this when indexing the array
			line := v.program.Lines[v.currentAstLine-1]
			err := v.runLine(line)
			if err != nil {
				if v.errorHandler != nil {
					v.lock.Unlock()
					cont := v.errorHandler(v, err)
					v.lock.Lock()
					if !cont {
						v.pause()
					}
				} else {
					// no error handler. Kill VM.
					panic(errKillVM)
				}
			}
		} else {
			// nothing to to but to trigger a line-change notification
			v.currentSourceLine = v.currentAstLine
			v.currentSourceColoumn = 0
			v.sourceLineChanged()
		}

		v.executedLines++
		if v.maxExecutedLines > 0 && v.executedLines > v.maxExecutedLines {
			panic(errKillVM)
		}
	}
}

// called when execution advances to a new source-line
// advancing to a new source-line could trigger a step, a breakpoint or a state-change
func (v *VM) sourceLineChanged() {

	if v.state == StateStepping {
		if v.stepHandler != nil {
			v.lock.Unlock()
			v.stepHandler(v)
			v.lock.Lock()
		}
		v.pause()
	}

	// check if we hit a breakpoint
	if _, exists := v.breakpoints[v.currentSourceLine]; exists {
		if v.breakpointHandler != nil {
			v.lock.Unlock()
			continueExecution := v.breakpointHandler(v)
			v.lock.Lock()
			if !continueExecution {
				v.pause()
			}
		}
	}

	// the state can only be changed when one source-line completed (or inside pause())
	v.receiveState()
}

// check if the statement that is to be executed is on a different line then the previous one
func (v *VM) checkSourceLineChanged(stmt ast.Statement) {
	if stmt.Start().File == "" && (stmt.Start().Line != v.currentSourceLine || v.jumped) {
		v.jumped = false
		v.currentSourceLine = stmt.Start().Line
		v.sourceLineChanged()
	}
}

// get the permission from the coordinator to run the next line
// also react on state-change-requests while waiting for permission
func (v *VM) aquireCoordinatorPermission() {
	for {
		// release the lock while waiting for permission
		// but re-aquire it later

		v.lock.Unlock()
		select {
		case <-v.coordinatorPermission:
			v.lock.Lock()
			return
		case statechange := <-v.stateRequests:
			v.lock.Lock()
			v.changeState(statechange)
		}
	}
}

func (v *VM) runLine(line *ast.Line) error {
	if v.coordinator != nil {
		v.aquireCoordinatorPermission()
		// no metter what happens. If we run coordinated, report done to coordinator on return
		defer func() {
			v.coordinatorDone <- struct{}{}
		}()
	}

	// an empty line has no statements that would trigger actions like breakpoints
	// trigger these actions manually
	if len(line.Statements) == 0 {
		v.currentSourceLine = line.Start().Line
		v.currentSourceColoumn = 0
		v.sourceLineChanged()
	}

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

func (v *VM) runStmt(stmt ast.Statement) error {
	v.currentSourceColoumn = stmt.Start().Coloumn
	v.checkSourceLineChanged(stmt)
	switch e := stmt.(type) {
	case *ast.Assignment:
		return v.runAssignment(e)
	case *ast.IfStatement:
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
	case *ast.GoToStatement:
		line, err := v.runExpr(e.Line)
		if err != nil {
			return RuntimeError{err, stmt}
		}
		if !line.IsNumber() {
			return RuntimeError{fmt.Errorf("Can not goto a string (%s)", line.String()), e}
		}
		linenr := line.Number().IntPart()

		if linenr < 1 {
			linenr = 1
		}
		if linenr > 20 {
			linenr = 20
		}

		// goto one line before the actual target. After the jump, currentAstLine will be incremented and then match the target
		v.currentAstLine = int(linenr) - 1
		v.jumped = true
		return errAbortLine
	case *ast.Dereference:
		_, err := v.runDeref(e)
		return err
	default:
		return RuntimeError{fmt.Errorf("UNKNWON-STATEMENT:%T", e), stmt}
	}
}

func (v *VM) runAssignment(as *ast.Assignment) error {
	var newValue *Variable
	var err error
	if as.Operator != "=" {
		binop := ast.BinaryOperation{
			Exp1: &ast.Dereference{
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

func (v *VM) runExpr(expr ast.Expression) (*Variable, error) {
	switch e := expr.(type) {
	case *ast.StringConstant:
		return &Variable{Value: e.Value}, nil
	case *ast.NumberConstant:
		num, err := decimal.NewFromString(e.Value)
		if err != nil {
			return nil, err
		}
		return &Variable{Value: num}, nil
	case *ast.BinaryOperation:
		return v.runBinOp(e)
	case *ast.UnaryOperation:
		return v.runUnaryOp(e)
	case *ast.Dereference:
		return v.runDeref(e)
	default:
		return nil, RuntimeError{fmt.Errorf("UNKNWON-EXPRESSION:%T", e), expr}
	}
}

func (v *VM) runDeref(d *ast.Dereference) (*Variable, error) {
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

func (v *VM) runBinOp(op *ast.BinaryOperation) (*Variable, error) {
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

func (v *VM) runUnaryOp(op *ast.UnaryOperation) (*Variable, error) {
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
