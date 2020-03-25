package testing

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dbaumgarten/yodk/pkg/nolol"

	"github.com/shopspring/decimal"
	yaml "gopkg.in/yaml.v2"

	"github.com/dbaumgarten/yodk/pkg/vm"
)

// Test defines a test-run
type Test struct {
	// The absolut path where the test-file was located. Used to retrieve the script files.
	AbsolutePath string
	// Scripts to use in this test
	Scripts []Script
	// Cases for this test
	Cases []Case
}

// Script contains run-options for a script in the test
type Script struct {
	// The absolut path where the test-file was located. Used to retrieve the script files.
	AbsolutePath string
	// Name of the script to run
	Name string
	// Maximum number of iterations for the script (0=infinite)
	Iterations int
	// Maximum number of lines to run from the script (0=infinite)
	MaxLines int
	// the content of the script. If empty, it is loaded from disk at run-time
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
}

func prefixVarname(inp string) string {
	if !strings.HasPrefix(inp, ":") {
		return ":" + inp
	}
	return inp
}

// Parse parses a yaml file into a Test
// absolutePath is the path from where the test was loaded
// scripts are loaded relative to this path
func Parse(file []byte, absolutePath string) (Test, error) {
	var test Test
	err := yaml.Unmarshal(file, &test)
	test.AbsolutePath = absolutePath
	for i, script := range test.Scripts {
		if script.Iterations == 0 {
			test.Scripts[i].Iterations = 1
		}
		test.Scripts[i].AbsolutePath = absolutePath
	}
	return test, err
}

// InitializeVariables adds the variables required for the testcase
// to the variables of the given Coordinator
func (c Case) InitializeVariables(coord *vm.Coordinator) {
	for key, value := range c.Inputs {
		//key = strings.ToLower(key)
		variable := &vm.Variable{}
		if number, isnum := value.(int); isnum {
			variable.Value = decimal.NewFromFloat(float64(number))
		} else if number, isnum := value.(float64); isnum {
			variable.Value = decimal.NewFromFloat(number)
		} else {
			variable = &vm.Variable{
				Value: value,
			}
		}
		coord.SetVariable(prefixVarname(key), variable)
	}
}

// GetCode returns the code for script either from the script struct itself or from the referenced file
func (script Script) GetCode() (string, error) {
	file := filepath.Join(filepath.Dir(script.AbsolutePath), script.Name)
	if script.Content == "" {
		f, err := ioutil.ReadFile(file)
		if err != nil {
			return "", err
		}
		return string(f), nil
	}
	return script.Content, nil
}

// CreateVMs creates and sets up the required vms for this test
// coord is the coordinato to use with the VMs
// Run() has been called on the returned VMs, but they are paused until coord.Run() is called
// The error handler of the VMs is set to errF
func (t Test) CreateVMs(coord *vm.Coordinator, errF vm.ErrorHandlerFunc) ([]*vm.YololVM, error) {
	vms := make([]*vm.YololVM, 0)
	for _, script := range t.Scripts {
		v := vm.NewYololVMCoordinated(coord)
		v.SetIterations(script.Iterations)
		v.SetMaxExecutedLines(script.MaxLines)
		v.SetErrorHandler(errF)
		vms = append(vms, v)

		scriptContent, err := script.GetCode()
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(script.Name, ".nolol") {
			conv := nolol.NewConverter()
			dir := filepath.Dir(filepath.Join(filepath.Dir(script.AbsolutePath), script.Name))
			prog, err := conv.ConvertFromSource(string(scriptContent), nolol.DiskFileSystem{Dir: dir})
			if err != nil {
				return nil, err
			}
			v.Run(prog)
		} else {
			v.RunSource(string(scriptContent))
		}
	}
	return vms, nil
}

// CheckResults compares the global variables of coord with the expected results for c
// and returns found errors
func (c Case) CheckResults(coord *vm.Coordinator) []error {
	fails := make([]error, 0)
	for key, value := range c.Outputs {
		//key = strings.ToLower(key)
		key = prefixVarname(key)
		expected := &vm.Variable{
			Value: value,
		}
		actual, exists := coord.GetVariable(key)
		var fail error
		if !exists {
			fail = fmt.Errorf("Expected output variable %s does not exist", key)
		} else {
			if actual.Itoa() != expected.Itoa() {
				fail = fmt.Errorf("Case '%s': Output '%s' has value '%s' but should be '%s' ", c.Name, key, actual.Itoa(), expected.Itoa())
			}
		}
		if fail != nil {
			fails = append(fails, fail)
		}
	}
	return fails
}

// Run runs a the given test and return found errors
// caseCallback is called before executing a case. Can be used for logging.
// Main method of the test class
func (t Test) Run(caseCallback func(c Case)) []error {

	fails := make([]error, 0)
	flock := &sync.Mutex{}

	errHandler := func(vm *vm.YololVM, err error) bool {
		flock.Lock()
		defer flock.Unlock()
		fails = append(fails, err)
		return false
	}

	for _, c := range t.Cases {
		if caseCallback != nil {
			caseCallback(c)
		}
		coord := vm.NewCoordinator()

		c.InitializeVariables(coord)

		_, err := t.CreateVMs(coord, errHandler)
		if err != nil {
			return []error{err}
		}

		coord.Run()
		coord.WaitForTermination()

		caseFails := c.CheckResults(coord)
		fails = append(fails, caseFails...)
	}

	return fails
}
