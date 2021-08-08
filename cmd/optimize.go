package cmd

import (
	"fmt"
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
	Use:   "optimize [file]+",
	Short: "Optimize yolo programs",
	Long:  `Perform optimizations on yolol-programs`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, file := range args {
			fmt.Println("Optimizing file:", file)
			optimize(file)
		}

	},
	Args: cobra.MinimumNArgs(1),
}

func optimize(filepath string) {
	var outfile string
	if outputFile != "" {
		outfile = outputFile
	} else {
		outfile = strings.Replace(filepath, path.Ext(filepath), "", -1) + ".opt" + path.Ext(filepath)
	}
	p := parser.NewParser()
	file := loadInputFile(filepath)
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
}

func init() {
	rootCmd.AddCommand(optimizeCmd)
	optimizeCmd.Flags().StringVarP(&outputFile, "out", "o", "", "The output file")
}
