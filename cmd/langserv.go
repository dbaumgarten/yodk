package cmd

import (
	"context"
	"log"

	"github.com/dbaumgarten/yodk/pkg/langserver"
	"github.com/spf13/cobra"
)

var logfile string

// langservCmd represents the langserv command
var langservCmd = &cobra.Command{
	Use:   "langserv",
	Short: "Start language server",
	Run: func(cmd *cobra.Command, args []string) {

		stream := langserver.NewStdioStream()
		stream.Log = configureLogging()
		err := langserver.Run(context.Background(), stream)
		if err != nil {
			log.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(langservCmd)
	langservCmd.Flags().StringVar(&logfile, "logfile", "", "Name of the file to log into")
}
