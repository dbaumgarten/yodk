package debug

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/testing"
	"github.com/dbaumgarten/yodk/pkg/vm"
)

// JoinPath wraps filepath.Join, but returns only the second part if the second part is an absolute path
func JoinPath(base string, other string) string {
	if filepath.IsAbs(other) || base == "" {
		return other
	}
	return filepath.Join(base, other)
}

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
	Vms []*vm.VM
	// list of variable translations for the VMs
	// used to undo variable shortening performed by nolol using compilation
	VariableTranslations []map[string]string
	// number of the case in the given test to execute
	CaseNumber int
	// a folder all script paths are relative to
	Worspace string
	// a set of vms that have finished execution. This list is NOT managed by the Helper-class
	FinishedVMs map[int]bool
}

func (h Helper) ScriptIndexByPath(path string) int {
	for i, s := range h.ScriptNames {
		if JoinPath(h.Worspace, s) == path {
			return i
		}
	}
	return -1
}

func (h Helper) ScriptIndexByName(name string) int {
	for i, s := range h.ScriptNames {
		if s == name {
			return i
		}
	}
	return -1
}

func (h Helper) CurrentVM() *vm.VM {
	return h.Vms[h.CurrentScript]
}

// VMPrepareFunc receives a VM and prepares it for debugging
// (set error handlers etc.)
type VMPrepareFunc func(yvm *vm.VM, filename string)

// FromScripts receives a list of yolol/nolol filenames and creates a Helper from them
func FromScripts(workspace string, scripts []string, prepareVM VMPrepareFunc) (*Helper, error) {
	h := &Helper{
		ScriptNames:          scripts,
		Scripts:              make([]string, len(scripts)),
		VariableTranslations: make([]map[string]string, len(scripts)),
		Vms:                  make([]*vm.VM, len(scripts)),
		CurrentScript:        0,
		Coordinator:          vm.NewCoordinator(),
		Worspace:             workspace,
		FinishedVMs:          make(map[int]bool),
	}

	for i, inputFileName := range h.ScriptNames {
		filecontent, err := ioutil.ReadFile(JoinPath(workspace, inputFileName))
		if err != nil {
			return nil, err
		}

		h.Scripts[i] = string(filecontent)

		var thisVM *vm.VM

		if strings.HasSuffix(inputFileName, ".yolol") {
			thisVM, err = vm.CreateFromSource(h.Scripts[i])
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(inputFileName, ".nolol") {
			converter := nolol.NewConverter()
			yololcode, err := converter.ConvertFile(inputFileName)
			if err != nil {
				return nil, err
			}
			h.VariableTranslations[i] = converter.GetVariableTranslations()
			thisVM = vm.Create(yololcode)
		} else {
			return nil, fmt.Errorf("Invalid file extension on: %s", inputFileName)
		}

		h.Vms[i] = thisVM
		thisVM.SetIterations(0)
		thisVM.SetCoordinator(h.Coordinator)
		prepareVM(thisVM, inputFileName)
		thisVM.Resume()
	}
	return h, nil
}

// FromTest creates a Helper from the given test-file
func FromTest(workspace string, testfile string, casenr int, prepareVM VMPrepareFunc) (*Helper, error) {
	testfile = JoinPath(workspace, testfile)
	testfilecontent, err := ioutil.ReadFile(testfile)
	if err != nil {
		return nil, err
	}

	t, err := testing.Parse(testfilecontent, testfile)
	if err != nil {
		return nil, err
	}

	h := &Helper{
		ScriptNames:          make([]string, len(t.Scripts)),
		Scripts:              make([]string, len(t.Scripts)),
		VariableTranslations: make([]map[string]string, len(t.Scripts)),
		Vms:                  make([]*vm.VM, len(t.Scripts)),
		CurrentScript:        0,
		Coordinator:          vm.NewCoordinator(),
		Worspace:             filepath.Dir(testfile),
		FinishedVMs:          make(map[int]bool),
	}

	for i, script := range t.Scripts {
		h.ScriptNames[i] = script.Name
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
