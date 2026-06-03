package cmd

import (
	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var emulateCmd = &cobra.Command{
	Use:   "emulate",
	Short: "Emulate device features",
	Run: func(cmd *cobra.Command, args []string) {
		viewport, _ := cmd.Flags().GetString("viewport")
		userAgent, _ := cmd.Flags().GetString("user-agent")
		colorScheme, _ := cmd.Flags().GetString("color-scheme")
		cpuThrottle, _ := cmd.Flags().GetFloat64("cpu-throttle")
		network, _ := cmd.Flags().GetString("network")
		geo, _ := cmd.Flags().GetString("geo")

		resp, err := daemon.Call(daemon.Request{
			Command: "emulate",
			Args: daemon.MarshalArgs(daemon.EmulateArgs{
				Viewport:    viewport,
				ColorScheme: colorScheme,
				UserAgent:   userAgent,
				CPUThrottle: cpuThrottle,
				Geo:         geo,
				Network:     network,
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
	emulateCmd.Flags().String("viewport", "", "Viewport as WxHxDPR")
	emulateCmd.Flags().String("user-agent", "", "User agent string")
	emulateCmd.Flags().String("color-scheme", "", "dark, light, or auto")
	emulateCmd.Flags().Float64("cpu-throttle", 0, "CPU throttle rate (1=normal)")
	emulateCmd.Flags().String("network", "", "Network condition: Offline, Slow 3G, Fast 3G, Slow 4G, Fast 4G")
	emulateCmd.Flags().String("geo", "", "Geolocation as latxlng")
	rootCmd.AddCommand(emulateCmd)
}
