package daemon

import (
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/heapprofiler"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/cdproto/tracing"
	"github.com/chromedp/chromedp"

	"github.com/user/cdp/internal/a11y"
)

func (d *Daemon) dispatch(req Request) *Response {
	ctx := d.pageCtx
	timeout := time.Duration(req.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	switch req.Command {
	case "pages":
		return d.handlePages()
	case "select":
		var args SelectArgs
		json.Unmarshal(req.Args, &args)
		return d.handleSelect(args)
	case "close":
		var args ClosePageArgs
		json.Unmarshal(req.Args, &args)
		return d.handleClose(args)
	case "open":
		var args OpenPageArgs
		json.Unmarshal(req.Args, &args)
		return d.handleOpen(args)
	case "shutdown":
		go func() {
			time.Sleep(100 * time.Millisecond)
			d.shutdown()
		}()
		return &Response{OK: true, Text: "Shutting down\n"}
	case "status":
		return &Response{OK: true, Text: "Daemon running\n"}
	}

	if err := d.ensurePage(); err != nil {
		return &Response{Error: formatTargetError(err)}
	}
	ctx = d.pageCtx

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	switch req.Command {
	case "snapshot":
		var args SnapshotArgs
		json.Unmarshal(req.Args, &args)
		return d.handleSnapshot(ctx, args)
	case "click":
		var args ClickArgs
		json.Unmarshal(req.Args, &args)
		return d.handleClick(ctx, args)
	case "hover":
		var args HoverArgs
		json.Unmarshal(req.Args, &args)
		return d.handleHover(ctx, args)
	case "drag":
		var args DragArgs
		json.Unmarshal(req.Args, &args)
		return d.handleDrag(ctx, args)
	case "fill":
		var args FillArgs
		json.Unmarshal(req.Args, &args)
		return d.handleFill(ctx, args)
	case "fill-form":
		var args FillFormArgs
		json.Unmarshal(req.Args, &args)
		return d.handleFillForm(ctx, args)
	case "type":
		var args TypeArgs
		json.Unmarshal(req.Args, &args)
		return d.handleType(ctx, args)
	case "key":
		var args KeyArgs
		json.Unmarshal(req.Args, &args)
		return d.handleKey(ctx, args)
	case "eval":
		var args EvalArgs
		json.Unmarshal(req.Args, &args)
		return d.handleEval(ctx, args)
	case "navigate":
		var args NavigateArgs
		json.Unmarshal(req.Args, &args)
		return d.handleNavigate(ctx, args)
	case "screenshot":
		var args ScreenshotArgs
		json.Unmarshal(req.Args, &args)
		return d.handleScreenshot(ctx, args)
	case "wait":
		var args WaitArgs
		json.Unmarshal(req.Args, &args)
		return d.handleWait(ctx, args)
	case "dialog":
		var args DialogArgs
		json.Unmarshal(req.Args, &args)
		return d.handleDialog(ctx, args)
	case "resize":
		var args ResizeArgs
		json.Unmarshal(req.Args, &args)
		return d.handleResize(ctx, args)
	case "upload":
		var args UploadArgs
		json.Unmarshal(req.Args, &args)
		return d.handleUpload(ctx, args)
	case "emulate":
		var args EmulateArgs
		json.Unmarshal(req.Args, &args)
		return d.handleEmulate(ctx, args)
	case "console":
		var args ConsoleArgs
		json.Unmarshal(req.Args, &args)
		return d.handleConsole(ctx, args)
	case "network":
		return d.handleNetwork(ctx)
	case "perf-start":
		var args PerfStartArgs
		json.Unmarshal(req.Args, &args)
		return d.handlePerfStart(ctx, args)
	case "perf-stop":
		var args PerfStopArgs
		json.Unmarshal(req.Args, &args)
		return d.handlePerfStop(ctx, args)
	case "perf-insight":
		var args PerfInsightArgs
		json.Unmarshal(req.Args, &args)
		return d.handlePerfInsight(args)
	case "memory":
		var args MemoryArgs
		json.Unmarshal(req.Args, &args)
		return d.handleMemory(ctx, args)
	default:
		return &Response{Error: fmt.Sprintf("unknown command: %s", req.Command)}
	}
}

func (d *Daemon) handlePages() *Response {
	targets, err := chromedp.Targets(d.blankCtx)
	if err != nil {
		return &Response{Error: err.Error()}
	}

	type pageInfo struct {
		ID       string `json:"id"`
		URL      string `json:"url"`
		Title    string `json:"title"`
		Selected bool   `json:"selected"`
	}

	var pages []pageInfo
	var textLines []string
	for _, t := range targets {
		if t.Type != "page" {
			continue
		}
		id := string(t.TargetID)
		selected := strings.HasPrefix(
			strings.ToUpper(id),
			strings.ToUpper(d.selectedPage))
		pages = append(pages, pageInfo{
			ID:       id,
			URL:      t.URL,
			Title:    t.Title,
			Selected: selected,
		})
		marker := "  "
		if selected {
			marker = "* "
		}
		short := id
		if len(short) > 8 {
			short = short[:8]
		}
		textLines = append(textLines, fmt.Sprintf("%s%s %s \"%s\"", marker, short, t.URL, t.Title))
	}

	if len(pages) == 0 {
		return &Response{OK: true, Text: "No pages open\n", Data: mustJSON([]pageInfo{})}
	}

	return &Response{OK: true, Text: strings.Join(textLines, "\n") + "\n", Data: mustJSON(pages)}
}

func (d *Daemon) handleSelect(args SelectArgs) *Response {
	if err := d.selectPage(args.PageID, args.Focus); err != nil {
		return &Response{Error: err.Error()}
	}
	return &Response{OK: true,
		Text: fmt.Sprintf("Selected page %s\n", args.PageID),
		Data: mustJSON(map[string]string{"selected": args.PageID}),
	}
}

func (d *Daemon) handleClose(args ClosePageArgs) *Response {
	if err := chromedp.Run(d.blankCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		return target.CloseTarget(target.ID(args.PageID)).Do(ctx)
	})); err != nil {
		return &Response{Error: err.Error()}
	}
	return &Response{OK: true,
		Text: fmt.Sprintf("Closed page %s\n", args.PageID),
		Data: mustJSON(map[string]string{"closed": args.PageID}),
	}
}

