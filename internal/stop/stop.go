package stop

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/monms/monms/internal/cli"
)

const shutdownGrace = 2 * time.Second

// StopAll sends SIGTERM (then SIGKILL) to all monms processes using this binary,
// excluding the current process. Returns stopped PIDs.
func StopAll() ([]int, error) {
	exe, err := resolveExecutable()
	if err != nil {
		return nil, err
	}

	pids, err := findMatchingPIDs(exe, os.Getpid())
	if err != nil {
		return nil, err
	}
	if len(pids) == 0 {
		return nil, nil
	}

	targets := append([]int(nil), pids...)

	for _, pid := range targets {
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			slog.Warn("stop: sigterm failed", "pid", pid, "err", err)
		}
	}

	deadline := time.Now().Add(shutdownGrace)
	for time.Now().Before(deadline) {
		if len(alivePIDs(targets)) == 0 {
			return targets, nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	for _, pid := range alivePIDs(targets) {
		if err := syscall.Kill(pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
			slog.Warn("stop: sigkill failed", "pid", pid, "err", err)
		}
	}

	return targets, nil
}

// RunCLI stops all running processes that share the invoked monms binary path.
func RunCLI(args []string) error {
	if _, wantHelp := cli.ParseHelpRequest(append([]string{"stop"}, args...)); wantHelp {
		cli.PrintHelp("stop")
		return nil
	}
	if len(args) > 0 {
		cli.PrintHelp("stop")
		return fmt.Errorf("unknown arguments: %s", strings.Join(args, " "))
	}

	targets, err := StopAll()
	if err != nil {
		return fmt.Errorf("stop: %w", err)
	}
	if len(targets) == 0 {
		fmt.Println("no running monms instances")
		return nil
	}

	fmt.Printf("stopped %d instance(s): %v\n", len(targets), targets)
	return nil
}

func resolveExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	return filepath.Abs(exe)
}

func normalizeProcExe(link string) string {
	if idx := strings.Index(link, " (deleted)"); idx >= 0 {
		return link[:idx]
	}
	return link
}

func sameExecutable(a, b string) bool {
	a = normalizeProcExe(a)
	b = normalizeProcExe(b)

	aAbs, errA := filepath.Abs(a)
	bAbs, errB := filepath.Abs(b)
	if errA == nil {
		a = aAbs
	}
	if errB == nil {
		b = bAbs
	}

	aEval, errA := filepath.EvalSymlinks(a)
	bEval, errB := filepath.EvalSymlinks(b)
	if errA == nil {
		a = aEval
	}
	if errB == nil {
		b = bEval
	}

	return a == b
}

func alivePIDs(pids []int) []int {
	var out []int
	for _, pid := range pids {
		if err := syscall.Kill(pid, 0); err == nil {
			out = append(out, pid)
		}
	}
	return out
}
