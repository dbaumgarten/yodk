package cmd

import (
	"io/ioutil"
	"path"

	"github.com/dbaumgarten/yodk/nolol"
	"github.com/dbaumgarten/yodk/optimizers"
	"github.com/dbaumgarten/yodk/parser"

	"github.com/spf13/cobra"
)

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile [file]",
	Short: "Compile a nolol programm to yolol",
	Run: func(cmd *cobra.Command, args []string) {
		outfile := path.Base(args[0]) + ".out"
		p := nolol.NewNololParser()
		file := loadInputFile(args[0])
		parsed, errs := p.Parse(file)
		if errs != nil {
			exitOnError(errs, "parsing file")
		}
		converter := nolol.NewNololConverter()
		converted, err := converter.Convert(parsed)
		exitOnError(err, "compiling")
		opt := optimizers.NewCompoundOptimizer()
		err = opt.Optimize(converted)
		exitOnError(err, "performing optimisation")
		gen := parser.YololGenerator{}
		generated := gen.Generate(converted)
		ioutil.WriteFile(outfile, []byte(generated), 0700)
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(compileCmd)
	compileCmd.Flags().StringVarP(&outputFile, "out", "o", "<inputfile>.out", "The output file")
}
