package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <uid> <filepath>",
	Short: "Upload a file through a file input",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "upload",
			Args: daemon.MarshalArgs(daemon.UploadArgs{
				UID:      args[0],
				FilePath: args[1],
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
	uploadCmd.Flags().Bool("snapshot", false, "Include snapshot after upload")
	rootCmd.AddCommand(uploadCmd)
}
