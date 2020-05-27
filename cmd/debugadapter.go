package cmd

import (
	"log"
	"os"

	"github.com/dbaumgarten/yodk/pkg/debug"
	"github.com/spf13/cobra"
)

// debugAdapterCmd represents the debugadapter command
var debugAdapterCmd = &cobra.Command{
	Use:   "debugadapter",
	Short: "Start a vscode debugadapter",
	Run: func(cmd *cobra.Command, args []string) {

		stream := debug.StdioReadWriteCloser{}
		configureLogging()

		debug.HandleConnection(stream, debug.NewYODKHandler)
	},
}

func init() {
	rootCmd.AddCommand(debugAdapterCmd)
	debugAdapterCmd.Flags().StringVar(&logfile, "logfile", "", "Name of the file to log into")
}

// configures logging according to provided flags. if logging is enabled returns true
func configureLogging() bool {
	if logfile != "" {
		f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			os.Exit(1)
		}

		log.SetOutput(f)
		log.SetFlags(log.Ltime | log.Lshortfile)
		return true
	}
	return false
}
