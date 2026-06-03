package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var networkCmd = &cobra.Command{
	Use:   "network [reqid]",
	Short: "List or get network requests",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "network",
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
	networkCmd.Flags().StringSlice("type", nil, "Filter by resource type")
	networkCmd.Flags().Int("page-size", 0, "Max requests to return")
	networkCmd.Flags().Int("page", 0, "Page number")
	networkCmd.Flags().Bool("preserved", false, "Include preserved requests")
	networkCmd.Flags().String("req-file", "", "Save request body to file")
	networkCmd.Flags().String("res-file", "", "Save response body to file")
	rootCmd.AddCommand(networkCmd)
}
