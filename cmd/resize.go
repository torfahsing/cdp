package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var resizeCmd = &cobra.Command{
	Use:   "resize <width> <height>",
	Short: "Resize page viewport",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		w, err := strconv.Atoi(args[0])
		if err != nil {
			output.PrintError(fmt.Errorf("invalid width: %s", args[0]))
		}
		h, err := strconv.Atoi(args[1])
		if err != nil {
			output.PrintError(fmt.Errorf("invalid height: %s", args[1]))
		}

		resp, err := daemon.Call(daemon.Request{
			Command: "resize",
			Args: daemon.MarshalArgs(daemon.ResizeArgs{
				Width:  w,
				Height: h,
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
	rootCmd.AddCommand(resizeCmd)
}
