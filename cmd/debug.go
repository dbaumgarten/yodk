package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/dbaumgarten/yodk/vm"
	"github.com/spf13/cobra"
)

var yvm *vm.YololVM
var debugShell *ishell.Shell
var inputProg string

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug [file]",
	Short: "Debug a yolol program",
	Long:  `Execute program interactively in debugger`,
	Run: func(cmd *cobra.Command, args []string) {
		inputProg = loadInputFile(args[0])

		debugShell.Println("Loaded and paused programm. Enter 'x' to execute")
		debugShell.Run()
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(debugCmd)

	yvm = vm.NewYololVM()
	yvm.Pause()

	yvm.SetBreakpointHandler(func(x *vm.YololVM) bool {
		debugShell.Println("--Hit Breakpoint at line: ", x.CurrentLine())
		return false
	})

	yvm.SetErrorHandler(func(x *vm.YololVM, err error) bool {
		debugShell.Println("--A runtime error occured--")
		debugShell.Println(err)
		debugShell.Println("Execution paused")
		x.Pause()
		return true
	})

	yvm.SetFinishHandler(func(x *vm.YololVM) {
		debugShell.Println("--Program finished--")
		debugShell.Println("Enter x to restart")
	})

	debugShell = ishell.New()
	debugShell.AddCmd(&ishell.Cmd{
		Name: "x",
		Help: "execute program from start",
		Func: func(c *ishell.Context) {
			debugShell.Println("--Started--")
			go yvm.Run(inputProg)
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name: "p",
		Help: "pause execution",
		Func: func(c *ishell.Context) {
			yvm.Pause()
			debugShell.Println("--Paused--")
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name: "r",
		Help: "resume paused program",
		Func: func(c *ishell.Context) {
			go yvm.Resume()
			debugShell.Println("--Resumed--")
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name: "bp",
		Help: "add breakpoint at line",
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
			yvm.AddBreakpoint(line)
			debugShell.Println("--Breakpoint added--")
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name: "rbp",
		Help: "remove breakpoint at line",
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
			yvm.RemoveBreakpoint(line)
			debugShell.Println("--Breakpoint removed--")
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name: "lbp",
		Help: "list breakpoints",
		Func: func(c *ishell.Context) {
			debugShell.Println("Breakpoints at lines:", yvm.ListBreakpoints())
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name: "v",
		Help: "print all current variables",
		Func: func(c *ishell.Context) {
			debugShell.Println("--Variables--")
			debugShell.Println(yvm.PrintVariables())
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name: "?",
		Help: "show vm-state",
		Func: func(c *ishell.Context) {
			debugShell.Printf("--State: %d\n", yvm.State())
		},
	})
	debugShell.AddCmd(&ishell.Cmd{
		Name: "s",
		Help: "show programm",
		Func: func(c *ishell.Context) {
			current := yvm.CurrentLine()
			bps := yvm.ListBreakpoints()
			progLines := strings.Split(inputProg, "\n")
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
}

func contains(arr []int, val int) bool {
	for _, e := range arr {
		if e == val {
			return true
		}
	}
	return false
}
