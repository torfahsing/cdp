package cmd

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
)

var (
	flagPort       int
	flagJSON       bool
	flagTimeout    int
	flagDaemonServe bool
)

var rootCmd = &cobra.Command{
	Use:   "cdp",
	Short: "Chrome DevTools Protocol CLI",
	Long:  "CLI tool for controlling Chrome via the Chrome DevTools Protocol.",
}

func Execute() {
	// Check for --daemon-serve before cobra parses subcommands
	for _, arg := range os.Args[1:] {
		if arg == "--daemon-serve" {
			flagDaemonServe = true
			break
		}
	}
	if flagDaemonServe {
		rootCmd.ParseFlags(os.Args[1:])
		d := daemon.New(flagPort)
		if err := d.Serve(); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	defaultPort := 9222
	if v := os.Getenv("CDP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			defaultPort = p
		}
	}
	rootCmd.PersistentFlags().IntVar(&flagPort, "port", defaultPort, "Chrome remote debugging port")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().IntVar(&flagTimeout, "timeout", 30000, "Timeout in milliseconds")
	rootCmd.PersistentFlags().BoolVar(&flagDaemonServe, "daemon-serve", false, "Run as daemon (internal)")
	rootCmd.PersistentFlags().MarkHidden("daemon-serve")
}
