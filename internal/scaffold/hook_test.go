package scaffold_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/scaffold"
)

func TestInitInstallsPreCommitHook(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH, skipping hook install test")
	}

	ws := t.TempDir()
	if err := scaffold.RunInit([]string{"--workspace=" + ws}); err != nil {
		t.Fatalf("RunInit: %v", err)
	}

	hookPath := filepath.Join(ws, ".git", "hooks", "pre-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("hook not created: %v", err)
	}

	if !strings.HasPrefix(string(data), "#!/bin/sh") {
		t.Fatalf("hook missing shebang, got prefix: %q", string(data[:min(20, len(data))]))
	}
	if !strings.Contains(string(data), "monms-validate-hook") {
		t.Fatal("hook missing idempotency marker 'monms-validate-hook'")
	}
	if !strings.Contains(string(data), "git checkout -- .") {
		t.Fatal("hook missing rollback command 'git checkout -- .' (AGT-05)")
	}
	if !strings.Contains(string(data), `validate -w "$WS_ROOT"`) {
		t.Fatal("hook missing workspace-scoped validate -w")
	}

	fi, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("stat hook: %v", err)
	}
	if fi.Mode().Perm()&0o111 == 0 {
		t.Fatalf("hook not executable, mode=%v", fi.Mode().Perm())
	}
}

func TestInitPreCommitHookIdempotent(t *testing.T) {
	ws := t.TempDir()
	hooksDir := filepath.Join(ws, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("setup hooks dir: %v", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	original := "# monms-validate-hook\noriginal content"
	if err := os.WriteFile(hookPath, []byte(original), 0o755); err != nil {
		t.Fatalf("write hook: %v", err)
	}

	if err := scaffold.RunInit([]string{"--workspace=" + ws}); err != nil {
		t.Fatalf("RunInit: %v", err)
	}

	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	if string(data) != original {
		t.Fatalf("idempotency broken: hook was overwritten\ngot:  %q\nwant: %q", string(data), original)
	}
}

func TestInitPreCommitHookOverwritesNonMonmsHook(t *testing.T) {
	ws := t.TempDir()
	hooksDir := filepath.Join(ws, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("setup hooks dir: %v", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\n# other tool\nexit 0"), 0o755); err != nil {
		t.Fatalf("write non-monms hook: %v", err)
	}

	if err := scaffold.RunInit([]string{"--workspace=" + ws}); err != nil {
		t.Fatalf("RunInit: %v", err)
	}

	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	if !strings.Contains(string(data), "monms-validate-hook") {
		t.Fatal("non-monms hook was not overwritten; expected monms-validate-hook marker after RunInit")
	}
}

// TestPreCommitHookRollback validates hook script content for the rollback path (AGT-05).
// Full exec rollback test is in integration (Wave 3 press_releases test covers AGT-03/AGT-04).
// Hook script correctness is verified by content inspection here.
func TestPreCommitHookRollback(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, ".git", "hooks"), 0o755); err != nil {
		t.Fatalf("setup hooks dir: %v", err)
	}

	if err := scaffold.RunInit([]string{"--workspace=" + ws}); err != nil {
		t.Fatalf("RunInit: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(ws, ".git", "hooks", "pre-commit"))
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}

	if !bytes.Contains(data, []byte("git checkout -- .")) {
		t.Fatal("hook script missing rollback: 'git checkout -- .'")
	}
	if !bytes.Contains(data, []byte("exit 1")) {
		t.Fatal("hook script missing 'exit 1' after rollback")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
