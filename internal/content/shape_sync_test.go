package content

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/testutil"
	"github.com/monms/monms/internal/site"
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

func initBareRemoteWithTag(t *testing.T, ws, tag, content string) {
	t.Helper()
	requireGit(t)

	root := filepath.Dir(ws)
	bare := filepath.Join(root, "origin.git")
	runGitTest(t, root, "init", "--bare", bare)

	runGitTest(t, ws, "init")
	readme := filepath.Join(ws, "README.md")
	testutil.WriteFile(t, readme, content)
	runGitTest(t, ws, "add", "README.md")
	runGitTest(t, ws, "commit", "-m", "initial")
	runGitTest(t, ws, "tag", tag)
	runGitTest(t, ws, "remote", "add", "origin", bare)
	runGitTest(t, ws, "push", "-u", "origin", "HEAD", "--tags")
}

func TestApplyShapeSyncFromSite(t *testing.T) {
	t.Run("disabled config is no-op", func(t *testing.T) {
		ws := testutil.NewSite(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "shapeSync": { "enabled": false, "ref": "v1.0.0" }
}`)
		if err := ApplyShapeSyncFromSite(ws); err != nil {
			t.Fatalf("ApplyShapeSyncFromSite: %v", err)
		}
	})

	t.Run("enabled sync checks out ref", func(t *testing.T) {
		requireGit(t)
		ws := testutil.NewSite(t)
		initBareRemoteWithTag(t, ws, "v1.0.0", "tagged\n")

		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "shapeSync": { "enabled": true, "ref": "v1.0.0", "remote": "origin" }
}`)
		if err := ApplyShapeSyncFromSite(ws); err != nil {
			t.Fatalf("ApplyShapeSyncFromSite: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(ws, "README.md"))
		if err != nil {
			t.Fatalf("read README: %v", err)
		}
		if string(data) != "tagged\n" {
			t.Fatalf("README = %q, want tagged", data)
		}
	})

	t.Run("failOnError false continues on sync failure", func(t *testing.T) {
		ws := testutil.NewSite(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "shapeSync": { "enabled": true, "ref": "v1.0.0", "failOnError": false }
}`)
		if err := ApplyShapeSyncFromSite(ws); err != nil {
			t.Fatalf("expected warning-only failure, got %v", err)
		}
	})

	t.Run("failOnError true returns sync error", func(t *testing.T) {
		ws := testutil.NewSite(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "shapeSync": { "enabled": true, "ref": "v1.0.0", "failOnError": true }
}`)
		err := ApplyShapeSyncFromSite(ws)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, site.ErrDirtyWorktree) && !strings.Contains(err.Error(), "not a git repository") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
