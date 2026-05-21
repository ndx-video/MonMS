package scaffold

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/config"
)

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
