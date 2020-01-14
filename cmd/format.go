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

// formatCmd represents the format command
var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format a code-file",

	Run: func(cmd *cobra.Command, args []string) {
		outfile := strings.Replace(args[0], path.Ext(args[0]), "", -1) + "-fmt" + path.Ext(args[0])
		file := loadInputFile(args[0])
		generated := ""
		var err error
		if strings.HasSuffix(args[0], ".yolol") {
			p := parser.NewParser()
			parsed, errs := p.Parse(file)
			if errs != nil {
				exitOnError(errs, "parsing file")
			}
			gen := parser.Printer{}
			generated, err = gen.Print(parsed)
			exitOnError(err, "generating code")
		} else if strings.HasSuffix(args[0], ".nolol") {
			p := nolol.NewParser()
			parsed, errs := p.Parse(file)
			if errs != nil {
				exitOnError(errs, "parsing file")
			}
			printer := nolol.NewPrinter()
			generated, err = printer.Print(parsed)
			exitOnError(err, "generating code")
		} else {
			exitOnError(fmt.Errorf("Unsupported file-type"), "opening file")
		}
		ioutil.WriteFile(outfile, []byte(generated), 0700)
	},
}

func init() {
	rootCmd.AddCommand(formatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// formatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// formatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
