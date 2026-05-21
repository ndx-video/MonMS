package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// NewWorkspace creates a temp directory with minimal valid workspace structure.
func NewWorkspace(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "templates/layouts/base.gohtml"), `{{define "base"}}
<!DOCTYPE html>
<html><body>{{template "body" .}}</body></html>
{{end}}
`)
	writeFile(t, filepath.Join(dir, "templates/index.gohtml"), `{{define "body"}}ok{{end}}
`)
	writeFile(t, filepath.Join(dir, "assets/main.css"), "/* test */\n")
	if err := os.MkdirAll(filepath.Join(dir, "schema"), 0o755); err != nil {
		t.Fatalf("mkdir schema: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "templates/fragments"), 0o755); err != nil {
		t.Fatalf("mkdir templates/fragments: %v", err)
	}

	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
