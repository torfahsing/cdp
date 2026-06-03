package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var clickCmd = &cobra.Command{
	Use:   "click <uid>",
	Short: "Click an element by UID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbl, _ := cmd.Flags().GetBool("dbl")
		includeSnapshot, _ := cmd.Flags().GetBool("snapshot")

		resp, err := daemon.Call(daemon.Request{
			Command: "click",
			Args: daemon.MarshalArgs(daemon.ClickArgs{
				UID:      args[0],
				DblClick: dbl,
				Snapshot: includeSnapshot,
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
	clickCmd.Flags().Bool("dbl", false, "Double click")
	clickCmd.Flags().Bool("snapshot", false, "Include snapshot after click")
	rootCmd.AddCommand(clickCmd)
}
