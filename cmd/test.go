// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dbaumgarten/yodk/testing"
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
			fails := testing.RunTest(test, func(c testing.Case) {
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
