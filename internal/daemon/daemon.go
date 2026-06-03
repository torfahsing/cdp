package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"

	"github.com/user/cdp/internal/chrome"
	"github.com/user/cdp/internal/state"
)

type Daemon struct {
	mu   sync.Mutex
	port int

	allocCtx    context.Context
	allocCancel context.CancelFunc
	blankCtx    context.Context
	blankCancel context.CancelFunc

	pageCtx      context.Context
	pageCancel   context.CancelFunc
	selectedPage string

	uids map[string]state.UIDMapping

	listener  net.Listener
	idleTimer *time.Timer

	rootCancel context.CancelFunc
}

func New(port int) *Daemon {
	return &Daemon{
		port: port,
		uids: make(map[string]state.UIDMapping),
	}
}

func (d *Daemon) Serve() error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".cdp")
	os.MkdirAll(dir, 0o755)

	// Write PID
	os.WriteFile(PidPath(), []byte(strconv.Itoa(os.Getpid())), 0o644)

	// Setup logging
	logFile, err := os.Create(filepath.Join(dir, "daemon.log"))
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}
	log.Printf("daemon starting on port %d, pid %d", d.port, os.Getpid())

	// Connect to Chrome
	if err := d.connectChrome(); err != nil {
		log.Printf("failed to connect to Chrome: %v", err)
		os.Remove(PidPath())
		return err
	}
	log.Printf("connected to Chrome")

	// Load persisted state for selected page
	st, _ := state.Load()
	d.selectedPage = st.SelectedPage

	if d.selectedPage == "" {
		d.pageCtx = d.blankCtx
		d.pageCancel = d.blankCancel
	}

	// Listen on Unix socket
	sockPath := SockPath()
	os.Remove(sockPath)
	d.listener, err = net.Listen("unix", sockPath)
	if err != nil {
		d.cleanup()
		return fmt.Errorf("listen: %w", err)
	}
	log.Printf("listening on %s", sockPath)

	// Idle timer — 30 min
	d.idleTimer = time.AfterFunc(30*time.Minute, func() {
		log.Printf("idle timeout, shutting down")
		d.shutdown()
	})

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Printf("signal received, shutting down")
		d.shutdown()
	}()

	// Accept loop
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed") {
				return nil
			}
			log.Printf("accept error: %v", err)
			continue
		}
		go d.handleConn(conn)
	}
}

func (d *Daemon) connectChrome() error {
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * 500 * time.Millisecond
			log.Printf("retrying Chrome connection in %v (attempt %d/5)", delay, attempt+1)
			time.Sleep(delay)
		}

		ctx, cancel := context.WithCancel(context.Background())
		allocCtx, allocCancel, err := chrome.Connect(ctx, d.port)
		if err != nil {
			cancel()
			lastErr = err
			continue
		}

		blankCtx, blankCancel := chromedp.NewContext(allocCtx)
		if err := chromedp.Run(blankCtx); err != nil {
			blankCancel()
			allocCancel()
			cancel()
			lastErr = fmt.Errorf("browser connection: %w", err)
			continue
		}

		d.rootCancel = cancel
		d.allocCtx = allocCtx
		d.allocCancel = allocCancel
		d.blankCtx = blankCtx
		d.blankCancel = blankCancel
		return nil
	}
	return lastErr
}

func (d *Daemon) ensurePage() error {
	if d.pageCtx != nil {
		// Check if current page is still alive
		var title string
		if err := chromedp.Run(d.pageCtx, chromedp.Title(&title)); err == nil {
			return nil
		}
		if d.pageCtx != d.blankCtx {
			d.pageCancel()
		}
		d.pageCtx = nil
	}

	// Resolve target
	targets, err := chromedp.Targets(d.blankCtx)
	if err != nil {
		return fmt.Errorf("list targets: %w", err)
	}

	var selected target.ID
	var firstPage target.ID
	for _, t := range targets {
		if t.Type != "page" || (t.URL == "about:blank" && t.Title == "") {
			continue
		}
		if firstPage == "" {
			firstPage = t.TargetID
		}
		if d.selectedPage != "" && strings.HasPrefix(
			strings.ToUpper(string(t.TargetID)),
			strings.ToUpper(d.selectedPage)) {
			selected = t.TargetID
			break
		}
	}

	targetID := selected
	if targetID == "" {
		targetID = firstPage
	}
	if targetID == "" {
		return fmt.Errorf("no pages open in Chrome")
	}

	return d.attachToTarget(targetID)
}

func (d *Daemon) attachToTarget(targetID target.ID) error {
	pageCtx, pageCancel := chromedp.NewContext(d.allocCtx, chromedp.WithTargetID(targetID))
	if err := chromedp.Run(pageCtx); err != nil {
		pageCancel()
		return fmt.Errorf("attach to target %s: %w", targetID, err)
	}
	d.pageCtx = pageCtx
	d.pageCancel = pageCancel
	d.selectedPage = string(targetID)
	return nil
}

func (d *Daemon) selectPage(pageID string, focus bool) error {
	if d.pageCancel != nil {
		if d.pageCtx != d.blankCtx {
			d.pageCancel()
		}
		d.pageCtx = nil
	}

	d.selectedPage = pageID

	// Persist to state file for pages command display
	st, _ := state.Load()
	st.SelectedPage = pageID
	st.Save()

	if focus {
		targets, _ := chromedp.Targets(d.blankCtx)
		for _, t := range targets {
			if strings.HasPrefix(strings.ToUpper(string(t.TargetID)), strings.ToUpper(pageID)) {
				chromedp.Run(d.blankCtx, chromedp.ActionFunc(func(ctx context.Context) error {
					return target.ActivateTarget(t.TargetID).Do(ctx)
				}))
				break
			}
		}
	}

	return d.ensurePage()
}

func (d *Daemon) handleConn(conn net.Conn) {
	defer conn.Close()

	var req Request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		log.Printf("decode error: %v", err)
		return
	}

	d.idleTimer.Reset(30 * time.Minute)

	d.mu.Lock()
	resp := wrapResponse(d.dispatch(req))
	d.mu.Unlock()

	json.NewEncoder(conn).Encode(resp)
}

func (d *Daemon) shutdown() {
	log.Printf("shutting down")
	if d.pageCancel != nil {
		d.pageCancel()
	}
	if d.blankCancel != nil {
		d.blankCancel()
	}
	if d.allocCancel != nil {
		d.allocCancel()
	}
	if d.rootCancel != nil {
		d.rootCancel()
	}
	if d.listener != nil {
		d.listener.Close()
	}
	os.Remove(SockPath())
	os.Remove(PidPath())
	os.Exit(0)
}

func (d *Daemon) cleanup() {
	if d.allocCancel != nil {
		d.allocCancel()
	}
	if d.rootCancel != nil {
		d.rootCancel()
	}
}
