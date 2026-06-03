package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var evaluateCmd = &cobra.Command{
	Use:   "eval <javascript>",
	Short: "Evaluate JavaScript in the page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dialog, _ := cmd.Flags().GetString("dialog")

		resp, err := daemon.Call(daemon.Request{
			Command: "eval",
			Args: daemon.MarshalArgs(daemon.EvalArgs{
				Script: args[0],
				Dialog: dialog,
			}),
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
	evaluateCmd.Flags().String("dialog", "", "Handle dialog: accept, dismiss, or prompt text")
	rootCmd.AddCommand(evaluateCmd)
}
