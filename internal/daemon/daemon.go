package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/monms/monms/internal/cli"
	"github.com/monms/monms/internal/config"
	"github.com/monms/monms/internal/logging"
)

const daemonEnv = "MONMS_DAEMON"

// ServeRequested reports whether args target the serve subcommand.
func ServeRequested(args []string) bool {
	return len(args) > 0 && args[0] == "serve"
}

// StripDaemonFlag removes -d/--daemon from args and reports whether one was present.
func StripDaemonFlag(args []string) (clean []string, requested bool) {
	for _, arg := range args {
		switch arg {
		case "-d", "--daemon":
			requested = true
		default:
			clean = append(clean, arg)
		}
	}
	return clean, requested
}

// ShouldDetach reports whether the current invocation should fork into the background.
func ShouldDetach(args []string) bool {
	if os.Getenv(daemonEnv) == "1" {
		return false
	}
	clean, requested := StripDaemonFlag(args)
	if !requested {
		return false
	}
	if len(clean) == 0 {
		return true
	}
	switch clean[0] {
	case "init", "validate", "content", "stop", "restart":
		return false
	default:
		return true
	}
}

// Start forks a detached serve process and writes site runtime files.
func Start(siteAbs string, args []string) error {
	childArgs, _ := StripDaemonFlag(args)
	childArgs = config.StripSiteFlags(childArgs)
	childArgs = cli.EnsureServeSubcommand(childArgs)
	exe, err := executablePath()
	if err != nil {
		return err
	}

	monmsDir := filepath.Join(siteAbs, ".monms")
	if err := os.MkdirAll(monmsDir, 0o755); err != nil {
		return fmt.Errorf("daemon: create runtime dir: %w", err)
	}

	logDir := logging.LogDirForSite(siteAbs)

	cmd := exec.Command(exe, childArgs...)
	cmd.Stdin = nil
	cmd.Env = append(os.Environ(), daemonEnv+"=1", "MONMS_SITE="+siteAbs)
	if err := setDetached(cmd); err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("daemon: start: %w", err)
	}

	pidPath := filepath.Join(monmsDir, "monms.pid")
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)+"\n"), 0o644); err != nil {
		return fmt.Errorf("daemon: write pid file: %w", err)
	}

	fmt.Printf("daemon started (pid %d, logs %s)\n", cmd.Process.Pid, logDir)
	return nil
}

func executablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("daemon: resolve executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("daemon: resolve executable: %w", err)
	}
	return filepath.Abs(exe)
}

func PIDFilePath(siteAbs string) string {
	return filepath.Join(siteAbs, ".monms", "monms.pid")
}

func LogFilePath(siteAbs string) string {
	return logging.LogDirForSite(siteAbs)
}

func ReadPIDFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse pid file: %w", err)
	}
	return pid, nil
}
