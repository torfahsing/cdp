package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/user/cdp/internal/daemon"
)

func Print(jsonMode bool, textOutput string, jsonData any) {
	if jsonMode {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(jsonData)
	} else {
		fmt.Print(textOutput)
	}
}

func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func PrintResponse(jsonMode bool, resp *daemon.Response) {
	if !resp.OK {
		PrintError(fmt.Errorf("%s", resp.Error))
	}
	if jsonMode {
		if resp.Data != nil {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(json.RawMessage(resp.Data))
		}
	} else {
		fmt.Print(resp.Text)
	}
}
