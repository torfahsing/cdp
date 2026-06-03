package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var waitCmd = &cobra.Command{
	Use:   "wait <text> [text...]",
	Short: "Wait for text to appear on page",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "wait",
			Args:    daemon.MarshalArgs(daemon.WaitArgs{Texts: args}),
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
	rootCmd.AddCommand(waitCmd)
}
