package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var typeCmd = &cobra.Command{
	Use:   "type <text>",
	Short: "Type text into the focused element",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		submitKey, _ := cmd.Flags().GetString("submit")

		resp, err := daemon.Call(daemon.Request{
			Command: "type",
			Args: daemon.MarshalArgs(daemon.TypeArgs{
				Text:      args[0],
				SubmitKey: submitKey,
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
	typeCmd.Flags().String("submit", "", "Key to press after typing (e.g. Enter)")
	rootCmd.AddCommand(typeCmd)
}
