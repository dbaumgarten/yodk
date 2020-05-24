package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/util"
	"github.com/spf13/cobra"
)

var formatReadable bool

// formatCmd represents the format command
var formatCmd = &cobra.Command{
	Use:   "format [file]+",
	Short: "Format yolol/nolol files",

	Run: func(cmd *cobra.Command, args []string) {
		for _, file := range args {
			fmt.Println("Formatting file:", file)
			format(file)
		}
	},
}

func format(filepath string) {
	file := loadInputFile(filepath)
	generated := ""
	var err error
	if strings.HasSuffix(filepath, ".yolol") {
		p := parser.NewParser()
		parsed, errs := p.Parse(file)
		if errs != nil {
			exitOnError(errs, "parsing file")
		}
		gen := parser.Printer{}
		if formatReadable {
			gen.Mode = parser.PrintermodeReadable
		} else {
			gen.Mode = parser.PrintermodeCompact
		}
		generated, err = gen.Print(parsed)
		exitOnError(err, "generating code")
		err = util.CheckForFormattingErrorYolol(parsed, generated)
		exitOnError(err, "formatting code")
	} else if strings.HasSuffix(filepath, ".nolol") {
		p := nolol.NewParser()
		parsed, errs := p.Parse(file)
		if errs != nil {
			exitOnError(errs, "parsing file")
		}
		printer := nolol.NewPrinter()
		generated, err = printer.Print(parsed)
		exitOnError(err, "generating code")
		err = util.CheckForFormattingErrorNolol(parsed, generated)
		exitOnError(err, "formatting code")
	} else {
		exitOnError(fmt.Errorf("Unsupported file-type"), "opening file")
	}

	ioutil.WriteFile(filepath, []byte(generated), 0700)
}

func init() {
	rootCmd.AddCommand(formatCmd)
	formatCmd.Flags().BoolVarP(&formatReadable, "readable", "r", false, "Do add extra spaces to improve readability")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// formatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// formatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
