package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var consoleCmd = &cobra.Command{
	Use:   "console [msgid]",
	Short: "List or get console messages",
	Run: func(cmd *cobra.Command, args []string) {
		types, _ := cmd.Flags().GetStringSlice("type")

		resp, err := daemon.Call(daemon.Request{
			Command: "console",
			Args:    daemon.MarshalArgs(daemon.ConsoleArgs{Types: types}),
			Port:    flagPort,
			Timeout: flagTimeout,
			JSON:    flagJSON,
		})
		if err != nil {
			output.PrintError(err)
		}
		output.PrintResponse(flagJSON, resp)
	},
}

func init() {
	consoleCmd.Flags().StringSlice("type", nil, "Filter by message type")
	consoleCmd.Flags().Int("page-size", 0, "Max messages to return")
	consoleCmd.Flags().Int("page", 0, "Page number")
	consoleCmd.Flags().Bool("preserved", false, "Include preserved messages")
	rootCmd.AddCommand(consoleCmd)
}
