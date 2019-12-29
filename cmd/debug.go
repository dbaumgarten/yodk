package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/dbaumgarten/yodk/nolol"
	"github.com/dbaumgarten/yodk/parser"

	"github.com/abiosoft/ishell"
	"github.com/dbaumgarten/yodk/vm"
	"github.com/spf13/cobra"
)

// index of the current script (the script targeted by commands)
// used to index into vms, inputProgs and inputFileNames
var currentScript int

// the coordinater to coordinate the running vms
var coordinator *vm.Coordinator
var debugShell *ishell.Shell

// source code of the running scripts
var inputProgs []string

// names of the running scripts
var inputFileNames []string

// list of vms for the running scripts
var vms []*vm.YololVM

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug [file] [file] [file]",
	Short: "Debug yolol/nolol programs",
	Long:  `Execute programs interactively in debugger`,
	Run: func(cmd *cobra.Command, args []string) {
		inputFileNames = args
		inputProgs = make([]string, len(inputFileNames))
		vms = make([]*vm.YololVM, len(inputFileNames))
		for i, filename := range inputFileNames {
			if !strings.HasSuffix(filename, ".yolol") && !strings.HasSuffix(filename, ".nolol") {
				fmt.Println("Unknown file-extension for file: ", filename)
				os.Exit(1)
			}
			inputProgs[i] = loadInputFile(args[i])
		}

		load()
		debugShell.Run()
	},
	Args: cobra.MinimumNArgs(1),
}

// create a VM for every script
func load() {
	coordinator = vm.NewCoordinator()

	for i := 0; i < len(inputFileNames); i++ {
		inputFileName := inputFileNames[i]
		inputProg := inputProgs[i]
		thisVM := vm.NewYololVMCoordinated(coordinator)
		vms[i] = thisVM
		thisVM.EnableLoop(true)
		if i == 0 {
			thisVM.Pause()
			currentScript = 0
		}
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
		if strings.HasSuffix(inputFileName, ".yolol") {

			thisVM.RunSource(inputProg)
		} else if strings.HasSuffix(inputFileName, ".nolol") {
			converter := nolol.NewConverter()
			yololcode, err := converter.ConvertFromSource(inputProg)
			if err != nil {
				exitOnError(err, "parsing nolol code")
			}
			thisVM.Run(yololcode)
		}
		debugShell.Printf("--Loaded %s--\n", inputFileName)
	}
	coordinator.Run()
	debugShell.Println("Loaded and paused programs. Enter 'c' to resume execution.")
}

// initialize the shell
func init() {
	rootCmd.AddCommand(debugCmd)

	debugShell = ishell.New()

	debugShell.AddCmd(&ishell.Cmd{
		Name:    "reset",
		Aliases: []string{"r"},
		Help:    "reset debugger",
		Func: func(c *ishell.Context) {
			load()
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name:    "scripts",
		Aliases: []string{"p"},
		Help:    "list scripts",
		Func: func(c *ishell.Context) {
			for i, file := range inputFileNames {
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
		Aliases: []string{"p"},
		Help:    "change currently viewed script",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 1 {
				debugShell.Println("You must enter a script name (run scripts to list them).")
				return
			}
			for i, file := range inputFileNames {
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
			progLines := strings.Split(inputProgs[currentScript], "\n")
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
			if !strings.HasSuffix(inputFileNames[currentScript], ".nolol") {
				debugShell.Print("Disas is only available when debugging nolol code")
			}
			current := vms[currentScript].CurrentAstLine()
			conv := nolol.NewConverter()
			ast, _ := conv.ConvertFromSource(inputProgs[currentScript])
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
