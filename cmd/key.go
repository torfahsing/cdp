package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var keyCmd = &cobra.Command{
	Use:   "key <key>",
	Short: "Press a key or key combination",
	Long:  `Press a key or combination (e.g. "Enter", "Control+A", "Shift+1")`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "key",
			Args:    daemon.MarshalArgs(daemon.KeyArgs{Key: args[0]}),
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
	keyCmd.Flags().Bool("snapshot", false, "Include snapshot after keypress")
	rootCmd.AddCommand(keyCmd)
}
