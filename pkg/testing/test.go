package testing

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	yaml "gopkg.in/yaml.v2"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/vm"
)

// Test defines a test-run
type Test struct {
	// The path where the test-file was located. Used to retrieve the script files.
	Path string
	// Scripts to use in this test
	Scripts []Script
	// Cases for this test
	Cases []Case
	// Maximum number of lines to run from the script (0=infinite)
	MaxLines int
	// Stop when is a map from global variable-name to value
	// Execution is stopped when at least one of the listed variables is equal to the value
	StopWhen map[string]interface{}
}

// Script contains run-options for a script in the test
type Script struct {
	// The absolut path where the test-file was located. Used to retrieve the script files.
	TestPath string
	// Name of the script to run
	Name string
	// Content is the content of the script. If empty, it is loaded from disk at run-time
	Content string
}

// Case defines inputs and expected outputs for a run
type Case struct {
	// Name of the testcase
	Name string
	// Values of gloal variables before run
	Inputs map[string]interface{}
	// Expected values of global vars after run
	Outputs map[string]interface{}
	// The same as Script.StopWhen. Both are merged together so this can be used to override/extend the script stop-conditions
	StopWhen map[string]interface{}
}

// CaseRunner represents a prepared test-case that is ready to run
type CaseRunner struct {
	Coordinator     *vm.Coordinator
	VMs             []*vm.VM
	VarTranslations []map[string]string
	Case            *Case
	StopConditions  map[string]*vm.Variable
}

func prefixVarname(inp string) string {
	if !strings.HasPrefix(inp, ":") {
		return ":" + inp
	}
	return inp
}

// Parse parses a yaml file into a Test
// path is the path from where the test was loaded. This is needed as the scripts are located relatice to the test-file
func Parse(file []byte, path string) (Test, error) {
	var test Test
	err := yaml.Unmarshal(file, &test)
	if err != nil {
		return test, fmt.Errorf("The provided test-file is invalid: %s", err.Error())
	}
	test.Path = path
	if test.MaxLines == 0 {
		test.MaxLines = 2000
	}
	for i := range test.Scripts {
		test.Scripts[i].TestPath = path
	}
	return test, nil
}

// GetCode returns the code for script either from the script struct itself or from the referenced file
func (script Script) GetCode() (string, error) {
	file := filepath.Join(filepath.Dir(script.TestPath), script.Name)
	if script.Content == "" {
		f, err := ioutil.ReadFile(file)
		if err != nil {
			return "", err
		}
		return string(f), nil
	}
	return script.Content, nil
}

// Run runs all test-cases
func (t Test) Run(callback func(Case)) []error {
	fails := make([]error, 0)
	for i := range t.Cases {
		if callback != nil {
			callback(t.Cases[i])
		}
		runner, err := t.GetRunner(i)
		if err != nil {
			fails = append(fails, err)
			continue
		}
		casefails := runner.Run()
		fails = append(fails, casefails...)
	}
	return fails
}

// GetRunner creates an executable TestRunner for the given testcase
func (t Test) GetRunner(casenr int) (runner *CaseRunner, err error) {
	c := t.Cases[casenr]

	runner = &CaseRunner{
		Coordinator:    vm.NewCoordinator(),
		Case:           &c,
		StopConditions: make(map[string]*vm.Variable, len(t.Scripts)),
		VMs:            make([]*vm.VM, len(t.Scripts)),
	}

	c.initializeVariables(runner.Coordinator)
	runner.VMs, runner.VarTranslations, err = t.createVMs(runner.Coordinator)

	runner.StopConditions = mergeStopConditions(&t, &c)

	varChangeHandler := func(vm *vm.VM, name string, value *vm.Variable) bool {
		if cond, is := runner.StopConditions[name]; is {
			if cond.Equals(value) {
				// stop condition reached. Terminate all VMs
				go runner.Coordinator.Terminate()
				return false
			}
		}
		return true
	}

	for _, vm := range runner.VMs {
		vm.SetVariableChangedHandler(varChangeHandler)
	}

	return runner, nil
}

