package cmd

import (
	"errors"
	"github.com/scryner/swagroller/handler"
	"github.com/spf13/cobra"
)

const defaultOutput = "out"

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build swagger static html page",
	RunE: func(cmd *cobra.Command, args []string) error {
		// get filename
		if len(args) < 1 {
			return errors.New("filename is required")
		}
		filename := args[0]

		// get output directory
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = defaultOutput
		}

		// do build
		return handler.DoBuild(filename, output)
	},
}

func init() {
	buildCmd.Flags().StringP("output", "o", defaultOutput, "Specify output directory")
}
