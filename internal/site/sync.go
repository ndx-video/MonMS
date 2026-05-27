package site

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrDirtyWorktree is returned when sync refuses to run on a dirty git worktree.
var ErrDirtyWorktree = errors.New("site git worktree has uncommitted changes")

// Sync fetches tags from remote and checks out ref in the site Git repo.
func Sync(siteAbs string, opts SyncOptions) error {
	opts = opts.normalized()
	if opts.Ref == "" {
		return fmt.Errorf("site sync: ref is required")
	}

	gitDir := filepath.Join(siteAbs, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("site sync: not a git repository (missing .git)")
		}
		return fmt.Errorf("site sync: stat .git: %w", err)
	}

	if !opts.Force {
		dirty, err := isDirtyWorktree(siteAbs)
		if err != nil {
			return err
		}
		if dirty {
			return fmt.Errorf("site sync: %w (use --force to discard local changes)", ErrDirtyWorktree)
		}
	}

	if err := runGit(siteAbs, "fetch", "--tags", opts.Remote); err != nil {
		return fmt.Errorf("site sync: fetch: %w", err)
	}

	checkoutArgs := []string{"checkout"}
	if opts.Force {
		checkoutArgs = append(checkoutArgs, "--force")
	}
	checkoutArgs = append(checkoutArgs, opts.Ref)

	if err := runGit(siteAbs, checkoutArgs...); err != nil {
		return fmt.Errorf("site sync: checkout: %w", err)
	}

	slog.Info("site sync: complete", "ref", opts.Ref, "remote", opts.Remote, "site", siteAbs)
	return nil
}

func (o SyncOptions) normalized() SyncOptions {
	o.Ref = strings.TrimSpace(o.Ref)
	o.Remote = strings.TrimSpace(o.Remote)
	if o.Remote == "" {
		o.Remote = "origin"
	}
	return o
}

// ShapeSyncOptionsFromConfig builds sync options from config when enabled.
func ShapeSyncOptionsFromConfig(cfg *ShapeSyncConfig) (SyncOptions, bool) {
	if cfg == nil || !cfg.Enabled {
		return SyncOptions{}, false
	}
	ref := strings.TrimSpace(cfg.Ref)
	if ref == "" {
		return SyncOptions{}, false
	}
	return SyncOptions{
		Ref:    ref,
		Remote: cfg.Remote,
		Force:  cfg.Force,
	}, true
}

func isDirtyWorktree(siteAbs string) (bool, error) {
	out, err := gitOutput(siteAbs, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("site sync: status: %w", err)
	}
	return strings.TrimSpace(out) != "", nil
}

func runGit(dir string, args ...string) error {
	_, err := gitOutput(dir, args...)
	return err
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return string(out), nil
}
