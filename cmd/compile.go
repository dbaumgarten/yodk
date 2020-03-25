package cmd

import (
	"fmt"
	"io/ioutil"
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
	converter.Debug(debugLog)
	converted, err := converter.ConvertFile(fpath)
	exitOnError(err, "converting to yolol")
	gen := parser.Printer{}
	generated, err := gen.Print(converted)
	exitOnError(err, "generating code")
	err = ioutil.WriteFile(outfile, []byte(generated), 0700)
	exitOnError(err, "writing file")
}

func init() {
	rootCmd.AddCommand(compileCmd)
	compileCmd.Flags().StringVarP(&outputFile, "out", "o", "<inputfile>.out", "The output file")
	compileCmd.Flags().BoolVarP(&debugLog, "debug", "d", false, "Print debug logs while parsing")
}
