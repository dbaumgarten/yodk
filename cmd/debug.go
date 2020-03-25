package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/testing"

	"github.com/abiosoft/ishell"
	"github.com/dbaumgarten/yodk/pkg/vm"
	"github.com/spf13/cobra"
)

// cli args passed to this command
var cliArgs []string

// index of the current script (the script targeted by commands)
// used to index into vms, inputScripts and scriptFileNames
var currentScript int

// the coordinater to coordinate the running vms
var coordinator *vm.Coordinator
var debugShell *ishell.Shell

// source code of the running scripts
var inputScripts []string

// names of the running scripts
var scriptFileNames []string

// list of vms for the running scripts
var vms []*vm.YololVM

// number of the case in the given test to execute
var caseNumber int

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug [script]+ / debug [testfile]",
	Short: "Debug yolol/nolol programs or tests",
	Long:  `Execute programs interactively in debugger`,
	Run: func(cmd *cobra.Command, args []string) {
		cliArgs = args
		load()
		debugShell.Run()
	},
	Args: cobra.MinimumNArgs(1),
}

// load input scripts
// decide whether to load a bunch of scripts or a test-file
func load() {
	containsScript := false
	containsTest := false
	for _, arg := range cliArgs {
		if strings.HasSuffix(arg, ".yaml") {
			containsTest = true
		} else if strings.HasSuffix(arg, ".yolol") || strings.HasSuffix(arg, ".nolol") {
			containsScript = true
		} else {
			fmt.Println("Unknown file-extension for file: ", arg)
			os.Exit(1)
		}
	}

	if containsScript && containsTest {
		fmt.Println("Can not mix test-files and scripts.")
		os.Exit(1)
	}

	if len(cliArgs) > 1 && containsTest {
		fmt.Println("Can only debug one test at once")
		os.Exit(1)
	}

	if containsTest {
		loadTest()
	} else {
		loadScripts()
	}

	debugShell.Println("Loaded and paused programs. Enter 'c' to resume execution.")
}

// create a VM for every script
func loadScripts() {
	scriptFileNames = cliArgs
	inputScripts = make([]string, len(cliArgs))
	vms = make([]*vm.YololVM, len(cliArgs))
	currentScript = 0
	for i, filename := range cliArgs {
		inputScripts[i] = loadInputFile(filename)
	}

	coordinator = vm.NewCoordinator()

	for i := 0; i < len(scriptFileNames); i++ {
		inputFileName := scriptFileNames[i]
		inputProg := inputScripts[i]
		thisVM := vm.NewYololVMCoordinated(coordinator)
		thisVM.SetIterations(0)
		vms[i] = thisVM
		prepareVM(thisVM, inputFileName)

		if strings.HasSuffix(inputFileName, ".yolol") {
			thisVM.RunSource(inputProg)
		} else if strings.HasSuffix(inputFileName, ".nolol") {
			converter := nolol.NewConverter()
			dir := filepath.Dir(inputFileName)
			yololcode, err := converter.ConvertFromSource(inputProg, nolol.DiskFileSystem{Dir: dir})
			if err != nil {
				exitOnError(err, "parsing nolol code")
			}
			thisVM.Run(yololcode)
		}
		debugShell.Printf("--Loaded %s--\n", inputFileName)
	}
}

// create VMs from the given test-file
func loadTest() {
	var err error

	file := loadInputFile(cliArgs[0])
	abs, _ := filepath.Abs(cliArgs[0])
	t, err := testing.Parse([]byte(file), abs)
	exitOnError(err, "parsing test-file")

	scriptFileNames = make([]string, len(t.Scripts))
	for i, script := range t.Scripts {
		scriptFileNames[i] = filepath.Join(filepath.Dir(t.AbsolutePath), script.Name)
	}

	inputScripts = make([]string, len(t.Scripts))
	for i, script := range t.Scripts {
		inputScripts[i], _ = script.GetCode()
	}

	c := t.Cases[caseNumber-1]

	coordinator = vm.NewCoordinator()
	c.InitializeVariables(coordinator)

	v, err := t.CreateVMs(coordinator, nil)
	exitOnError(err, "creating VMs for test")
	vms = v
	for i, iv := range vms {
		prepareVM(iv, scriptFileNames[i])
	}
	currentScript = 0
}

// prepares the given VM for use in the debugger
func prepareVM(thisVM *vm.YololVM, inputFileName string) {
	thisVM.SetBreakpointHandler(func(x *vm.YololVM) bool {
		debugShell.Printf("--Hit Breakpoint at %s:%d--\n", inputFileName, x.CurrentSourceLine())
		return false
	})
	thisVM.SetErrorHandler(func(x *vm.YololVM, err error) bool {
		debugShell.Printf("--A runtime error occured at %s:%d--\n", inputFileName, x.CurrentSourceLine())
		debugShell.Println(err)
		debugShell.Println("--Execution paused--")
		return false
	})
	thisVM.SetFinishHandler(func(x *vm.YololVM) {
		debugShell.Printf("--Program %s finished--\n", inputFileName)
	})
}

