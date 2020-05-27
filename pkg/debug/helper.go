package debug

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/testing"
	"github.com/dbaumgarten/yodk/pkg/vm"
)

// Helper bundles a lot of stuff you need to debuy yolol/nolol-code
type Helper struct {
	// index of the current script (the script targeted by commands)
	// used to index into vms, inputScripts and scriptFileNames
	CurrentScript int
	// the coordinater to coordinate the running vms
	Coordinator *vm.Coordinator
	// source code of the running Scripts
	Scripts []string
	// names of the running scripts
	ScriptNames []string
	// list of Vms for the running scripts
	Vms []*vm.YololVM
	// list of variable translations for the VMs
	// used to undo variable shortening performed by nolol using compilation
	VariableTranslations []map[string]string
	// number of the case in the given test to execute
	CaseNumber int
}

// VMPrepareFunc receives a VM and prepares it for debugging
// (set error handlers etc.)
type VMPrepareFunc func(yvm *vm.YololVM, filename string)

// FromScripts receives a list of yolol/nolol filenames and creates a Helper from them
func FromScripts(scripts []string, prepareVM VMPrepareFunc) (*Helper, error) {
	h := &Helper{
		ScriptNames:          scripts,
		Scripts:              make([]string, len(scripts)),
		VariableTranslations: make([]map[string]string, len(scripts)),
		Vms:                  make([]*vm.YololVM, len(scripts)),
		CurrentScript:        0,
		Coordinator:          vm.NewCoordinator(),
	}

	for i, inputFileName := range h.ScriptNames {
		filecontent, err := ioutil.ReadFile(inputFileName)
		if err != nil {
			return nil, err
		}

		h.Scripts[i] = string(filecontent)

		thisVM := vm.NewYololVMCoordinated(h.Coordinator)
		h.Vms[i] = thisVM
		thisVM.SetIterations(0)
		prepareVM(thisVM, inputFileName)

		if strings.HasSuffix(inputFileName, ".yolol") {
			thisVM.RunSource(h.Scripts[i])
		} else if strings.HasSuffix(inputFileName, ".nolol") {
			converter := nolol.NewConverter()
			yololcode, err := converter.ConvertFile(inputFileName)
			if err != nil {
				return nil, err
			}
			h.VariableTranslations[i] = converter.GetVariableTranslations()
			thisVM.Run(yololcode)
		}
	}
	return h, nil
}

// FromTest creates a Helper from the given test-file
func FromTest(testfile string, casenr int, prepareVM VMPrepareFunc) (*Helper, error) {

	testfilecontent, err := ioutil.ReadFile(testfile)
	if err != nil {
		return nil, err
	}

	// the source-files are relative to the test-file location. Therefore we need the absolute test-file location
	absoluteFilepath, _ := filepath.Abs(testfile)

	t, err := testing.Parse(testfilecontent, absoluteFilepath)
	if err != nil {
		return nil, err
	}

	h := &Helper{
		ScriptNames:          make([]string, len(t.Scripts)),
		Scripts:              make([]string, len(t.Scripts)),
		VariableTranslations: make([]map[string]string, len(t.Scripts)),
		Vms:                  make([]*vm.YololVM, len(t.Scripts)),
		CurrentScript:        0,
		Coordinator:          vm.NewCoordinator(),
	}

	for i, script := range t.Scripts {
		h.ScriptNames[i] = filepath.Join(filepath.Dir(t.AbsolutePath), script.Name)
		h.Scripts[i], err = script.GetCode()
		if err != nil {
			return nil, err
		}
	}

	c := t.Cases[casenr-1]

	h.Coordinator = vm.NewCoordinator()
	c.InitializeVariables(h.Coordinator)

	h.Vms, h.VariableTranslations, err = t.CreateVMs(h.Coordinator, nil)
	if err != nil {
		return nil, err
	}

	for i, iv := range h.Vms {
		prepareVM(iv, h.ScriptNames[i])
	}
	return h, nil
}
