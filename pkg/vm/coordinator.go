package vm

import (
	"strings"
	"sync"
)

// Coordinator is responsible for coordinating the execution of multiple VMs
// It coordinates the line-by-line execution of the scripts and provides shared global variables
type Coordinator struct {
	vms              []*VM
	runLineChannels  []chan struct{}
	lineDoneChannels []chan struct{}
	globalVariables  map[string]*Variable
	varLock          *sync.Mutex
}

// NewCoordinator returns a new coordinator
func NewCoordinator() *Coordinator {
	return &Coordinator{
		vms:              make([]*VM, 0),
		runLineChannels:  make([]chan struct{}, 0),
		lineDoneChannels: make([]chan struct{}, 0),
		globalVariables:  make(map[string]*Variable),
		varLock:          &sync.Mutex{},
	}
}

// Run starts the coordinated exection
// Once run has been called, no new VMs MUSt be added!!!
func (c *Coordinator) Run() {
	go c.run()
}

// Terminate all coordinated vms
// Once all VMs terminate the coordinator-goroutine will also shut-down
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
// getting variables is case-insensitive
func (c *Coordinator) GetVariable(name string) (*Variable, bool) {
	name = strings.ToLower(name)
	c.varLock.Lock()
	defer c.varLock.Unlock()
	val, exists := c.globalVariables[name]
	return val, exists
}

// GetVariables gets the current state of all global variables
// All returned variables have normalized (lowercased) names
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
// setting variables is case-insensitive
func (c *Coordinator) SetVariable(name string, value *Variable) error {
	name = strings.ToLower(name)
	c.varLock.Lock()
	defer c.varLock.Unlock()
	c.globalVariables[name] = value
	return nil
}

// registerVM registers a VM with the coordinator
// is called by the vm in SetCoordinator.
// returns two channels. The first is used to signal to the VM that it may run a line
// the second one is used by the VM to signal it finished the line.
// To remove the vm from coordination close the second channel
func (c *Coordinator) registerVM(vm *VM) (<-chan struct{}, chan<- struct{}) {
	runChannel := make(chan struct{})
	doneChannel := make(chan struct{})
	c.vms = append(c.vms, vm)
	c.runLineChannels = append(c.runLineChannels, runChannel)
	c.lineDoneChannels = append(c.lineDoneChannels, doneChannel)
	return runChannel, doneChannel
}

// remove a VM from the coordination
// execurted once a VM closes it's done-channel
func (c *Coordinator) remove(idx int) {
	c.vms = append(c.vms[:idx], c.vms[idx+1:]...)
	close(c.runLineChannels[idx])
	c.runLineChannels = append(c.runLineChannels[:idx], c.runLineChannels[idx+1:]...)
	c.lineDoneChannels = append(c.lineDoneChannels[:idx], c.lineDoneChannels[idx+1:]...)
}

func (c *Coordinator) run() {
	for {
		for i := 0; i < len(c.runLineChannels); i++ {
			runch := c.runLineChannels[i]
			donech := c.lineDoneChannels[i]

			select {
			case runch <- struct{}{}:
				// the vm resceived the permission to run. Continue execution normally
			case <-donech:
				// the client closed the donechannel. This means he does not longer participate in coordination
				c.remove(i)
				continue
			}

			_, open := <-donech
			if !open {
				c.remove(i)
			}
		}
		if len(c.vms) == 0 {
			return
		}
	}
}
