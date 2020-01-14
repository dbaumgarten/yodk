package cmd

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/optimizers"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/spf13/cobra"
)

var outputFile string

// optimizeCmd represents the compile command
var optimizeCmd = &cobra.Command{
	Use:   "optimize [file]",
	Short: "Optimize a yolo programm",
	Long:  `Perform optimizations on a yolol-programm`,
	Run: func(cmd *cobra.Command, args []string) {
		outfile := strings.Replace(args[0], path.Ext(args[0]), "", -1) + "-opt" + path.Ext(args[0])
		p := parser.NewParser()
		file := loadInputFile(args[0])
		parsed, errs := p.Parse(file)
		if errs != nil {
			exitOnError(errs, "parsing file")
		}
		opt := optimizers.NewCompoundOptimizer()
		err := opt.Optimize(parsed)
		exitOnError(err, "performing optimisation")
		gen := parser.Printer{}
		generated, err := gen.Print(parsed)
		exitOnError(err, "generating code")
		ioutil.WriteFile(outfile, []byte(generated), 0700)
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(optimizeCmd)
	optimizeCmd.Flags().StringVarP(&outputFile, "out", "o", "<inputfile>.out", "The output file")
}
