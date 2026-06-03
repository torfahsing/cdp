package cmd

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/cdp/internal/daemon"
	"github.com/user/cdp/internal/output"
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Capture page or element screenshot",
	Run: func(cmd *cobra.Command, args []string) {
		filePath, _ := cmd.Flags().GetString("file")
		fullPage, _ := cmd.Flags().GetBool("full-page")
		format, _ := cmd.Flags().GetString("format")
		quality, _ := cmd.Flags().GetInt("quality")
		uid, _ := cmd.Flags().GetString("uid")

		resp, err := daemon.Call(daemon.Request{
			Command: "screenshot",
			Args: daemon.MarshalArgs(daemon.ScreenshotArgs{
				Format:   format,
				Quality:  quality,
				FullPage: fullPage,
				UID:      uid,
			}),
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

		buf, _ := base64.StdEncoding.DecodeString(resp.Binary)

		if filePath != "" {
			if err := os.WriteFile(filePath, buf, 0o644); err != nil {
				output.PrintError(err)
			}
			output.Print(flagJSON,
				fmt.Sprintf("Saved screenshot to %s (%dKB)\n", filePath, len(buf)/1024),
				map[string]any{"file": filePath, "size": len(buf)})
			return
		}

		if flagJSON {
			output.Print(true, "", map[string]string{
				"data":   resp.Binary,
				"format": format,
				"size":   fmt.Sprintf("%d", len(buf)),
			})
		} else {
			os.Stdout.Write(buf)
		}
	},
}

func init() {
	screenshotCmd.Flags().String("file", "", "Save to file path")
	screenshotCmd.Flags().Bool("full-page", false, "Capture full page")
	screenshotCmd.Flags().String("format", "png", "Image format: png, jpeg, webp")
	screenshotCmd.Flags().Int("quality", 80, "Image quality (jpeg/webp)")
	screenshotCmd.Flags().String("uid", "", "Element UID to screenshot")
	rootCmd.AddCommand(screenshotCmd)
}
