package testing

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"sync"

	"github.com/dbaumgarten/yodk/nolol"

	"github.com/shopspring/decimal"
	yaml "gopkg.in/yaml.v2"

	"github.com/dbaumgarten/yodk/vm"
)

// Test defines a test-run
type Test struct {
	AbsolutePath string
	Scripts      []Script
	Cases        []Case
}

// Script contains run-options for a script in the test
type Script struct {
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
	Name    string
	Inputs  map[string]interface{}
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
	}
	return test, err
}

// RunTest runs a the given test
// if caseCallback is not nill it is called once before execution of every case
// if a scripts content field is empty, it is loaded from disk
func RunTest(t Test, caseCallback func(c Case)) []error {

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

		for key, value := range c.Inputs {
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

		vms := make([]*vm.YololVM, 0)
		for _, script := range t.Scripts {
			v := vm.NewYololVMCoordinated(coord)
			v.SetErrorHandler(errHandler)
			vms = append(vms, v)
			file := path.Join(path.Dir(t.AbsolutePath), script.Name)
			var scriptContent string
			if script.Content == "" {
				f, err := ioutil.ReadFile(file)
				if err != nil {
					fails = append(fails, err)
					return fails
				}
				scriptContent = string(f)
			} else {
				scriptContent = script.Content
			}

			v.SetIterations(script.Iterations)
			v.SetMaxExecutedLines(script.MaxLines)

			if strings.HasSuffix(script.Name, ".nolol") {
				conv := nolol.NewConverter()
				prog, err := conv.ConvertFromSource(string(scriptContent))
				if err != nil {
					fails = append(fails, err)
					return fails
				}
				v.Run(prog)
			} else {
				v.RunSource(string(scriptContent))
			}
		}

		coord.Run()
		coord.WaitForTermination()

		for key, value := range c.Outputs {
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
	}

	return fails
}
