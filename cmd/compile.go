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

var spaceless bool

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
	converter.SetSpaceless(spaceless)
	converter.SetDebug(debugLog)
	converted, compileerr := converter.LoadFile(fpath).Convert()

	// compilation failed completely. Fail now!
	if converted == nil {
		exitOnError(compileerr, "converting '"+fpath+"' to yolol")
	}

	gen := parser.Printer{}
	if spaceless {
		gen.Mode = parser.PrintermodeSpaceless
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
	compileCmd.Flags().BoolVar(&spaceless, "spaceless", false, "If true, output code with minimal spaces (might break script)")
}
