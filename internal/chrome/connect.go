package chrome

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type versionResponse struct {
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func Connect(ctx context.Context, port int) (context.Context, context.CancelFunc, error) {
	// Try DevToolsActivePort first — it's a local file read (no network, no prompts).
	// This is the primary method when using chrome://inspect remote debugging.
	wsURL, err := discoverViaActivePort()
	if err != nil {
		// Fall back to HTTP /json/version (works with --remote-debugging-port)
		wsURL, err = discoverViaHTTP(port)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot connect to Chrome on port %d — is remote debugging enabled?\n  tried: DevToolsActivePort file (not found or unreadable)\n  tried: HTTP http://127.0.0.1:%d/json/version (failed)", port, port)
		}
	}

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, wsURL)
	return allocCtx, allocCancel, nil
}

func discoverViaHTTP(port int) (string, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/json/version", port)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var ver versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&ver); err != nil {
		return "", err
	}
	if ver.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("empty webSocketDebuggerUrl")
	}
	return ver.WebSocketDebuggerURL, nil
}

func discoverViaActivePort() (string, error) {
	for _, dir := range chromeUserDataDirs() {
		path := filepath.Join(dir, "DevToolsActivePort")
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		var port, wsPath string
		if scanner.Scan() {
			port = strings.TrimSpace(scanner.Text())
		}
		if scanner.Scan() {
			wsPath = strings.TrimSpace(scanner.Text())
		}
		if port != "" && wsPath != "" {
			return fmt.Sprintf("ws://127.0.0.1:%s%s", port, wsPath), nil
		}
	}
	return "", fmt.Errorf("DevToolsActivePort not found")
}

func chromeUserDataDirs() []string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library", "Application Support", "Google", "Chrome"),
			filepath.Join(home, "Library", "Application Support", "Google", "Chrome Canary"),
			filepath.Join(home, "Library", "Application Support", "Chromium"),
		}
	case "linux":
		return []string{
			filepath.Join(home, ".config", "google-chrome"),
			filepath.Join(home, ".config", "google-chrome-beta"),
			filepath.Join(home, ".config", "chromium"),
		}
	default: // windows
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Local")
		}
		return []string{
			filepath.Join(appData, "Google", "Chrome", "User Data"),
			filepath.Join(appData, "Chromium", "User Data"),
		}
	}
}
