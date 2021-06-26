package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/parser"

	"github.com/spf13/cobra"
)

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile [file]+",
	Short: "Compile nolol programms to yolol",
	Run: func(cmd *cobra.Command, args []string) {
		for _, file := range args {
			fmt.Println("Compiling file:", file)
			compileFile(file)
		}
	},
	Args: cobra.MinimumNArgs(1),
}

func compileFile(fpath string) {
	outfile := strings.Replace(fpath, path.Ext(fpath), ".yolol", -1)
	converter := nolol.NewConverter()
	converter.SetDebug(debugLog)
	converter.SetChipType(chipType)

	converted, compileerr := converter.LoadFile(fpath).Convert()

	// compilation failed completely. Fail now!
	if converted == nil {
		exitOnError(compileerr, "converting '"+fpath+"' to yolol")
	}

	gen := parser.Printer{}
	generated, err := gen.Print(converted)
	exitOnError(err, "generating code")
	err = ioutil.WriteFile(outfile, []byte(generated), 0700)
	exitOnError(err, "writing file")

	if compileerr != nil {
		fmt.Println("Compilation succeeded with errors. Please check the output:", compileerr)
		os.Exit(1)
	}

}

func init() {
	rootCmd.AddCommand(compileCmd)
	compileCmd.Flags().StringVarP(&outputFile, "out", "o", "<inputfile>.out", "The output file")
	compileCmd.Flags().BoolVarP(&debugLog, "debug", "d", false, "Print debug logs while parsing")
	compileCmd.Flags().StringVarP(&chipType, "chip", "c", "auto", "Chip-type to validate for. (auto|professional|advanced|basic)")
}
