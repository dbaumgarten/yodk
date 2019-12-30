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

// IsRunning returns true if the sub-vms are allowed to run
func (c *Coordinator) IsRunning() bool {
	c.condLock.Lock()
	defer c.condLock.Unlock()
	return c.currentVM >= 0
}

// Terminate all coordinated vms
func (c *Coordinator) Terminate() {
	for _, v := range c.vms {
		v.Terminate()
	}
}

// WaitForTermination blocks until all coordinated vms terminate
func (c *Coordinator) WaitForTermination() {
	for _, v := range c.vms {
		v.WaitForTermination()
	}
}

// GetVariable gets the current state of a global variable
func (c *Coordinator) GetVariable(name string) (*Variable, bool) {
	c.varLock.Lock()
	defer c.varLock.Unlock()
	val, exists := c.globalVariables[name]
	return val, exists
}

// GetVariables gets the current state of all global variables
func (c *Coordinator) GetVariables() map[string]Variable {
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
func (c *Coordinator) SetVariable(name string, value *Variable) error {
	c.varLock.Lock()
	defer c.varLock.Unlock()
	c.globalVariables[name] = value
	return nil
}

// registerVM registers a VM with the coordinator
// is called in the run() method of the VM
// if the vm is already registered nothing happens
func (c *Coordinator) registerVM(vm *YololVM) {
	c.condLock.Lock()
	defer c.condLock.Unlock()
	idx := -1
	for i, val := range c.vms {
		if val == vm {
			idx = i
		}
	}
	if idx == -1 {
		c.vms = append(c.vms, vm)
	}
}

// unRegisterVM stops a vm from participating in the coordinated execution
// a vm unregisters itself after termination to not block still running vms
func (c *Coordinator) unRegisterVM(vm *YololVM) {
	c.condLock.Lock()
	defer c.condLock.Unlock()
	idx := -1
	for i, val := range c.vms {
		if val == vm {
			idx = i
			break
		}
	}
	if idx != -1 {
		c.vms = append(c.vms[:idx], c.vms[idx+1:]...)
		if c.currentVM == idx && len(c.vms) > 0 {
			c.currentVM = c.currentVM % len(c.vms)
			c.condition.Broadcast()
		}
	}
}

// waitForTurn blocks until it is vms turn to execute a line
// vm must have registered itself before calling this
func (c *Coordinator) waitForTurn(vm *YololVM) {
	c.condLock.Lock()
	for c.currentVM == -1 || c.vms[c.currentVM] != vm {
		c.condition.Wait()
	}
	c.condLock.Unlock()
}

// finishTurn is called by the current active vm after completing its line
// allows the next vm to run
func (c *Coordinator) finishTurn() {
	c.condLock.Lock()
	defer c.condLock.Unlock()
	c.currentVM = (c.currentVM + 1) % len(c.vms)
	c.condition.Broadcast()
}
