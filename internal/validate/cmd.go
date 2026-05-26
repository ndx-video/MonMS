package validate

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/cli"
	"github.com/monms/monms/internal/config"
)

// RunCLI is the entry point for the `monms validate` subcommand (D-36, D-39).
// Called from main.go early-dispatch arm before PocketBase construction.
func RunCLI(args []string) error {
	if cmd, wantHelp := cli.ParseHelpRequest(append([]string{"validate"}, args...)); wantHelp && cmd == "validate" {
		cli.PrintHelp("validate")
		return nil
	}

	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() { cli.PrintHelp("validate") }
	var wsFlag string
	fs.StringVar(&wsFlag, "workspace", "", "workspace path (default: MONMS_WORKSPACE or ./workspace)")
	if err := fs.Parse(config.StripWorkspaceFlags(args)); err != nil {
		if err == flag.ErrHelp {
			cli.PrintHelp("validate")
			return nil
		}
		return err
	}

	_, wsAbs, err := config.ResolveWorkspace(args, os.Environ())
	if err != nil {
		return err
	}

	files := fs.Args()

	if len(files) == 0 {
		staged, err := getStagedGohtml(wsAbs)
		if err != nil {
			// D-42 graceful: git may be unavailable — warn but don't fail
			slog.Warn("validate: git not available, --files required", "err", err)
		} else {
			files = staged
		}
	}

	if len(files) == 0 {
		slog.Info("validate: no staged .gohtml files, nothing to validate")
		return nil
	}

	return ValidateFiles(wsAbs, files)
}

// getStagedGohtml returns absolute paths of staged .gohtml files in wsAbs.
// Uses git diff --cached with .Dir set to wsAbs (T-02-04: no user-supplied git args).
func getStagedGohtml(wsAbs string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM")
	cmd.Dir = wsAbs
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}

	var result []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" && strings.HasSuffix(line, ".gohtml") {
			result = append(result, filepath.Join(wsAbs, line))
		}
	}
	return result, nil
}
