//go:build integration

package test

import (
	"os/exec"
	"strings"
	"testing"
)

func cdp(args ...string) (string, error) {
	cmd := exec.Command("./cdp", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func TestPagesCommand(t *testing.T) {
	out, err := cdp("pages")
	if err != nil {
		t.Fatalf("cdp pages failed: %v\n%s", err, out)
	}
	if out == "" {
		t.Fatal("expected at least one page")
	}
}

func TestNavigateAndSnapshot(t *testing.T) {
	out, err := cdp("navigate", "https://example.com")
	if err != nil {
		t.Fatalf("navigate failed: %v\n%s", err, out)
	}

	out, err = cdp("snapshot")
	if err != nil {
		t.Fatalf("snapshot failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "uid=") {
		t.Fatalf("snapshot should contain UIDs, got: %s", out)
	}
}

func TestEval(t *testing.T) {
	out, err := cdp("eval", "document.title")
	if err != nil {
		t.Fatalf("eval failed: %v\n%s", err, out)
	}
	if out == "" {
		t.Fatal("expected eval result")
	}
}

func TestScreenshot(t *testing.T) {
	_, err := cdp("screenshot", "--file", "/tmp/cdp-test.png")
	if err != nil {
		t.Fatalf("screenshot failed: %v", err)
	}
}
