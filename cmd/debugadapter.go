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
		configureFileLogging()
		debug.StartSession(stream, debug.NewYODKHandler(), debugLog)
	},
}

func init() {
	rootCmd.AddCommand(debugAdapterCmd)
	debugAdapterCmd.Flags().StringVar(&logfile, "logfile", "", "Name of the file to log into. Defaults to stderr")
	debugAdapterCmd.Flags().BoolVarP(&debugLog, "debug", "d", false, "Enable verbose debug-logging")
	debugLog = true
}

// configures logging according to provided flags
func configureFileLogging() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	if logfile != "" {
		f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			os.Exit(1)
		}

		log.SetOutput(f)
	}
}
