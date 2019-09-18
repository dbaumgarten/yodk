package cmd

import (
	"io/ioutil"
	"path"

	"github.com/dbaumgarten/yodk/generators"
	"github.com/dbaumgarten/yodk/optimizers"
	"github.com/dbaumgarten/yodk/parser"
	"github.com/spf13/cobra"
)

var outputFile string

// optimizeCmd represents the compile command
var optimizeCmd = &cobra.Command{
	Use:   "optimize [file]",
	Short: "Optimize a yolo programm",
	Long:  `Parse the input file, run optimizers and re-generate yolol-code from the ast.`,
	Run: func(cmd *cobra.Command, args []string) {
		outfile := path.Base(args[0]) + ".out"
		p := parser.NewParser()
		file := loadInputFile(args[0])
		parsed, err := p.Parse(file)
		exitOnError(err, "parsing file")
		opt := optimizers.NewCompoundOptimizer()
		err = opt.Optimize(parsed)
		exitOnError(err, "performing optimisation")
		gen := generators.YololGenerator{}
		generated := gen.Generate(parsed)
		ioutil.WriteFile(outfile, []byte(generated), 0700)
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(optimizeCmd)
	optimizeCmd.Flags().StringVarP(&outputFile, "out", "o", "<inputfile>.out", "The output file")
}
