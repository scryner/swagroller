package cmd

import (
	"errors"
	"github.com/scryner/swagroller/handler"
	"github.com/spf13/cobra"
)

const defaultListenPort = 8080

var servCmd = &cobra.Command{
	Use:   "serv",
	Short: "Serve swagger static html page",
	RunE: func(cmd *cobra.Command, args []string) error {
		// get filename
		if len(args) < 1 {
			return errors.New("filename is required")
		}
		filename := args[0]

		// get listen port
		listenPort, _ := cmd.Flags().GetInt("port")
		if listenPort == 0 {
			listenPort = defaultListenPort
		}

		// get open browser option
		openBrowser, _ := cmd.Flags().GetBool("open-browser")

		return handler.DoServer(filename, listenPort, openBrowser)
	},
}

func init() {
	servCmd.Flags().IntP("port", "p", defaultListenPort, "Specify listen port")
	servCmd.Flags().Bool("open-browser", true, "Open browser when app starting")
}
