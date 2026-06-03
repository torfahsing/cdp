package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var dialogCmd = &cobra.Command{
	Use:   "dialog <accept|dismiss>",
	Short: "Handle a browser dialog",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]
		if action != "accept" && action != "dismiss" {
			output.PrintError(fmt.Errorf("action must be 'accept' or 'dismiss'"))
		}

		promptText, _ := cmd.Flags().GetString("text")

		resp, err := daemon.Call(daemon.Request{
			Command: "dialog",
			Args: daemon.MarshalArgs(daemon.DialogArgs{
				Action: action,
				Text:   promptText,
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
	dialogCmd.Flags().String("text", "", "Prompt text for dialog")
	rootCmd.AddCommand(dialogCmd)
}
