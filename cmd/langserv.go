package cmd

import (
	"context"
	"log"
	"os"

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
		if logfile != "" {
			stream.Log = true
			f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				os.Exit(1)
			}
			defer f.Close()

			log.SetOutput(f)
			log.SetFlags(log.Ltime | log.Lshortfile)
			log.Println("Language server started")
		}
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
