//go:build !linux

package stop

import (
	"fmt"
	"runtime"
)

func findMatchingPIDs(exePath string, selfPID int) ([]int, error) {
	return nil, fmt.Errorf("unsupported on %s (linux only)", runtime.GOOS)
}
