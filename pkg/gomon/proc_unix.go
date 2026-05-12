//go:build !windows

package gomon

import (
	"os/exec"
	"syscall"
	"time"
)

func setProcGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

func killProcessTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	time.Sleep(80 * time.Millisecond)
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}
