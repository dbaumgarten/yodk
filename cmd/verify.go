package cmd

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/validators"
	"github.com/spf13/cobra"
)

var debugLog bool
var chipType string

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify [file]+",
	Short: "Check if a yolol programm is valid",
	Long:  `Tries to parse a yolol file`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, filepath := range args {
			p := parser.NewParser()
			p.SetDebugLog(debugLog)
			file := loadInputFile(filepath)
			parsed, errs := p.Parse(file)
			exitOnError(errs, "parsing file '"+filepath+"'")

			err := validators.ValidateCodeLength(file)
			exitOnError(err, "validating code")

			chip, err := validators.AutoChooseChipType(chipType, filepath)
			exitOnError(err, "determining chip-type")

			err = validators.ValidateAvailableOperations(parsed, chip)
			exitOnError(err, "validating code")

			fmt.Println(filepath, "is valid")
		}
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().BoolVarP(&debugLog, "debug", "d", false, "Print debug logs while parsing")
	verifyCmd.Flags().StringVarP(&chipType, "chip", "c", "auto", "Chip-type to validate for. (auto|professional|advanced|basic)")
}
