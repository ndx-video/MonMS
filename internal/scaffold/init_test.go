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
		"schema/hero_content.json",
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

// TestBaseLayoutCDN verifies DEMO-03: scaffolded workspace includes pinned CDN assets.
func TestBaseLayoutCDN(t *testing.T) {
	dir := t.TempDir()
	runInitInDir(t, dir)

	basePath := filepath.Join(dir, "templates/layouts/base.gohtml")
	base, err := os.ReadFile(basePath)
	if err != nil {
		t.Fatalf("read base.gohtml: %v", err)
	}
	baseStr := string(base)

	required := []struct {
		token string
		desc  string
	}{
		{"htmx.org@1.9.12", "HTMX CDN pin"},
		{"alpinejs@3.14.8", "Alpine CDN pin"},
		{"cdn.tailwindcss.com", "Tailwind Play CDN"},
		{"/assets/main.css", "main.css link"},
	}
	for _, r := range required {
		if !strings.Contains(baseStr, r.token) {
			t.Fatalf("base.gohtml missing %s (%s)", r.token, r.desc)
		}
	}

	indexPath := filepath.Join(dir, "templates/index.gohtml")
	index, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read index.gohtml: %v", err)
	}
	if !strings.Contains(string(index), "hero__title") {
		t.Fatal("index.gohtml missing hero__title class")
	}
}

