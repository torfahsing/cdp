package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Take accessibility tree snapshot with UIDs",
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		filePath, _ := cmd.Flags().GetString("file")

		resp, err := daemon.Call(daemon.Request{
			Command: "snapshot",
			Args:    daemon.MarshalArgs(daemon.SnapshotArgs{Verbose: verbose}),
			Port:    flagPort,
			Timeout: flagTimeout,
			JSON:    flagJSON,
		})
		if err != nil {
			output.PrintError(err)
		}
		if !resp.OK {
			output.PrintError(fmt.Errorf("%s", resp.Error))
		}

		if filePath != "" {
			if err := os.WriteFile(filePath, []byte(resp.Text), 0o644); err != nil {
				output.PrintError(err)
			}
			output.Print(flagJSON, "Saved snapshot to "+filePath+"\n",
				map[string]string{"file": filePath})
			return
		}

		output.PrintResponse(flagJSON, resp)
	},
}

func init() {
	snapshotCmd.Flags().Bool("verbose", false, "Include all elements including generic/none roles")
	snapshotCmd.Flags().String("file", "", "Save snapshot to file")
	rootCmd.AddCommand(snapshotCmd)
}
