//go:build integration

package test

import (
	"encoding/json"
	"os/exec"
	"os"
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
		ID    string `json:"id"`
		URL   string `json:"url"`
		Title string `json:"title"`
	}
	if err := json.Unmarshal([]byte(out), &pages); err != nil {
		t.Fatalf("parse pages JSON: %v\n%s", err, out)
	}
	ids := make(map[string]bool, len(pages))
	for _, p := range pages {
		// Filter the daemon's own scratch about:blank tab, matching
		// handlePages / ensurePage behaviour.
		if p.URL == "about:blank" && p.Title == "" {
			continue
		}
		ids[p.ID] = true
	}
	return ids
}

// testURL returns an absolute file:// URL pointing at test/fixtures/test.html.
// go test ./test/ runs with its working directory set to the test package dir,
// so we join fixtures relative to os.Getwd() (no extra "test" segment).
func testURL() string {
	return "file://" + filepath.Join(os.Getwd(), "fixtures", "test.html")
}

// extractID parses an open-command output line to extract the page ID from its
// trailing "(<id>)" suffix. Returns empty string on parse failure.
func extractID(output string) string {
	parts := strings.Split(output, "(")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(strings.TrimSuffix(parts[1], ")"))
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
	url := testURL()
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
	idA := extractID(out1)
	idB := extractID(out2)
	if idA == "" || idB == "" {
		t.Fatalf("could not extract page IDs from open output: %q %q", out1, out2)
	}

	// Switch between tabs multiple times (includes --focus variant)
	for _, id := range []string{idA, idB, idA, idA} {
		focus := false
		if id == idA {
			focus = true // last select uses --focus
		}
		args := []string{"select", id}
		if focus {
			args = append(args, "--focus")
		}
		out, err := runCDP(t, args...)
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
	url := testURL()
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
		id := extractID(out)
		if id != "" {
			userTabIDs[id] = true
		}
	}
	if len(userTabIDs) < 2 {
		t.Fatalf("expected at least 2 user-tab IDs, got %d", len(userTabIDs))
	}

	// Select one of the user tabs (the other must not be closed).
	idA := extractID(out1)
	if _, err := runCDP(t, "select", idA); err != nil {
		t.Fatalf("cdp select failed: %v", err)
	}

	// Shut down the daemon (cdp disconnect sends the shutdown command
	// over the Unix socket; equivalent to the internal cdp shutdown
	// subcommand).
	out, err := runCDP(t, "disconnect")
	if err != nil {
		t.Fatalf("cdp disconnect failed: %v\n%s", err, out)
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
		ID    string `json:"id"`
		URL   string `json:"url"`
		Title string `json:"title"`
	}
	if err := json.Unmarshal([]byte(out), &pages); err != nil {
		t.Fatalf("parse pages JSON after shutdown: %v\n%s", err, out)
	}
	after := make(map[string]bool, len(pages))
	for _, p := range pages {
		// Filter the daemon's own scratch about:blank tab.
		if p.URL == "about:blank" && p.Title == "" {
			continue
		}
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

	// Capture baseline pages (daemon's own scratch tab only, filtered out later).
	before := pageIDs(t)

	// Open two test pages in the background
	url := testURL()
	_, err = runCDP(t, "open", url, "--background")
	if err != nil {
		t.Fatalf("cdp open 1 failed: %v", err)
	}
	_, err = runCDP(t, "open", url, "--background")
	if err != nil {
		t.Fatalf("cdp open 2 failed: %v", err)
	}

	// Identify the two newly opened pages by comparing against the baseline.
	// cdp pages --json returns full 32-char target IDs which cdp close requires.
	afterOpen := pageIDs(t)
	var newlyOpened []string
	for id := range afterOpen {
		if !before[id] {
			newlyOpened = append(newlyOpened, id)
		}
	}
	if len(newlyOpened) < 2 {
		t.Fatalf("expected 2 newly opened pages, got %d (before=%d after=%d)", len(newlyOpened), len(before), len(afterOpen))
	}
	idToClose := newlyOpened[0] // full 32-char target ID from cdp pages --json

	// Close one tab explicitly
	out, err := runCDP(t, "close", idToClose)
	if err != nil {
		t.Fatalf("cdp close %s failed: %v\n%s", idToClose, err, out)
	}
	// Output contains the short 8-char prefix, not the full ID.
	shortPrefix := idToClose[:8]
	if !strings.Contains(out, "Closed page "+shortPrefix) {
		t.Errorf("expected 'Closed page %s' output, got: %s", shortPrefix, out)
	}

	// The closed full ID must be gone, but the other newly-opened tab must survive.
	after := pageIDs(t)
	if after[idToClose] {
		t.Errorf("page %s was not closed by cdp close", idToClose)
	}
	idSurvivor := newlyOpened[1]
	if !after[idSurvivor] {
		t.Errorf("page %s should have survived cdp close", idSurvivor)
	}
}