// initialize the shell
func init() {
	debugCmd.Flags().IntVarP(&caseNumber, "case", "c", 1, "Numer of the case to execute when debugging a test")
	rootCmd.AddCommand(debugCmd)

	debugShell = ishell.New()

	debugShell.AddCmd(&ishell.Cmd{
		Name:    "reset",
		Aliases: []string{"r"},
		Help:    "reset debugger",
		Func: func(c *ishell.Context) {
			coordinator.Terminate()
			load()
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "scripts",
		Aliases: []string{"ll"},
		Help:    "list scripts",
		Func: func(c *ishell.Context) {
			for i, file := range scriptFileNames {
				line := "  "
				if i == currentScript {
					line = "> "
				}
				line += file
				debugShell.Println(line)
			}
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "choose",
		Aliases: []string{"cd"},
		Help:    "change currently viewed script",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 1 {
				debugShell.Println("You must enter a script name (run scripts to list them).")
				return
			}
			for i, file := range scriptFileNames {
				if file == c.Args[0] {
					currentScript = i
					debugShell.Printf("--Changed to %s--\n", file)
					return
				}
			}
			debugShell.Printf("--Unknown script %s--", c.Args[0])
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "pause",
		Aliases: []string{"p"},
		Help:    "pause execution",
		Func: func(c *ishell.Context) {
			vms[currentScript].Pause()
			debugShell.Println("--Paused--")
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "continue",
		Aliases: []string{"c"},
		Help:    "continue paused execution",
		Func: func(c *ishell.Context) {
			if !coordinator.IsRunning() {
				coordinator.Run()
				return
			}
			if vms[currentScript].State() != vm.StatePaused {
				debugShell.Println("The current script is not paused.")
				return
			}
			err := vms[currentScript].Resume()
			if err == nil {
				debugShell.Println("--Resumed--")
			} else {
				debugShell.Println(err)
			}
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "step",
		Aliases: []string{"s"},
		Help:    "execute the next line and pause again",
		Func: func(c *ishell.Context) {
			if vms[currentScript].Step() == nil {
				debugShell.Println("--Line executed. Paused again--")
			}
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "break",
		Aliases: []string{"b"},
		Help:    "add breakpoint at line",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 1 {
				debugShell.Println("You must enter a line number for the breakpoint.")
				return
			}
			line, err := strconv.Atoi(c.Args[0])
			if err != nil {
				debugShell.Println("Error parsing line-number: ", err)
				return
			}
			vms[currentScript].AddBreakpoint(line)
			debugShell.Println("--Breakpoint added--")
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "delete",
		Aliases: []string{"d"},
		Help:    "delete breakpoint at line",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 1 {
				debugShell.Println("You must enter a line number for the breakpoint.")
				return
			}
			line, err := strconv.Atoi(c.Args[0])
			if err != nil {
				debugShell.Println("Error parsing line-number: ", err)
				return
			}
			vms[currentScript].RemoveBreakpoint(line)
			debugShell.Println("--Breakpoint removed--")
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "vars",
		Aliases: []string{"v"},
		Help:    "print all current variables",
		Func: func(c *ishell.Context) {
			debugShell.Println("--Variables--")
			vars := sortVariables(vms[currentScript].GetVariables())
			for _, variable := range vars {
				if variable.val.IsString() {
					debugShell.Println(variable.name, "'"+variable.val.String()+"'")
				}
				if variable.val.IsNumber() {
					debugShell.Println(variable.name, variable.val.Itoa())
				}
			}
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "info",
		Aliases: []string{"i"},
		Help:    "show vm-state",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)
			statestr := ""
			switch vms[currentScript].State() {
			case vm.StateIdle:
				statestr = "READY"
			case vm.StateRunning:
				statestr = "RUNNING"
			case vm.StatePaused:
				statestr = "PAUSED"
			case vm.StateStep:
				statestr = "STEPPING"
			case vm.StateDone:
				statestr = "DONE"
			case vm.StateKill:
				statestr = "TERMINATING"
			}
			debugShell.Printf("--State: %s\n", statestr)
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "list",
		Aliases: []string{"l"},
		Help:    "show programm source code",
		Func: func(c *ishell.Context) {
			current := vms[currentScript].CurrentSourceLine()
			bps := vms[currentScript].ListBreakpoints()
			progLines := strings.Split(inputScripts[currentScript], "\n")
			debugShell.Println("--Programm--")
			pfx := ""
			for i, line := range progLines {
				if i+1 == current {
					pfx = ">"
				} else {
					pfx = " "
				}
				if contains(bps, i+1) {
					pfx += "x"
				} else {
					pfx += " "
				}
				pfx += fmt.Sprintf("%3d ", i+1)
				debugShell.Println(pfx + line)
			}
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "disas",
		Aliases: []string{"d"},
		Help:    "show yolol code for nolol source",
		Func: func(c *ishell.Context) {
			if !strings.HasSuffix(scriptFileNames[currentScript], ".nolol") {
				debugShell.Print("Disas is only available when debugging nolol code")
			}
			dir := filepath.Dir(scriptFileNames[currentScript])
			current := vms[currentScript].CurrentAstLine()
			conv := nolol.NewConverter()
			ast, err := conv.ConvertFromSource(inputScripts[currentScript], nolol.DiskFileSystem{Dir: dir})
			if err != nil {
				fmt.Println("Error when converting nolol: ", err.Error())
				return
			}
			yolol, _ := (&parser.Printer{}).Print(ast)
			progLines := strings.Split(yolol, "\n")
			debugShell.Println("--Programm--")
			pfx := ""
			for i, line := range progLines {
				if i+1 == current {
					pfx = ">"
				} else {
					pfx = " "
				}
				pfx += fmt.Sprintf("%3d ", i+1)
				debugShell.Println(pfx + line)
			}
		},
	})
}

type namedVariable struct {
	name string
	val  vm.Variable
}

func sortVariables(vars map[string]vm.Variable) []namedVariable {
	sorted := make([]namedVariable, 0, len(vars))
	for k, v := range vars {
		sorted = append(sorted, namedVariable{
			k,
			v,
		})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].name < sorted[j].name
	})
	return sorted
}

func contains(arr []int, val int) bool {
	for _, e := range arr {
		if e == val {
			return true
		}
	}
	return false
}
