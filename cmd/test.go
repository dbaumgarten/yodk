package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dbaumgarten/yodk/pkg/testing"
	"github.com/spf13/cobra"
)

// testCmd represents the format command
var testCmd = &cobra.Command{
	Use:   "test [testfile] [testfile] ...",
	Short: "Run tests",

	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			file := loadInputFile(arg)
			absolutePath, _ := filepath.Abs(arg)
			test, err := testing.Parse([]byte(file), absolutePath)
			exitOnError(err, "loading test case")
			fmt.Println("Running file: " + arg)
			fails := test.Run(func(c testing.Case) {
				fmt.Println("  Running case: " + c.Name)
			})
			if len(fails) == 0 {
				fmt.Println("Tests OK")
			} else {
				fmt.Println("There were errors when running the tests:")
				for _, err := range fails {
					fmt.Println(err)
				}
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
