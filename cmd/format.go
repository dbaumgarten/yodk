package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/util"
	"github.com/spf13/cobra"
)

var formatMode string

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

	if formatMode != "readable" && formatMode != "compact" && formatMode != "spaceless" {
		fmt.Println("Fomatting mode must be one of: readable|compact|spaceless")
		os.Exit(1)
	}

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
		switch formatMode {
		case "readable":
			gen.Mode = parser.PrintermodeReadable
			break
		case "compact":
			gen.Mode = parser.PrintermodeCompact
			break
		case "spaceless":
			gen.Mode = parser.PrintermodeSpaceless
			break
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
	formatCmd.Flags().StringVarP(&formatMode, "mode", "m", "compact", "Formatting mode [readable,compact,spaceless]")
}