func (d *Daemon) handleOpen(args OpenPageArgs) *Response {
	var targetID target.ID
	if err := chromedp.Run(d.blankCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		t, err := target.CreateTarget(args.URL).Do(ctx)
		if err != nil {
			return err
		}
		targetID = t
		if !args.Background {
			return target.ActivateTarget(t).Do(ctx)
		}
		return nil
	})); err != nil {
		return &Response{Error: err.Error()}
	}

	if !args.Background {
		d.selectPage(string(targetID), false)
	}

	short := string(targetID)
	if len(short) > 8 {
		short = short[:8]
	}
	return &Response{OK: true,
		Text: fmt.Sprintf("Opened %s (%s)\n", args.URL, short),
		Data: mustJSON(map[string]string{"id": string(targetID), "url": args.URL}),
	}
}

func (d *Daemon) handleSnapshot(ctx context.Context, args SnapshotArgs) *Response {
	elements, uidMap, err := a11y.FetchTree(ctx, args.Verbose)
	if err != nil {
		return &Response{Error: err.Error()}
	}
	d.uids = uidMap
	text := a11y.FormatText(elements)
	return &Response{OK: true, Text: text, Data: mustJSON(elements)}
}

func (d *Daemon) handleClick(ctx context.Context, args ClickArgs) *Response {
	nodeID, err := a11y.ResolveUIDFromMap(d.uids, args.UID)
	if err != nil {
		return &Response{Error: err.Error()}
	}

	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		obj, err := dom.ResolveNode().WithBackendNodeID(nodeID).Do(ctx)
		if err != nil {
			return fmt.Errorf("cannot resolve node for UID %s: %w", args.UID, err)
		}

		js := `function() { this.scrollIntoViewIfNeeded(true); }`
		if args.DblClick {
			js = `function() { this.scrollIntoViewIfNeeded(true); this.dispatchEvent(new MouseEvent('dblclick', {bubbles: true})); }`
		}

		_, _, err = runtime.CallFunctionOn(js).WithObjectID(obj.ObjectID).Do(ctx)
		if err != nil {
			return err
		}

		if !args.DblClick {
			_, _, err = runtime.CallFunctionOn(`function() { this.click(); }`).WithObjectID(obj.ObjectID).Do(ctx)
		}
		return err
	})); err != nil {
		return &Response{Error: err.Error()}
	}

	action := "Clicked"
	if args.DblClick {
		action = "Double-clicked"
	}

	result := map[string]any{"action": action, "uid": args.UID}

	if args.Snapshot {
		elements, uidMap, err := a11y.FetchTree(ctx, false)
		if err == nil {
			d.uids = uidMap
			result["snapshot"] = elements
		}
	}

	return &Response{OK: true,
		Text: fmt.Sprintf("%s %s\n", action, args.UID),
		Data: mustJSON(result),
	}
}

