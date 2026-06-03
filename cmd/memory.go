package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var memoryCmd = &cobra.Command{
	Use:   "memory <filepath>",
	Short: "Capture heap snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "memory",
			Args:    daemon.MarshalArgs(daemon.MemoryArgs{FilePath: args[0]}),
			Port:    flagPort,
			Timeout: 120000,
			JSON:    flagJSON,
		})
		if err != nil {
			output.PrintError(err)
		}
		output.PrintResponse(flagJSON, resp)
	},
}

func init() {
	rootCmd.AddCommand(memoryCmd)
}
