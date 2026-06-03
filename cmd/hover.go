package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var hoverCmd = &cobra.Command{
	Use:   "hover <uid>",
	Short: "Hover over an element",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "hover",
			Args:    daemon.MarshalArgs(daemon.HoverArgs{UID: args[0]}),
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
	hoverCmd.Flags().Bool("snapshot", false, "Include snapshot after hover")
	rootCmd.AddCommand(hoverCmd)
}
