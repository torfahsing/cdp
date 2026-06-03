package chrome

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"

	"github.com/user/cdp/internal/state"
)

type httpTarget struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// ConnectToPage connects to the selected page in Chrome and returns a context
// for interacting with it. The browser connection must be established via a
// temporary blank tab first so that Chrome's target discovery is enabled —
// WithTargetID only works after setDiscoverTargets has been called.
func ConnectToPage(ctx context.Context, port int) (context.Context, context.CancelFunc, error) {
	allocCtx, allocCancel, err := Connect(ctx, port)
	if err != nil {
		return nil, nil, err
	}

	// Create a temporary blank tab to establish the browser connection.
	// This triggers Target.setDiscoverTargets, which is required before
	// WithTargetID can find existing tabs.
	blankCtx, blankCancel := chromedp.NewContext(allocCtx)
	if err := chromedp.Run(blankCtx); err != nil {
		blankCancel()
		allocCancel()
		return nil, nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	st, _ := state.Load()

	selectedTarget, err := resolveTargetCDP(blankCtx, st.SelectedPage)
	if err != nil {
		blankCancel()
		allocCancel()
		return nil, nil, err
	}

	pageCtx, pageCancel := chromedp.NewContext(allocCtx, chromedp.WithTargetID(selectedTarget))
	cancelAll := func() {
		pageCancel()
		blankCancel()
		allocCancel()
	}

	return pageCtx, cancelAll, nil
}

func resolveTargetCDP(ctx context.Context, selectedPage string) (target.ID, error) {
	cdpTargets, err := chromedp.Targets(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list targets: %w", err)
	}

	var targets []httpTarget
	for _, t := range cdpTargets {
		targets = append(targets, httpTarget{
			ID:    string(t.TargetID),
			Type:  string(t.Type),
			Title: t.Title,
			URL:   t.URL,
		})
	}
	return pickTarget(targets, selectedPage)
}

// ListTargets returns all page targets. Used by the pages command.
func ListTargets(port int) ([]httpTarget, error) {
	targets, err := listTargetsHTTP(port)
	if err != nil {
		return nil, err
	}
	return targets, nil
}

func listTargetsHTTP(port int) ([]httpTarget, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/json/list", port)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var targets []httpTarget
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no targets from HTTP")
	}
	return targets, nil
}

func pickTarget(targets []httpTarget, selectedPage string) (target.ID, error) {
	var selected target.ID
	var firstPage target.ID
	for _, t := range targets {
		if t.Type != "page" {
			continue
		}
		if t.URL == "about:blank" && t.Title == "" {
			continue
		}
		if firstPage == "" {
			firstPage = target.ID(t.ID)
		}
		if selectedPage != "" && strings.HasPrefix(strings.ToUpper(t.ID), strings.ToUpper(selectedPage)) {
			selected = target.ID(t.ID)
			break
		}
	}

	if selected != "" {
		return selected, nil
	}
	if firstPage != "" {
		return firstPage, nil
	}
	return "", fmt.Errorf("no pages open in Chrome")
}
