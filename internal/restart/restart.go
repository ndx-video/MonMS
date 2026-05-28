package restart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/monms/monms/internal/cli"
	"github.com/monms/monms/internal/config"
	"github.com/monms/monms/internal/daemon"
	"github.com/monms/monms/internal/stop"
)

const portReleaseWait = 250 * time.Millisecond

// RunCLI stops running monms serve processes for this binary, then starts serve again.
func RunCLI(args []string) error {
	if _, wantHelp := cli.ParseHelpRequest(append([]string{"restart"}, args...)); wantHelp {
		cli.PrintHelp("restart")
		return nil
	}

	stopped, err := stop.StopAll()
	if err != nil {
		return fmt.Errorf("restart: %w", err)
	}
	if len(stopped) == 0 {
		fmt.Println("no running monms instances to stop")
	} else {
		fmt.Printf("stopped %d instance(s): %v\n", len(stopped), stopped)
	}

	serveArgs := cli.EnsureServeSubcommand(args)
	res, err := config.ResolveSiteMeta(serveArgs, os.Environ())
	if err != nil {
		return fmt.Errorf("restart: site: %w", err)
	}

	if daemon.ShouldDetach(serveArgs) {
		time.Sleep(portReleaseWait)
		return daemon.Start(res.Absolute, serveArgs)
	}

	time.Sleep(portReleaseWait)

	exe, err := executablePath()
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, serveArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() != 0 {
			os.Exit(exit.ExitCode())
		}
		return err
	}
	return nil
}

// ExecCurrentProcess stops other monms instances and re-execs the current argv.
func ExecCurrentProcess() error {
	if _, err := stop.StopAll(); err != nil {
		return fmt.Errorf("restart: stop: %w", err)
	}
	time.Sleep(portReleaseWait)

	exe, err := executablePath()
	if err != nil {
		return err
	}
	if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
		return fmt.Errorf("restart: exec: %w", err)
	}
	return nil
}

func executablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("restart: resolve executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("restart: resolve executable: %w", err)
	}
	return filepath.Abs(exe)
}