func (d *Daemon) handleHover(ctx context.Context, args HoverArgs) *Response {
	nodeID, err := a11y.ResolveUIDFromMap(d.uids, args.UID)
	if err != nil {
		return &Response{Error: err.Error()}
	}

	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		obj, err := dom.ResolveNode().WithBackendNodeID(nodeID).Do(ctx)
		if err != nil {
			return fmt.Errorf("cannot resolve node for UID %s: %w", args.UID, err)
		}
		_, _, err = runtime.CallFunctionOn(`function() {
			this.scrollIntoViewIfNeeded(true);
			this.dispatchEvent(new MouseEvent('mouseover', {bubbles: true}));
			this.dispatchEvent(new MouseEvent('mouseenter', {bubbles: true}));
		}`).WithObjectID(obj.ObjectID).Do(ctx)
		return err
	})); err != nil {
		return &Response{Error: err.Error()}
	}

	return &Response{OK: true,
		Text: fmt.Sprintf("Hovering %s\n", args.UID),
		Data: mustJSON(map[string]string{"uid": args.UID}),
	}
}

func (d *Daemon) handleDrag(ctx context.Context, args DragArgs) *Response {
	fromNodeID, err := a11y.ResolveUIDFromMap(d.uids, args.FromUID)
	if err != nil {
		return &Response{Error: err.Error()}
	}
	toNodeID, err := a11y.ResolveUIDFromMap(d.uids, args.ToUID)
	if err != nil {
		return &Response{Error: err.Error()}
	}

	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		fromObj, err := dom.ResolveNode().WithBackendNodeID(fromNodeID).Do(ctx)
		if err != nil {
			return err
		}
		toObj, err := dom.ResolveNode().WithBackendNodeID(toNodeID).Do(ctx)
		if err != nil {
			return err
		}

		js := `function() {
			this.scrollIntoViewIfNeeded(true);
			this.dispatchEvent(new DragEvent('dragstart', {bubbles: true}));
		}`
		if _, _, err := runtime.CallFunctionOn(js).WithObjectID(fromObj.ObjectID).Do(ctx); err != nil {
			return err
		}

		js = `function() {
			this.scrollIntoViewIfNeeded(true);
			this.dispatchEvent(new DragEvent('drop', {bubbles: true}));
			this.dispatchEvent(new DragEvent('dragend', {bubbles: true}));
		}`
		_, _, err = runtime.CallFunctionOn(js).WithObjectID(toObj.ObjectID).Do(ctx)
		return err
	})); err != nil {
		return &Response{Error: err.Error()}
	}

	return &Response{OK: true,
		Text: fmt.Sprintf("Dragged %s → %s\n", args.FromUID, args.ToUID),
		Data: mustJSON(map[string]string{"from": args.FromUID, "to": args.ToUID}),
	}
}

func (d *Daemon) handleFill(ctx context.Context, args FillArgs) *Response {
	nodeID, err := a11y.ResolveUIDFromMap(d.uids, args.UID)
	if err != nil {
		return &Response{Error: err.Error()}
	}

	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		obj, err := dom.ResolveNode().WithBackendNodeID(cdp.BackendNodeID(nodeID)).Do(ctx)
		if err != nil {
			return err
		}
		js := fmt.Sprintf(`function() {
			this.focus();
			this.value = %q;
			this.dispatchEvent(new Event('input', {bubbles: true}));
			this.dispatchEvent(new Event('change', {bubbles: true}));
		}`, args.Value)
		_, _, err = runtime.CallFunctionOn(js).WithObjectID(obj.ObjectID).Do(ctx)
		return err
	})); err != nil {
		return &Response{Error: err.Error()}
	}

	return &Response{OK: true,
		Text: fmt.Sprintf("Filled %s with %q\n", args.UID, args.Value),
		Data: mustJSON(map[string]string{"uid": args.UID, "value": args.Value}),
	}
}

