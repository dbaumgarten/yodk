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

var spaces bool

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
	converter.UseSpaces = spaces
	converter.Debug(debugLog)
	converted, compileerr := converter.ConvertFile(fpath)

	// compilation failed completely. Fail now!
	if converted == nil {
		exitOnError(compileerr, "converting to yolol")
	}

	gen := parser.Printer{}
	gen.Mode = parser.PrintermodeSpaceless
	if spaces {
		gen.Mode = parser.PrintermodeCompact
	}
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
	compileCmd.Flags().BoolVar(&spaces, "spaces", false, "If true, output code with spaces where normal languages would expect them")
}
