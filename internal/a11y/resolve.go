package a11y

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	"github.com/user/cdp/internal/state"
)

func ResolveUID(ctx context.Context, uid string) (cdp.BackendNodeID, error) {
	st, err := state.Load()
	if err != nil {
		return 0, fmt.Errorf("failed to load state: %w", err)
	}

	mapping, ok := st.UIDs[uid]
	if !ok {
		return 0, fmt.Errorf("unknown UID %q — run 'cdp snapshot' first", uid)
	}

	return cdp.BackendNodeID(mapping.BackendNodeID), nil
}

func ResolveUIDFromMap(uids map[string]state.UIDMapping, uid string) (cdp.BackendNodeID, error) {
	mapping, ok := uids[uid]
	if !ok {
		return 0, fmt.Errorf("unknown UID %q — run 'cdp snapshot' first", uid)
	}
	return cdp.BackendNodeID(mapping.BackendNodeID), nil
}

func ClickByUID(ctx context.Context, uid string, dblClick bool) error {
	nodeID, err := ResolveUID(ctx, uid)
	if err != nil {
		return err
	}

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		obj, err := dom.ResolveNode().WithBackendNodeID(nodeID).Do(ctx)
		if err != nil {
			return fmt.Errorf("cannot resolve node for UID %s: %w", uid, err)
		}

		js := `function() { this.scrollIntoViewIfNeeded(true); }`
		if dblClick {
			js = `function() { this.scrollIntoViewIfNeeded(true); this.dispatchEvent(new MouseEvent('dblclick', {bubbles: true})); }`
		}

		_, _, err = runtime.CallFunctionOn(js).
			WithObjectID(obj.ObjectID).
			Do(ctx)
		if err != nil {
			return err
		}

		if !dblClick {
			_, _, err = runtime.CallFunctionOn(`function() { this.click(); }`).
				WithObjectID(obj.ObjectID).
				Do(ctx)
		}
		return err
	}))
}

func HoverByUID(ctx context.Context, uid string) error {
	nodeID, err := ResolveUID(ctx, uid)
	if err != nil {
		return err
	}

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		obj, err := dom.ResolveNode().WithBackendNodeID(nodeID).Do(ctx)
		if err != nil {
			return fmt.Errorf("cannot resolve node for UID %s: %w", uid, err)
		}
		_, _, err = runtime.CallFunctionOn(`function() {
			this.scrollIntoViewIfNeeded(true);
			this.dispatchEvent(new MouseEvent('mouseover', {bubbles: true}));
			this.dispatchEvent(new MouseEvent('mouseenter', {bubbles: true}));
		}`).WithObjectID(obj.ObjectID).Do(ctx)
		return err
	}))
}