func (d *Daemon) handleFillForm(ctx context.Context, args FillFormArgs) *Response {
	var filled []map[string]string
	for uid, value := range args.Fields {
		nodeID, err := a11y.ResolveUIDFromMap(d.uids, uid)
		if err != nil {
			return &Response{Error: err.Error()}
		}
		if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			obj, err := dom.ResolveNode().WithBackendNodeID(cdp.BackendNodeID(nodeID)).Do(ctx)
			if err != nil {
				return err
			}
			js := fmt.Sprintf(`function() {
				this.focus();
				this.value = %q;
				this.dispatchEvent(new Event('input', {bubbles: true}));
				this.dispatchEvent(new Event('change', {bubbles: true}));
			}`, value)
			_, _, err = runtime.CallFunctionOn(js).WithObjectID(obj.ObjectID).Do(ctx)
			return err
		})); err != nil {
			return &Response{Error: err.Error()}
		}
		filled = append(filled, map[string]string{"uid": uid, "value": value})
	}

	var textLines []string
	for _, f := range filled {
		textLines = append(textLines, fmt.Sprintf("Filled %s with %q", f["uid"], f["value"]))
	}

	return &Response{OK: true,
		Text: strings.Join(textLines, "\n") + "\n",
		Data: mustJSON(filled),
	}
}

func (d *Daemon) handleType(ctx context.Context, args TypeArgs) *Response {
	for _, ch := range args.Text {
		if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return input.DispatchKeyEvent(input.KeyDown).
				WithText(string(ch)).
				Do(ctx)
		})); err != nil {
			return &Response{Error: err.Error()}
		}
	}

	if args.SubmitKey != "" {
		if err := chromedp.Run(ctx, chromedp.KeyEvent(args.SubmitKey)); err != nil {
			return &Response{Error: err.Error()}
		}
	}

	return &Response{OK: true,
		Text: fmt.Sprintf("Typed %q\n", args.Text),
		Data: mustJSON(map[string]string{"text": args.Text}),
	}
}

func (d *Daemon) handleKey(ctx context.Context, args KeyArgs) *Response {
	if err := chromedp.Run(ctx, chromedp.KeyEvent(args.Key)); err != nil {
		return &Response{Error: err.Error()}
	}
	return &Response{OK: true,
		Text: fmt.Sprintf("Pressed %s\n", args.Key),
		Data: mustJSON(map[string]string{"key": args.Key}),
	}
}

func (d *Daemon) handleEval(ctx context.Context, args EvalArgs) *Response {
	var result json.RawMessage
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		v, exc, err := runtime.Evaluate(args.Script).
			WithReturnByValue(true).
			WithAwaitPromise(true).
			Do(ctx)
		if err != nil {
			return err
		}
		if exc != nil {
			return fmt.Errorf("JS exception: %s", exc.Text)
		}
		if v != nil && len(v.Value) > 0 {
			result = json.RawMessage(v.Value)
		} else {
			result = json.RawMessage(`null`)
		}
		return nil
	})); err != nil {
		return &Response{Error: err.Error()}
	}
	return &Response{OK: true,
		Text: string(result) + "\n",
		Data: mustJSON(map[string]any{"result": result}),
	}
}

func (d *Daemon) handleNavigate(ctx context.Context, args NavigateArgs) *Response {
	var err error
	switch {
	case args.Back:
		err = chromedp.Run(ctx, chromedp.NavigateBack())
	case args.Forward:
		err = chromedp.Run(ctx, chromedp.NavigateForward())
	case args.Reload:
		err = chromedp.Run(ctx, chromedp.Reload())
	case args.URL != "":
		err = chromedp.Run(ctx, chromedp.Navigate(args.URL))
	default:
		return &Response{Error: "provide a URL or use back/forward/reload"}
	}
	if err != nil {
		return &Response{Error: err.Error()}
	}

	if args.WaitLoad {
		chromedp.Run(ctx, chromedp.WaitReady("body", chromedp.ByQuery))
	}

	var title, currentURL string
	chromedp.Run(ctx, chromedp.Title(&title), chromedp.Location(&currentURL))

	return &Response{OK: true,
		Text: fmt.Sprintf("Navigated to %s \"%s\"\n", currentURL, title),
		Data: mustJSON(map[string]string{"url": currentURL, "title": title}),
	}
}

