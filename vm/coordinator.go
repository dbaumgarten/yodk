package vm

import "sync"

// Coordinator is responsible for coordinating the execution of multiple VMs
// It coordinates the line-by-line execution of the scripts and provides shared global variables
type Coordinator struct {
	vms             []*YololVM
	currentVM       int
	globalVariables map[string]*Variable
	condLock        *sync.Mutex
	condition       *sync.Cond
	varLock         *sync.Mutex
}

// NewCoordinator returns a new coordinator
func NewCoordinator() *Coordinator {
	lock := &sync.Mutex{}
	return &Coordinator{
		vms:             make([]*YololVM, 0),
		currentVM:       -1,
		globalVariables: make(map[string]*Variable),
		condLock:        lock,
		condition:       sync.NewCond(lock),
		varLock:         &sync.Mutex{},
	}
}

// Run starts the coordinated exection
// Run() on the VMs must have been called before
func (c *Coordinator) Run() {
	c.condLock.Lock()
	defer c.condLock.Unlock()
	c.currentVM = 0
	c.condition.Broadcast()
}

// Stop stops the execution
func (c *Coordinator) Stop() {
	c.condLock.Lock()
	defer c.condLock.Unlock()
	c.currentVM = 0
	c.condition.Broadcast()
}

// getVariable gets the current state of a global variable
func (c *Coordinator) getVariable(name string) (*Variable, bool) {
	c.varLock.Lock()
	defer c.varLock.Unlock()
	val, exists := c.globalVariables[name]
	return val, exists
}

// GetVariables gets the current state of all global variables
func (c *Coordinator) getVariables() map[string]Variable {
	c.varLock.Lock()
	defer c.varLock.Unlock()
	varlist := make(map[string]Variable)
	for key, value := range c.globalVariables {
		varlist[key] = Variable{
			Value: value.Value,
		}
	}
	return varlist
}

// SetVariable sets the current state of a global variable
func (c *Coordinator) setVariable(name string, value *Variable) error {
	c.varLock.Lock()
	defer c.varLock.Unlock()
	c.globalVariables[name] = value
	return nil
}

func (c *Coordinator) registerVM(vm *YololVM) {
	c.vms = append(c.vms, vm)
}

func (c *Coordinator) waitForTurn(vm *YololVM) {
	c.condLock.Lock()
	for c.currentVM == -1 || c.vms[c.currentVM] != vm {
		c.condition.Wait()
	}
	c.condLock.Unlock()
}

func (c *Coordinator) finishTurn() {
	c.condLock.Lock()
	defer c.condLock.Unlock()
	c.currentVM = (c.currentVM + 1) % len(c.vms)
	c.condition.Broadcast()
}
