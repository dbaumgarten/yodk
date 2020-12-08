package debug

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
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
	// ValidBreakpoints contains a set of valid-breakpoint-locations for each script.
	// If no value is returned for a script, all breakpoints are valid.
	// This is needed because in a nolol-script not every line is valid for a breakpoint
	ValidBreakpoints map[int]map[int]bool
	// CompiledCode contains the generated yolol-code for for VMs that are running NOLOL
	CompiledCode map[int]string
	// If set to true, runtime-errors should not interrupt script execution
	IgnoreErrs bool
}

// JoinPath wraps filepath.Join, but returns only the second part if the second part is an absolute path
func JoinPath(base string, other string) string {
	if filepath.IsAbs(other) || base == "" {
		return other
	}
	return filepath.Join(base, other)
}

// ScriptIndexByPath returns the index of the script with the given path
func (h Helper) ScriptIndexByPath(path string) int {
	for i, s := range h.ScriptNames {
		if JoinPath(h.Worspace, s) == path {
			return i
		}
	}
	return -1
}

// ScriptIndexByName returns the index of the script with the given name
func (h Helper) ScriptIndexByName(name string) int {
	for i, s := range h.ScriptNames {
		if s == name {
			return i
		}
	}
	return -1
}

// ReverseVarnameTranslation returns the compiled name of a variable, given the original (source) name
func (h Helper) ReverseVarnameTranslation(vmidx int, search string) string {
	if h.VariableTranslations[vmidx] == nil {
		return search
	}
	for k, v := range h.VariableTranslations[vmidx] {
		if v == search {
			return k
		}
	}
	return ""
}

// CurrentVM returns the currently selected VM (only used in cli-debugger)
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
		Worspace:             normalizePath(workspace),
		FinishedVMs:          make(map[int]bool),
		ValidBreakpoints:     make(map[int]map[int]bool),
		CompiledCode:         make(map[int]string),
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
			h.ValidBreakpoints[i] = findValidBreakpoints(yololcode)
			h.VariableTranslations[i] = converter.GetVariableTranslations()
			pri := parser.Printer{
				Mode: parser.PrintermodeReadable,
			}
			yololcodestr, _ := pri.Print(yololcode)
			h.CompiledCode[i] = yololcodestr
			thisVM = vm.Create(yololcode)
		} else {
			return nil, fmt.Errorf("Invalid file extension on: %s", inputFileName)
		}

		h.Vms[i] = thisVM
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
		ValidBreakpoints:     make(map[int]map[int]bool),
		CompiledCode:         make(map[int]string),
		IgnoreErrs:           t.IgnoreErrs,
	}

	for i, script := range t.Scripts {
		h.ScriptNames[i] = script
		h.Scripts[i], err = t.GetScriptCode(i)
		if err != nil {
			return nil, err
		}
	}

	if casenr < 1 || casenr > len(t.Cases) {
		return nil, fmt.Errorf("The test-file does not contain a case number %d!", casenr)
	}

	runner, err := t.GetRunner(casenr - 1)
	if err != nil {
		return nil, err
	}

	h.Vms = runner.VMs
	h.Coordinator = runner.Coordinator
	h.VariableTranslations = runner.VarTranslations

	for i, iv := range h.Vms {
		prepareVM(iv, h.ScriptNames[i])
		if strings.HasSuffix(h.ScriptNames[i], ".nolol") {
			h.ValidBreakpoints[i] = findValidBreakpoints(iv.GetProgram())
			pri := parser.Printer{
				Mode: parser.PrintermodeReadable,
			}
			yololcodestr, _ := pri.Print(iv.GetProgram())
			h.CompiledCode[i] = yololcodestr
		}
	}
	return h, nil
}

// returns a set of valid breakpoint locations for the given program
// a breakpoint is valid if there is a statement or an empty line at the given location
func findValidBreakpoints(prog *ast.Program) map[int]bool {
	valid := make(map[int]bool)
	prog.Accept(ast.VisitorFunc(func(node ast.Node, visitType int) error {
		switch n := node.(type) {
		case *ast.Assignment:
			valid[n.Start().Line] = true
			break
		case *ast.GoToStatement:
			valid[n.Start().Line] = true
			break
		case *ast.Dereference:
			valid[n.Start().Line] = true
			break
		case *ast.IfStatement:
			valid[n.Start().Line] = true
			break
		case *ast.Line:
			// if the line is non-empty, the statements on the line will validate the breakpoint
			if len(n.Statements) == 0 {
				valid[n.Start().Line] = true
			}
			break
		}
		return nil
	}))
	return valid
}
