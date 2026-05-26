//go:build linux

package stop

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func findMatchingPIDs(exePath string, selfPID int) ([]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}

	var pids []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 || pid == selfPID {
			continue
		}

		link, err := os.Readlink(filepath.Join("/proc", entry.Name(), "exe"))
		if err != nil {
			continue
		}
		if sameExecutable(link, exePath) {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}
