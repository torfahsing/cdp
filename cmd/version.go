package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags "-X github.com/user/cdp/cmd.Version=..."
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the cdp version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
