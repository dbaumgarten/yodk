package cmd

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/spf13/cobra"
)

var debugLog bool

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify [file]+",
	Short: "Check if a yolol programm is valid",
	Long:  `Tries to parse a yolol file`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, filepath := range args {
			p := parser.NewParser()
			p.DebugLog = debugLog
			file := loadInputFile(filepath)
			_, errs := p.Parse(file)
			exitOnError(errs, "parsing file")
			fmt.Println(filepath, "is valid")
		}
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().BoolVarP(&debugLog, "debug", "d", false, "Print debug logs while parsing")
}
