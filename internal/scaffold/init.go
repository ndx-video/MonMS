package scaffold

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/config"
)

// preCommitHookScript is installed into workspace/.git/hooks/pre-commit (D-40).
// The monms-validate-hook comment is the idempotency marker — do not remove it.
const preCommitHookScript = `#!/bin/sh
# monms-validate-hook — DO NOT REMOVE THIS COMMENT (idempotency marker)

if [ -n "$MONMS_BIN" ]; then
  MONMS="$MONMS_BIN"
elif command -v monms >/dev/null 2>&1; then
  MONMS="monms"
else
  HOOK_DIR="$(cd "$(dirname "$0")" && pwd)"
  CANDIDATE="$HOOK_DIR/../../monms"
  if [ -x "$CANDIDATE" ]; then
    MONMS="$(cd "$(dirname "$CANDIDATE")" && pwd)/$(basename "$CANDIDATE")"
  else
    echo "monms: binary not found" >&2
    echo "  Set MONMS_BIN, add monms to PATH, or place binary at ../../monms relative to workspace" >&2
    exit 1
  fi
fi

STAGED=$(git diff --cached --name-only --diff-filter=ACM | grep '\.gohtml$')
if [ -z "$STAGED" ]; then
  exit 0
fi

if ! echo "$STAGED" | tr '\n' '\0' | xargs -0 "$MONMS" validate; then
  echo "" >&2
  echo "Pre-commit validation failed. Rolling back workspace to last stable state..." >&2
  git checkout -- .
  echo "Workspace restored. Fix the errors above, then re-apply your changes." >&2
  exit 1
fi

exit 0
`

type scaffoldFile struct {
	embedPath string
	destPath  string
}

var scaffoldFiles = []scaffoldFile{
	{"embed/base.gohtml", "templates/layouts/base.gohtml"},
	{"embed/index.gohtml", "templates/index.gohtml"},
	{"embed/errors.gohtml", "templates/errors/errors.gohtml"},
	{"embed/main.css", "assets/main.css"},
}

var scaffoldDirs = []string{
	"templates/layouts",
	"templates/fragments",
	"templates/errors",
	"assets",
	"schema",
}

// RunInit scaffolds a new workspace at the resolved path (D-05, D-07).
func RunInit(args []string) error {
	_, wsAbs, err := config.ResolveWorkspace(args, os.Environ())
	if err != nil {
		return err
	}

	if err := os.MkdirAll(wsAbs, 0o755); err != nil {
		return fmt.Errorf("create workspace root: %w", err)
	}

	for _, dir := range scaffoldDirs {
		if err := mkdirUnder(wsAbs, dir); err != nil {
			return err
		}
	}

	for _, sf := range scaffoldFiles {
		if err := writeScaffoldFile(wsAbs, sf.embedPath, sf.destPath); err != nil {
			return err
		}
	}

	for _, rel := range []string{"schema/.gitkeep", "templates/fragments/.gitkeep"} {
		if err := writeKeepFile(wsAbs, rel); err != nil {
			return err
		}
	}

	if err := maybeGitInit(wsAbs); err != nil {
		return err
	}

	if err := installPreCommitHook(wsAbs); err != nil {
		return err
	}

	slog.Info("workspace initialized", "path", wsAbs)
	return nil
}

func mkdirUnder(wsRoot, rel string) error {
	dest := filepath.Join(wsRoot, rel)
	if err := ensureUnderWorkspace(wsRoot, dest); err != nil {
		return err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", rel, err)
	}
	return nil
}

func writeScaffoldFile(wsRoot, embedPath, destRel string) error {
	dest := filepath.Join(wsRoot, destRel)
	if err := ensureUnderWorkspace(wsRoot, dest); err != nil {
		return err
	}

	if _, err := os.Stat(dest); err == nil {
		slog.Info("skip existing scaffold file", "path", destRel)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", destRel, err)
	}

	data, err := fs.ReadFile(scaffoldFS, embedPath)
	if err != nil {
		return fmt.Errorf("read embed %s: %w", embedPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("mkdir parent for %s: %w", destRel, err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", destRel, err)
	}
	return nil
}

func writeKeepFile(wsRoot, destRel string) error {
	dest := filepath.Join(wsRoot, destRel)
	if err := ensureUnderWorkspace(wsRoot, dest); err != nil {
		return err
	}
	if _, err := os.Stat(dest); err == nil {
		slog.Info("skip existing scaffold file", "path", destRel)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", destRel, err)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("mkdir parent for %s: %w", destRel, err)
	}
	if err := os.WriteFile(dest, nil, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", destRel, err)
	}
	return nil
}

// ensureUnderWorkspace prevents writes outside the resolved workspace (T-01-12).
func ensureUnderWorkspace(wsRoot, dest string) error {
	wsRoot = filepath.Clean(wsRoot)
	dest = filepath.Clean(dest)
	rel, err := filepath.Rel(wsRoot, dest)
	if err != nil {
		return fmt.Errorf("resolve path under workspace: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("refusing write outside workspace: %s", dest)
	}
	return nil
}

// installPreCommitHook writes the monms pre-commit hook into workspace/.git/hooks/pre-commit (D-40).
// Idempotent: skips if the file already contains the monms-validate-hook marker (D-40).
// Overwrites hooks that lack the marker, so non-monms hooks are replaced silently (T-02-07 accepted).
func installPreCommitHook(wsRoot string) error {
	gitDir := filepath.Join(wsRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		slog.Warn("no .git directory found, skipping pre-commit hook install", "dir", gitDir)
		return nil
	} else if err != nil {
		return fmt.Errorf("stat .git: %w", err)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		// A4: hooks/ may be absent right after git init on some systems.
		if err := os.MkdirAll(hooksDir, 0o755); err != nil {
			return fmt.Errorf("mkdir hooks dir: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("stat hooks dir: %w", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	if existing, err := os.ReadFile(hookPath); err == nil {
		if bytes.Contains(existing, []byte("monms-validate-hook")) {
			slog.Info("pre-commit hook already installed, skipping")
			return nil
		}
	}

	if err := os.WriteFile(hookPath, []byte(preCommitHookScript), 0o755); err != nil {
		return fmt.Errorf("install pre-commit hook: %w", err)
	}
	slog.Info("pre-commit hook installed", "path", hookPath)
	return nil
}

func maybeGitInit(wsRoot string) error {
	gitDir := filepath.Join(wsRoot, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		slog.Info("git repository already exists, skipping git init")
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat .git: %w", err)
	}

	if _, err := exec.LookPath("git"); err != nil {
		slog.Warn("git not found in PATH; skipping git init")
		return nil
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = wsRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %w: %s", err, string(out))
	}
	return nil
}