// initializeVariables adds the variables required for the testcase
// to the variables of the given Coordinator
func (c Case) initializeVariables(coord *vm.Coordinator) error {
	for key, value := range c.Inputs {
		//key = strings.ToLower(key)
		variable, err := vm.VariableFromType(value)
		if err != nil {
			return err
		}
		coord.SetVariable(prefixVarname(key), variable)
	}
	return nil
}

// createVMs creates and sets up the required vms for this test
// coord is the coordinator to use with the VMs
// Run() has been called on the returned VMs, but they are paused until coord.Run() is called
// Also returns variable-name translation-tables for nolol scripts
func (t Test) createVMs(coord *vm.Coordinator) ([]*vm.VM, []map[string]string, error) {
	vms := make([]*vm.VM, len(t.Scripts))
	translationTables := make([]map[string]string, len(t.Scripts))
	for i, script := range t.Scripts {
		var v *vm.VM

		if strings.HasSuffix(script.Name, ".nolol") {
			conv := nolol.NewConverter()
			file := filepath.Join(filepath.Dir(script.TestPath), script.Name)
			prog, err := conv.ConvertFile(file)
			translationTables[i] = conv.GetVariableTranslations()
			if err != nil {
				return nil, nil, err
			}
			v = vm.Create(prog)
		} else {
			scriptContent, err := script.GetCode()
			if err != nil {
				return nil, nil, err
			}
			v, err = vm.CreateFromSource(string(scriptContent))
			if err != nil {
				return nil, nil, err
			}
		}

		v.SetMaxExecutedLines(t.MaxLines)
		v.SetCoordinator(coord)
		vms[i] = v
		v.Resume()
	}
	return vms, translationTables, nil
}

func mergeStopConditions(test *Test, c *Case) map[string]*vm.Variable {
	conds := make(map[string]*vm.Variable)
	// If there are no stop-conditions, set a default
	if len(test.StopWhen) == 0 {
		conds[":done"] = vm.VariableFromString("1")
	}
	for k, v := range test.StopWhen {
		k = prefixVarname(k)
		conds[k], _ = vm.VariableFromType(v)
	}
	for k, v := range c.StopWhen {
		k = prefixVarname(k)
		conds[k], _ = vm.VariableFromType(v)
	}
	return conds
}

// Run executes the case-runner
func (cr CaseRunner) Run() []error {

	fails := make([]error, 0)
	flock := &sync.Mutex{}

	errHandler := func(vm *vm.VM, err error) bool {
		flock.Lock()
		defer flock.Unlock()
		fails = append(fails, err)
		go cr.Coordinator.Terminate()
		return true
	}

	for _, vm := range cr.VMs {
		vm.SetErrorHandler(errHandler)
	}

	cr.Coordinator.Run()
	cr.Coordinator.WaitForTermination()

	caseFails := cr.Case.checkResults(cr.Coordinator)
	fails = append(fails, caseFails...)
	return fails
}

// checkResults compares the global variables of coord with the expected results for c
// and returns found errors
func (c Case) checkResults(coord *vm.Coordinator) []error {
	fails := make([]error, 0)
	for key, value := range c.Outputs {
		key = prefixVarname(key)
		var fail error
		expected, err := vm.VariableFromType(value)
		if err != nil {
			fail = fmt.Errorf("Invalid type for expected var: %T", value)
			fails = append(fails, fail)
			continue
		}
		actual, exists := coord.GetVariable(key)

		if !exists {
			fail = fmt.Errorf("Expected output variable %s does not exist", key)
		} else {
			if !actual.SameType(expected) {
				fail = fmt.Errorf("Case '%s': Output '%s' has type '%s' but should be '%s' ", c.Name, key, actual.TypeName(), expected.TypeName())

			} else if !actual.Equals(expected) {
				fail = fmt.Errorf("Case '%s': Output '%s' has value %s but should be %s ", c.Name, key, actual.Repr(), expected.Repr())
			}
		}
		if fail != nil {
			fails = append(fails, fail)
		}
	}
	return fails
}
