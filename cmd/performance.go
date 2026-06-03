package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var perfCmd = &cobra.Command{
	Use:   "perf",
	Short: "Performance tracing commands",
}

var perfStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a performance trace",
	Run: func(cmd *cobra.Command, args []string) {
		shouldReload, _ := cmd.Flags().GetBool("reload")
		autoStop, _ := cmd.Flags().GetBool("auto-stop")

		resp, err := daemon.Call(daemon.Request{
			Command: "perf-start",
			Args: daemon.MarshalArgs(daemon.PerfStartArgs{
				Reload:   shouldReload,
				AutoStop: autoStop,
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

var perfStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the performance trace",
	Run: func(cmd *cobra.Command, args []string) {
		filePath, _ := cmd.Flags().GetString("file")

		resp, err := daemon.Call(daemon.Request{
			Command: "perf-stop",
			Args:    daemon.MarshalArgs(daemon.PerfStopArgs{FilePath: filePath}),
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

var perfInsightCmd = &cobra.Command{
	Use:   "insight <set-id> <name>",
	Short: "Analyze a performance insight",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "perf-insight",
			Args: daemon.MarshalArgs(daemon.PerfInsightArgs{
				SetID: args[0],
				Name:  args[1],
			}),
			Port:    flagPort,
			Timeout: flagTimeout,
			JSON:    flagJSON,
		})
		if err != nil {
			output.PrintError(err)
		}
		if resp == nil {
			output.PrintError(fmt.Errorf("no response"))
		}
		output.PrintResponse(flagJSON, resp)
	},
}

func init() {
	perfStartCmd.Flags().String("file", "", "Save trace to file")
	perfStartCmd.Flags().Bool("reload", true, "Reload page after starting trace")
	perfStartCmd.Flags().Bool("auto-stop", true, "Auto-stop after page load")
	perfStopCmd.Flags().String("file", "", "Save trace to file")

	perfCmd.AddCommand(perfStartCmd)
	perfCmd.AddCommand(perfStopCmd)
	perfCmd.AddCommand(perfInsightCmd)
	rootCmd.AddCommand(perfCmd)
}
