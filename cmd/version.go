package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// YodkVersion contains the version of this binary and is set on build time
var YodkVersion = "UNVERSIONED BUILD"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version of the yodk binary",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(YodkVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
