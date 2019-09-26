package cmd

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/dbaumgarten/yodk/nolol"
	"github.com/dbaumgarten/yodk/parser"

	"github.com/spf13/cobra"
)

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile [file]",
	Short: "Compile a nolol programm to yolol",
	Run: func(cmd *cobra.Command, args []string) {
		outfile := strings.Replace(args[0], path.Ext(args[0]), ".yolol", -1)
		file := loadInputFile(args[0])
		converter := nolol.NewNololConverter()
		converted, err := converter.ConvertFromSource(file)
		exitOnError(err, "converting to yolol")
		gen := parser.YololGenerator{}
		generated, err := gen.Generate(converted)
		exitOnError(err, "generating code")
		ioutil.WriteFile(outfile, []byte(generated), 0700)
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(compileCmd)
	compileCmd.Flags().StringVarP(&outputFile, "out", "o", "<inputfile>.out", "The output file")
}
