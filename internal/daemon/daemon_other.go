//go:build !linux

package daemon

import (
	"fmt"
	"os/exec"
	"runtime"
)

func setDetached(cmd *exec.Cmd) error {
	return fmt.Errorf("daemon: unsupported on %s (linux only)", runtime.GOOS)
}
