package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// NewSite creates a temp directory with minimal valid site structure.
func NewSite(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "templates/layouts/base.gohtml"), `{{define "base"}}
<!DOCTYPE html>
<html><body>{{template "body" .}}</body></html>
{{end}}
`)
	writeFile(t, filepath.Join(dir, "templates/index.gohtml"), `{{define "body"}}
<section class="hero">
  <h1 class="hero__title">MonMS is running</h1>
</section>
{{end}}
`)
	writeFile(t, filepath.Join(dir, "templates/errors/errors.gohtml"), `{{define "body"}}
<section class="error-page">
  <p class="error-page__code">{{.Code}}</p>
  <h1 class="error-page__title">Page not found</h1>
  <p class="error-page__message">{{.Message}}</p>
</section>
{{end}}
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

// WriteFile creates parent dirs and writes content (test helper).
func WriteFile(t *testing.T, path, content string) {
	writeFile(t, path, content)
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
