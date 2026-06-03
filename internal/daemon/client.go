package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func SockPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cdp", "daemon.sock")
}

func PidPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cdp", "daemon.pid")
}

func IsRunning() bool {
	conn, err := net.DialTimeout("unix", SockPath(), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func Call(req Request) (*Response, error) {
	if !IsRunning() {
		if err := ensureDaemon(req.Port); err != nil {
			return nil, err
		}
	}

	conn, err := net.DialTimeout("unix", SockPath(), 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to daemon: %w", err)
	}
	defer conn.Close()

	timeout := time.Duration(req.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	conn.SetDeadline(time.Now().Add(timeout + 5*time.Second))

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	conn.(*net.UnixConn).CloseWrite()

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return &resp, nil
}

func ensureDaemon(port int) error {
	// Clean stale socket
	if _, err := os.Stat(SockPath()); err == nil {
		os.Remove(SockPath())
	}

	// Clean stale PID file if process is dead
	if pidData, err := os.ReadFile(PidPath()); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(pidData))); err == nil {
			if proc, err := os.FindProcess(pid); err == nil {
				if proc.Signal(syscall.Signal(0)) != nil {
					os.Remove(PidPath())
				}
			}
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable: %w", err)
	}

	cmd := exec.Command(exe, "--daemon-serve",
		"--port", strconv.Itoa(port))
	setSysProcAttr(cmd)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}
	cmd.Process.Release()

	// Wait for daemon to be ready (up to 15s to allow for Chrome reconnection retries)
	for i := 0; i < 150; i++ {
		time.Sleep(100 * time.Millisecond)
		if IsRunning() {
			return nil
		}
	}
	return fmt.Errorf("daemon did not start within 15 seconds — check ~/.cdp/daemon.log for details")
}
