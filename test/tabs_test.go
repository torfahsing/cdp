//go:build integration

package test

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runCDP(t *testing.T, args ...string) (string, error) {
	t.Helper()
	bin, err := filepath.Abs(filepath.Join("..", "cdp"))
	if err != nil {
		t.Fatalf("resolve binary: %v", err)
	}
	out, err := exec.Command(bin, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func pageIDs(t *testing.T) map[string]bool {
	t.Helper()
	out, err := runCDP(t, "pages", "--json")
	if err != nil {
		t.Fatalf("cdp pages --json failed: %v\n%s", err, out)
	}
	var pages []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &pages); err != nil {
		t.Fatalf("parse pages JSON: %v\n%s", err, out)
	}
	ids := make(map[string]bool, len(pages))
	for _, p := range pages {
		ids[p.ID] = true
	}
	return ids
}

// TestSelectDoesNotCloseTabs verifies that switching between tabs via
// cdp select does not close any user tabs.
func TestSelectDoesNotCloseTabs(t *testing.T) {
	// Start fresh
	_, err := runCDP(t, "disconnect")
	if err != nil {
		t.Fatalf("cdp disconnect failed: %v", err)
	}

	// Open two test pages in the background
	url := "file://" + filepath.Join("fixtures", "test.html")
	out1, err := runCDP(t, "open", url, "--background")
	if err != nil {
		t.Fatalf("cdp open 1 failed: %v\n%s", err, out1)
	}
	out2, err := runCDP(t, "open", url, "--background")
	if err != nil {
		t.Fatalf("cdp open 2 failed: %v\n%s", err, out2)
	}

	// Snapshot IDs before switching
	before := pageIDs(t)
	if len(before) < 2 {
		t.Fatalf("expected at least 2 pages before switching, got %d", len(before))
	}

	// Extract page IDs from the open output (format: "Opened <url> (<id>)\n")
	var idA, idB string
	for _, out := range []string{out1, out2} {
		// The id is in parentheses at the end: "Opened ... (abc123)\n"
		parts := strings.Split(out, "(")
		if len(parts) < 2 {
			continue
		}
		id := strings.TrimSpace(strings.TrimSuffix(parts[1], ")"))
		if idA == "" {
			idA = id
		} else {
			idB = id
		}
	}
	if idA == "" || idB == "" {
		t.Fatalf("could not extract page IDs from open output: %q %q", out1, out2)
	}

	// Switch between tabs multiple times
	for _, id := range []string{idA, idB, idA, idA} {
		out, err := runCDP(t, "select", id, "--focus")
		if err != nil {
			t.Fatalf("cdp select %s failed: %v\n%s", id, err, out)
		}
	}

	// All original IDs must still be present
	after := pageIDs(t)
	for id := range before {
		if !after[id] {
			t.Errorf("page %s was closed during tab switching", id)
		}
	}
	if len(after) != len(before) {
		t.Errorf("page count changed: before=%d after=%d", len(before), len(after))
	}
}

// TestShutdownDoesNotCloseTabs verifies that daemon shutdown does not
// close user-opened tabs.
func TestShutdownDoesNotCloseTabs(t *testing.T) {
	// Start fresh
	_, err := runCDP(t, "disconnect")
	if err != nil {
		t.Fatalf("cdp disconnect failed: %v", err)
	}

	// Open two test pages in the background
	url := "file://" + filepath.Join("fixtures", "test.html")
	out1, err := runCDP(t, "open", url, "--background")
	if err != nil {
		t.Fatalf("cdp open 1 failed: %v\n%s", err, out1)
	}
	out2, err := runCDP(t, "open", url, "--background")
	if err != nil {
		t.Fatalf("cdp open 2 failed: %v\n%s", err, out2)
	}

	// Extract user-tab IDs from open output (the scratch tab's ID
	// will change after daemon restart, so we track only user tabs).
	userTabIDs := make(map[string]bool)
	for _, out := range []string{out1, out2} {
		parts := strings.Split(out, "(")
		if len(parts) < 2 {
			continue
		}
		id := strings.TrimSpace(strings.TrimSuffix(parts[1], ")"))
		if id != "" {
			userTabIDs[id] = true
		}
	}
	if len(userTabIDs) < 2 {
		t.Fatalf("expected at least 2 user-tab IDs, got %d", len(userTabIDs))
	}

	// Shutdown the daemon
	out, err := runCDP(t, "shutdown")
	if err != nil {
		t.Fatalf("cdp shutdown failed: %v\n%s", err, out)
	}

	// Wait briefly for the daemon to restart itself (it is re-launched by
	// the test runner's process model; if the daemon is already dead,
	// subsequent commands will fail with a clear error).
	// Reconnect by running a simple command.
	out, err = runCDP(t, "pages", "--json")
	if err != nil {
		t.Fatalf("cdp pages after shutdown failed: %v\n%s", err, out)
	}

	// Parse the post-shutdown page list
	var pages []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &pages); err != nil {
		t.Fatalf("parse pages JSON after shutdown: %v\n%s", err, out)
	}
	after := make(map[string]bool, len(pages))
	for _, p := range pages {
		after[p.ID] = true
	}

	// All user-tab IDs must survive the shutdown (the scratch tab
	// gets a new ID on daemon restart, so we only check user tabs).
	for id := range userTabIDs {
		if !after[id] {
			t.Errorf("page %s was closed during shutdown", id)
		}
	}
}

// TestCloseStillClosesTab verifies that the explicit cdp close command
// still closes a tab (regression guard: implicit closes were removed but
// the explicit close path must still work).
func TestCloseStillClosesTab(t *testing.T) {
	// Start fresh
	_, err := runCDP(t, "disconnect")
	if err != nil {
		t.Fatalf("cdp disconnect failed: %v", err)
	}

	// Open two test pages in the background
	url := "file://" + filepath.Join("fixtures", "test.html")
	out1, err := runCDP(t, "open", url, "--background")
	if err != nil {
		t.Fatalf("cdp open 1 failed: %v\n%s", err, out1)
	}
	out2, err := runCDP(t, "open", url, "--background")
	if err != nil {
		t.Fatalf("cdp open 2 failed: %v\n%s", err, out2)
	}

	// Extract the first page ID to close
	var idToClose string
	for _, out := range []string{out1, out2} {
		parts := strings.Split(out, "(")
		if len(parts) < 2 {
			continue
		}
		idToClose = strings.TrimSpace(strings.TrimSuffix(parts[1], ")"))
		break
	}
	if idToClose == "" {
		t.Fatalf("could not extract page ID from open output")
	}

	// Close one tab explicitly
	out, err := runCDP(t, "close", idToClose)
	if err != nil {
		t.Fatalf("cdp close %s failed: %v\n%s", idToClose, err, out)
	}
	if !strings.Contains(out, "Closed page "+idToClose) {
		t.Errorf("expected 'Closed page %s' output, got: %s", idToClose, out)
	}

	// The closed ID must be gone, but the other must survive
	after := pageIDs(t)
	if after[idToClose] {
		t.Errorf("page %s was not closed by cdp close", idToClose)
	}
	// At least one page should remain (the other test tab)
	if len(after) == 0 {
		t.Error("all pages were closed; expected at least one survivor")
	}
}