func (d *Daemon) handleScreenshot(ctx context.Context, args ScreenshotArgs) *Response {
	var buf []byte
	var err error

	var captureFormat page.CaptureScreenshotFormat
	switch args.Format {
	case "jpeg":
		captureFormat = page.CaptureScreenshotFormatJpeg
	case "webp":
		captureFormat = page.CaptureScreenshotFormatWebp
	default:
		captureFormat = page.CaptureScreenshotFormatPng
	}

	quality := args.Quality
	if quality == 0 {
		quality = 80
	}

	if args.UID != "" {
		nodeID, uidErr := a11y.ResolveUIDFromMap(d.uids, args.UID)
		if uidErr != nil {
			return &Response{Error: uidErr.Error()}
		}
		err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			obj, err := dom.ResolveNode().WithBackendNodeID(nodeID).Do(ctx)
			if err != nil {
				return err
			}
			_, _, err = runtime.CallFunctionOn(`function() { this.scrollIntoViewIfNeeded(true); }`).
				WithObjectID(obj.ObjectID).Do(ctx)
			if err != nil {
				return err
			}
			buf, err = page.CaptureScreenshot().
				WithFormat(captureFormat).
				WithQuality(int64(quality)).
				Do(ctx)
			return err
		}))
	} else if args.FullPage {
		err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			// captureBeyondViewport avoids SetDeviceMetricsOverride, which fires resize
			// events that React/nuqs picks up and uses to reset URL state.
			_, _, _, _, _, cssContentSize, layoutErr := page.GetLayoutMetrics().Do(ctx)
			if layoutErr != nil {
				return layoutErr
			}
			buf, err = page.CaptureScreenshot().
				WithFormat(captureFormat).
				WithQuality(int64(quality)).
				WithClip(&page.Viewport{
					X:      0,
					Y:      0,
					Width:  cssContentSize.Width,
					Height: cssContentSize.Height,
					Scale:  1,
				}).
				WithCaptureBeyondViewport(true).
				Do(ctx)
			return err
		}))
	} else {
		err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			buf, err = page.CaptureScreenshot().
				WithFormat(captureFormat).
				WithQuality(int64(quality)).
				Do(ctx)
			return err
		}))
	}
	if err != nil {
		return &Response{Error: err.Error()}
	}

	encoded := base64.StdEncoding.EncodeToString(buf)
	return &Response{OK: true,
		Text:   fmt.Sprintf("Screenshot captured (%dKB)\n", len(buf)/1024),
		Data:   mustJSON(map[string]any{"format": args.Format, "size": len(buf)}),
		Binary: encoded,
	}
}

func (d *Daemon) handleWait(ctx context.Context, args WaitArgs) *Response {
	for _, text := range args.Texts {
		if err := chromedp.Run(ctx,
			chromedp.WaitVisible(fmt.Sprintf(`//*[contains(text(), %q)]`, text), chromedp.BySearch),
		); err == nil {
			return &Response{OK: true,
				Text: fmt.Sprintf("Found %q\n", text),
				Data: mustJSON(map[string]string{"found": text}),
			}
		}
	}
	return &Response{Error: "timeout waiting for text"}
}

func (d *Daemon) handleDialog(ctx context.Context, args DialogArgs) *Response {
	accept := args.Action == "accept"
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return page.HandleJavaScriptDialog(accept).
			WithPromptText(args.Text).
			Do(ctx)
	})); err != nil {
		return &Response{Error: err.Error()}
	}
	return &Response{OK: true,
		Text: fmt.Sprintf("Dialog %sed\n", args.Action),
		Data: mustJSON(map[string]string{"action": args.Action}),
	}
}

func (d *Daemon) handleResize(ctx context.Context, args ResizeArgs) *Response {
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return emulation.SetDeviceMetricsOverride(int64(args.Width), int64(args.Height), 1.0, false).Do(ctx)
	})); err != nil {
		return &Response{Error: err.Error()}
	}
	return &Response{OK: true,
		Text: fmt.Sprintf("Resized to %dx%d\n", args.Width, args.Height),
		Data: mustJSON(map[string]any{"width": args.Width, "height": args.Height}),
	}
}

