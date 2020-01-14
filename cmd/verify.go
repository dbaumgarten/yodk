package cmd

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/spf13/cobra"
)

var debugLog bool

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify [file]",
	Short: "Check if a (y/n)olol programm compiles correctly",
	Long:  `Try to compile the file and run checks on it`,
	Run: func(cmd *cobra.Command, args []string) {
		p := parser.NewParser()
		p.DebugLog = debugLog
		file := loadInputFile(args[0])
		_, errs := p.Parse(file)

		exitOnError(errs, "parsing file")
		fmt.Println("File is valid")

	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().BoolVarP(&debugLog, "debug", "d", false, "Print debug logs while parsing")
}
