package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/workspace"
)

func runInitInDir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := RunInit([]string{"--workspace", dir}); err != nil {
		t.Fatalf("RunInit: %v", err)
	}
}

func TestInitScaffold(t *testing.T) {
	dir := t.TempDir()
	runInitInDir(t, dir)

	for _, rel := range []string{
		"schema",
		"templates/fragments",
		"templates/layouts",
		"templates/errors",
		"assets",
	} {
		path := filepath.Join(dir, rel)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", rel, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", rel)
		}
	}

	for _, rel := range []string{
		"templates/layouts/base.gohtml",
		"templates/index.gohtml",
		"templates/errors/errors.gohtml",
		"assets/main.css",
		"schema/.gitkeep",
		"templates/fragments/.gitkeep",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}

	if err := workspace.ValidateWorkspace(dir); err != nil {
		t.Fatalf("ValidateWorkspace: %v", err)
	}
}

func TestInitGit(t *testing.T) {
	t.Run("creates git repo when absent", func(t *testing.T) {
		dir := t.TempDir()
		runInitInDir(t, dir)
		if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
			t.Fatalf(".git missing after init: %v", err)
		}
	})

	t.Run("skips git init when present", func(t *testing.T) {
		dir := t.TempDir()
		gitDir := filepath.Join(dir, ".git")
		if err := os.Mkdir(gitDir, 0o755); err != nil {
			t.Fatalf("mkdir .git: %v", err)
		}
		marker := filepath.Join(gitDir, "existing")
		if err := os.WriteFile(marker, []byte("1"), 0o644); err != nil {
			t.Fatalf("write marker: %v", err)
		}
		runInitInDir(t, dir)
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("existing .git was replaced: %v", err)
		}
	})

	t.Run("no pb_data gitignore", func(t *testing.T) {
		dir := t.TempDir()
		runInitInDir(t, dir)
		data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
		if err == nil && strings.Contains(string(data), ".pb_data") {
			t.Fatal("init must not auto-gitignore .pb_data per D-28")
		}
	})
}

func TestBaseLayoutCDN(t *testing.T) {
	t.Skip("implemented in task 3")
}