func (d *Daemon) handleUpload(ctx context.Context, args UploadArgs) *Response {
	nodeID, err := a11y.ResolveUIDFromMap(d.uids, args.UID)
	if err != nil {
		return &Response{Error: err.Error()}
	}

	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		node, err := dom.DescribeNode().WithBackendNodeID(cdp.BackendNodeID(nodeID)).Do(ctx)
		if err != nil {
			return err
		}
		return dom.SetFileInputFiles([]string{args.FilePath}).WithNodeID(node.NodeID).Do(ctx)
	})); err != nil {
		return &Response{Error: err.Error()}
	}

	return &Response{OK: true,
		Text: fmt.Sprintf("Uploaded %s via %s\n", args.FilePath, args.UID),
		Data: mustJSON(map[string]string{"uid": args.UID, "file": args.FilePath}),
	}
}

func (d *Daemon) handleEmulate(ctx context.Context, args EmulateArgs) *Response {
	var actions []chromedp.Action
	var applied []string

	if args.Viewport != "" {
		parts := strings.Split(args.Viewport, "x")
		if len(parts) < 2 {
			return &Response{Error: "viewport format: WxHxDPR"}
		}
		w, _ := strconv.ParseInt(parts[0], 10, 64)
		h, _ := strconv.ParseInt(parts[1], 10, 64)
		dpr := 1.0
		if len(parts) >= 3 {
			dpr, _ = strconv.ParseFloat(parts[2], 64)
		}
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetDeviceMetricsOverride(w, h, dpr, false).Do(ctx)
		}))
		applied = append(applied, fmt.Sprintf("viewport=%s", args.Viewport))
	}

	if args.UserAgent != "" {
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetUserAgentOverride(args.UserAgent).Do(ctx)
		}))
		applied = append(applied, "user-agent set")
	}

	if args.ColorScheme != "" {
		cs := args.ColorScheme
		features := []*emulation.MediaFeature{{Name: "prefers-color-scheme", Value: cs}}
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetEmulatedMedia().WithFeatures(features).Do(ctx)
		}))
		applied = append(applied, fmt.Sprintf("color-scheme=%s", cs))
	}

	if args.CPUThrottle > 1 {
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetCPUThrottlingRate(args.CPUThrottle).Do(ctx)
		}))
		applied = append(applied, fmt.Sprintf("cpu-throttle=%.1fx", args.CPUThrottle))
	}

	if args.Geo != "" {
		parts := strings.Split(args.Geo, "x")
		if len(parts) == 2 {
			lat, _ := strconv.ParseFloat(parts[0], 64)
			lng, _ := strconv.ParseFloat(parts[1], 64)
			actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
				return emulation.SetGeolocationOverride().
					WithLatitude(lat).
					WithLongitude(lng).
					WithAccuracy(1).
					Do(ctx)
			}))
			applied = append(applied, fmt.Sprintf("geo=%s", args.Geo))
		}
	}

	if len(actions) == 0 {
		return &Response{Error: "no emulation flags provided"}
	}

	for _, action := range actions {
		if err := chromedp.Run(ctx, action); err != nil {
			return &Response{Error: err.Error()}
		}
	}

	return &Response{OK: true,
		Text: fmt.Sprintf("Emulating: %s\n", strings.Join(applied, ", ")),
		Data: mustJSON(map[string]any{"applied": applied}),
	}
}

func (d *Daemon) handleConsole(ctx context.Context, args ConsoleArgs) *Response {
	var result json.RawMessage
	js := `(function() {
		if (!window.__cdp_console) return [];
		return window.__cdp_console;
	})()`
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &result)); err != nil {
		return &Response{Error: err.Error()}
	}
	return &Response{OK: true,
		Text: "Console messages require instrumentation — use 'cdp eval' with console API\n",
		Data: result,
	}
}

