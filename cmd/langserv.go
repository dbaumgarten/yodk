package cmd

import (
	"context"
	"log"

	"github.com/dbaumgarten/yodk/pkg/langserver"
	"github.com/spf13/cobra"
)

var logfile string
var hotkeys bool

// langservCmd represents the langserv command
var langservCmd = &cobra.Command{
	Use:   "langserv",
	Short: "Start language server",
	Run: func(cmd *cobra.Command, args []string) {

		stream := langserver.NewStdioStream()
		configureFileLogging()
		stream.Log = debugLog
		err := langserver.Run(context.Background(), stream, hotkeys)
		if err != nil {
			log.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(langservCmd)
	langservCmd.Flags().StringVar(&logfile, "logfile", "", "Name of the file to log into. Defaults to stderr")
	langservCmd.Flags().BoolVarP(&debugLog, "debug", "d", false, "Enable verbose debug-logging")
	langservCmd.Flags().BoolVar(&hotkeys, "hotkeys", true, "Enable system-wide hotkeys for auto-typing")
}
