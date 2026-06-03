package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var dragCmd = &cobra.Command{
	Use:   "drag <from-uid> <to-uid>",
	Short: "Drag an element onto another",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "drag",
			Args: daemon.MarshalArgs(daemon.DragArgs{
				FromUID: args[0],
				ToUID:   args[1],
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
	dragCmd.Flags().Bool("snapshot", false, "Include snapshot after drag")
	rootCmd.AddCommand(dragCmd)
}
