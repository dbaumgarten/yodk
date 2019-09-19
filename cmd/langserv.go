package cmd

import (
	"log"
	"os"

	"github.com/dbaumgarten/yodk/langserver"
	"github.com/spf13/cobra"
)

// langservCmd represents the langserv command
var langservCmd = &cobra.Command{
	Use:   "langserv",
	Short: "Start language server",
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			os.Exit(1)
		}
		defer f.Close()

		log.SetOutput(f)
		log.SetFlags(log.Ltime | log.Llongfile)
		log.Println("Language server started")

		/*
			r := langserver.NewStdioStream()
			for {
				data, err := r.ReadObject(context.Background())
				if err != nil {
					log.Println(err)
				}
				log.Println(string(data))
			}
		*/

		err = langserver.Serve()
		if err != nil {
			log.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(langservCmd)
}
