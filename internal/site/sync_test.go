package site

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}
}

func runGitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.local",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.local",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
}

func setupTaggedRemoteRepo(t *testing.T) string {
	t.Helper()
	requireGit(t)

	root := t.TempDir()
	bare := filepath.Join(root, "origin.git")
	runGitTest(t, root, "init", "--bare", bare)

	ws := filepath.Join(root, "site")
	runGitTest(t, root, "clone", bare, ws)

	readme := filepath.Join(ws, "README.md")
	if err := os.WriteFile(readme, []byte("v1\n"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	runGitTest(t, ws, "add", "README.md")
	runGitTest(t, ws, "commit", "-m", "v1")
	runGitTest(t, ws, "tag", "v1.0.0")
	runGitTest(t, ws, "push", "origin", "HEAD")
	runGitTest(t, ws, "push", "origin", "v1.0.0")

	if err := os.WriteFile(readme, []byte("v2\n"), 0o644); err != nil {
		t.Fatalf("write readme v2: %v", err)
	}
	runGitTest(t, ws, "add", "README.md")
	runGitTest(t, ws, "commit", "-m", "v2")
	runGitTest(t, ws, "tag", "v2.0.0")
	runGitTest(t, ws, "push", "origin", "HEAD", "--tags")

	runGitTest(t, ws, "checkout", "v1.0.0")
	return ws
}

func readREADME(t *testing.T, ws string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(ws, "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	return string(data)
}

func TestSyncChecksOutTaggedRef(t *testing.T) {
	ws := setupTaggedRemoteRepo(t)

	if err := Sync(ws, SyncOptions{Ref: "v2.0.0"}); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if got := readREADME(t, ws); got != "v2\n" {
		t.Fatalf("README = %q, want v2", got)
	}
}

func TestSyncDirtyWorktreeBlocked(t *testing.T) {
	ws := setupTaggedRemoteRepo(t)

	readme := filepath.Join(ws, "README.md")
	if err := os.WriteFile(readme, []byte("dirty\n"), 0o644); err != nil {
		t.Fatalf("write dirty: %v", err)
	}

	err := Sync(ws, SyncOptions{Ref: "v2.0.0"})
	if err == nil {
		t.Fatal("expected error on dirty worktree")
	}
	if !errors.Is(err, ErrDirtyWorktree) {
		t.Fatalf("expected ErrDirtyWorktree, got %v", err)
	}
}

func TestSyncForceAllowsDirtyWorktree(t *testing.T) {
	ws := setupTaggedRemoteRepo(t)

	readme := filepath.Join(ws, "README.md")
	if err := os.WriteFile(readme, []byte("dirty\n"), 0o644); err != nil {
		t.Fatalf("write dirty: %v", err)
	}

	if err := Sync(ws, SyncOptions{Ref: "v2.0.0", Force: true}); err != nil {
		t.Fatalf("Sync force: %v", err)
	}
	if got := readREADME(t, ws); got != "v2\n" {
		t.Fatalf("README = %q, want v2", got)
	}
}

func TestSyncRequiresRef(t *testing.T) {
	ws := setupTaggedRemoteRepo(t)
	if err := Sync(ws, SyncOptions{}); err == nil {
		t.Fatal("expected error for empty ref")
	}
}

func TestSyncRequiresGitRepo(t *testing.T) {
	dir := t.TempDir()
	if err := Sync(dir, SyncOptions{Ref: "v1.0.0"}); err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func TestShapeSyncOptionsFromConfig(t *testing.T) {
	opts, ok := ShapeSyncOptionsFromConfig(nil)
	if ok || opts.Ref != "" {
		t.Fatalf("nil config should not enable sync: %+v %v", opts, ok)
	}

	opts, ok = ShapeSyncOptionsFromConfig(&ShapeSyncConfig{Enabled: true})
	if ok {
		t.Fatal("enabled without ref should not sync")
	}

	opts, ok = ShapeSyncOptionsFromConfig(&ShapeSyncConfig{
		Enabled: true,
		Ref:     "v1.2.0",
		Remote:  "upstream",
		Force:   true,
	})
	if !ok {
		t.Fatal("expected enabled config")
	}
	if opts.Ref != "v1.2.0" || opts.Remote != "upstream" || !opts.Force {
		t.Fatalf("got %+v", opts)
	}
}