func (d *Daemon) handleNetwork(ctx context.Context) *Response {
	var result json.RawMessage
	js := `(function() {
		return performance.getEntriesByType('resource').map(e => ({
			name: e.name,
			type: e.initiatorType,
			duration: Math.round(e.duration),
			size: e.transferSize || 0
		}));
	})()`
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &result)); err != nil {
		return &Response{Error: err.Error()}
	}

	var entries []struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Duration int    `json:"duration"`
		Size     int    `json:"size"`
	}
	json.Unmarshal(result, &entries)

	var textLines []string
	for _, e := range entries {
		textLines = append(textLines, fmt.Sprintf("[%s] %s %dms %dB", e.Type, e.Name, e.Duration, e.Size))
	}

	text := strings.Join(textLines, "\n")
	if text != "" {
		text += "\n"
	}

	return &Response{OK: true, Text: text, Data: result}
}

func (d *Daemon) handlePerfStart(ctx context.Context, args PerfStartArgs) *Response {
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return tracing.Start().
			WithTraceConfig(&tracing.TraceConfig{
				IncludedCategories: []string{
					"devtools.timeline",
					"v8.execute",
					"disabled-by-default-devtools.timeline",
				},
			}).
			Do(ctx)
	})); err != nil {
		return &Response{Error: err.Error()}
	}

	if args.Reload {
		chromedp.Run(ctx, chromedp.Reload())
	}

	return &Response{OK: true,
		Text: "Trace started\n",
		Data: mustJSON(map[string]string{"status": "recording"}),
	}
}

func (d *Daemon) handlePerfStop(ctx context.Context, args PerfStopArgs) *Response {
	var traceData []byte
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if dc, ok := ev.(*tracing.EventDataCollected); ok {
			traceData = append(traceData, []byte(fmt.Sprintf("%v", dc.Value))...)
		}
	})

	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return tracing.End().Do(ctx)
	})); err != nil {
		return &Response{Error: err.Error()}
	}

	time.Sleep(2 * time.Second)

	if args.FilePath != "" {
		if strings.HasSuffix(args.FilePath, ".gz") {
			f, err := os.Create(args.FilePath)
			if err != nil {
				return &Response{Error: err.Error()}
			}
			defer f.Close()
			w := gzip.NewWriter(f)
			w.Write(traceData)
			w.Close()
		} else {
			os.WriteFile(args.FilePath, traceData, 0o644)
		}
		return &Response{OK: true,
			Text: fmt.Sprintf("Trace saved to %s\n", args.FilePath),
			Data: mustJSON(map[string]string{"file": args.FilePath}),
		}
	}

	return &Response{OK: true,
		Text: "Trace stopped\n",
		Data: mustJSON(map[string]string{"status": "stopped", "size": fmt.Sprintf("%d", len(traceData))}),
	}
}

func (d *Daemon) handlePerfInsight(args PerfInsightArgs) *Response {
	return &Response{OK: true,
		Text: fmt.Sprintf("Performance insight analysis for %s/%s — requires trace data\n", args.SetID, args.Name),
		Data: mustJSON(map[string]string{"setId": args.SetID, "insight": args.Name, "status": "requires-trace"}),
	}
}

func (d *Daemon) handleMemory(ctx context.Context, args MemoryArgs) *Response {
	f, err := os.Create(args.FilePath)
	if err != nil {
		return &Response{Error: err.Error()}
	}
	defer f.Close()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if chunk, ok := ev.(*heapprofiler.EventAddHeapSnapshotChunk); ok {
			f.WriteString(chunk.Chunk)
		}
	})

	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return heapprofiler.TakeHeapSnapshot().Do(ctx)
	})); err != nil {
		return &Response{Error: err.Error()}
	}

	fi, _ := f.Stat()
	return &Response{OK: true,
		Text: fmt.Sprintf("Heap snapshot saved to %s (%dKB)\n", args.FilePath, fi.Size()/1024),
		Data: mustJSON(map[string]any{"file": args.FilePath, "size": fi.Size()}),
	}
}

func formatTargetError(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "-32602") ||
		strings.Contains(msg, "No target with given id") ||
		strings.Contains(msg, "target closed") ||
		strings.Contains(msg, "not found") {
		return fmt.Sprintf("Target page lost (possibly closed or navigated away). Run 'cdp pages' to find the page and 'cdp select <id>' to re-select it. Original error: %s", msg)
	}
	return msg
}

func wrapResponse(resp *Response) *Response {
	if resp != nil && resp.Error != "" {
		resp.Error = formatTargetError(fmt.Errorf("%s", resp.Error))
	}
	return resp
}

func mustJSON(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return json.RawMessage(`null`)
	}
	return data
}

