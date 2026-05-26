//go:build linux

package daemon

import (
	"os/exec"
	"syscall"
)

func setDetached(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return nil
}
