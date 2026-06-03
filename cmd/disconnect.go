package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from Chrome and stop the daemon",
	Run: func(cmd *cobra.Command, args []string) {
		if daemon.IsRunning() {
			resp, err := daemon.Call(daemon.Request{Command: "shutdown", Port: flagPort})
			if err == nil && resp.OK {
				output.Print(flagJSON, "Disconnected\n", map[string]string{"status": "disconnected"})
				return
			}
		}

		// Fallback: kill via PID file
		data, err := os.ReadFile(daemon.PidPath())
		if err == nil {
			pid, _ := strconv.Atoi(string(data))
			if pid > 0 {
				daemon.KillProcess(pid)
			}
		}
		os.Remove(daemon.SockPath())
		os.Remove(daemon.PidPath())
		fmt.Println("Disconnected")
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}
