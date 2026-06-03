package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var fillCmd = &cobra.Command{
	Use:   "fill <uid> <value>",
	Short: "Fill an input or select an option",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := daemon.Call(daemon.Request{
			Command: "fill",
			Args: daemon.MarshalArgs(daemon.FillArgs{
				UID:   args[0],
				Value: args[1],
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

var fillFormCmd = &cobra.Command{
	Use:   "fill-form",
	Short: "Fill multiple form fields at once",
	Run: func(cmd *cobra.Command, args []string) {
		fields, _ := cmd.Flags().GetStringSlice("field")
		if len(fields) == 0 {
			output.PrintError(fmt.Errorf("at least one --field uid=value required"))
		}

		fieldMap := make(map[string]string)
		for _, f := range fields {
			parts := strings.SplitN(f, "=", 2)
			if len(parts) != 2 {
				output.PrintError(fmt.Errorf("invalid field format %q, expected uid=value", f))
			}
			fieldMap[parts[0]] = parts[1]
		}

		resp, err := daemon.Call(daemon.Request{
			Command: "fill-form",
			Args:    daemon.MarshalArgs(daemon.FillFormArgs{Fields: fieldMap}),
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
	fillCmd.Flags().Bool("snapshot", false, "Include snapshot after fill")
	fillFormCmd.Flags().StringSlice("field", nil, "Field as uid=value (repeatable)")
	fillFormCmd.Flags().Bool("snapshot", false, "Include snapshot after fill")
	rootCmd.AddCommand(fillCmd)
	rootCmd.AddCommand(fillFormCmd)
}
