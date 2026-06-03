//go:build !windows

package daemon

import (
	"os/exec"
	"syscall"
)

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func KillProcess(pid int) {
	syscall.Kill(pid, syscall.SIGTERM)
}
