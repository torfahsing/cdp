package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var navigateCmd = &cobra.Command{
	Use:   "navigate [url]",
	Short: "Navigate to URL, or back/forward/reload",
	Run: func(cmd *cobra.Command, args []string) {
		back, _ := cmd.Flags().GetBool("back")
		forward, _ := cmd.Flags().GetBool("forward")
		reload, _ := cmd.Flags().GetBool("reload")
		ignoreCache, _ := cmd.Flags().GetBool("ignore-cache")
		waitLoad, _ := cmd.Flags().GetBool("wait-load")

		navArgs := daemon.NavigateArgs{
			Back:        back,
			Forward:     forward,
			Reload:      reload,
			IgnoreCache: ignoreCache,
			WaitLoad:    waitLoad,
		}
		if len(args) > 0 {
			navArgs.URL = args[0]
		}

		resp, err := daemon.Call(daemon.Request{
			Command: "navigate",
			Args:    daemon.MarshalArgs(navArgs),
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
	navigateCmd.Flags().Bool("back", false, "Navigate back")
	navigateCmd.Flags().Bool("forward", false, "Navigate forward")
	navigateCmd.Flags().Bool("reload", false, "Reload page")
	navigateCmd.Flags().Bool("ignore-cache", false, "Ignore cache on reload")
	navigateCmd.Flags().Bool("wait-load", false, "Wait for page load after navigation")
	navigateCmd.Flags().String("init-script", "", "JavaScript to run before page scripts")
	rootCmd.AddCommand(navigateCmd)
}
