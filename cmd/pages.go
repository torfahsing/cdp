package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var pagesCmd = &cobra.Command{
	Use:   "pages",
	Short: "List open pages",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "pages",
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

var selectCmd = &cobra.Command{
	Use:   "select <page-id>",
	Short: "Select a page for subsequent commands",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		focus, _ := cmd.Flags().GetBool("focus")

		resp, err := daemon.Call(daemon.Request{
			Command: "select",
			Args: daemon.MarshalArgs(daemon.SelectArgs{
				PageID: args[0],
				Focus:  focus,
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

var closeCmd = &cobra.Command{
	Use:   "close <page-id>",
	Short: "Close a page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "close",
			Args:    daemon.MarshalArgs(daemon.ClosePageArgs{PageID: args[0]}),
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

var openCmd = &cobra.Command{
	Use:   "open <url>",
	Short: "Open a new tab",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		background, _ := cmd.Flags().GetBool("background")

		resp, err := daemon.Call(daemon.Request{
			Command: "open",
			Args: daemon.MarshalArgs(daemon.OpenPageArgs{
				URL:        args[0],
				Background: background,
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
	selectCmd.Flags().Bool("focus", false, "Bring the page to front")
	openCmd.Flags().Bool("background", false, "Open in background")
	openCmd.Flags().String("isolated", "", "Isolated browser context name")
	openCmd.Flags().Int("timeout", 0, "Navigation timeout in ms")

	rootCmd.AddCommand(pagesCmd)
	rootCmd.AddCommand(selectCmd)
	rootCmd.AddCommand(closeCmd)
	rootCmd.AddCommand(openCmd)
}
