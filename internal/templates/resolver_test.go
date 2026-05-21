package templates

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestResolveSlug(t *testing.T) {
	t.Parallel()

	t.Run("empty slug resolves to index", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		got, err := ResolveSlug(ws, "")
		if err != nil {
			t.Fatalf("ResolveSlug: %v", err)
		}
		want := filepath.Join(ws, "templates", "index.gohtml")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("press resolves to directory index", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		writeTemplate(t, ws, "templates/press/index.gohtml", `{{define "body"}}press{{end}}`)

		got, err := ResolveSlug(ws, "press")
		if err != nil {
			t.Fatalf("ResolveSlug: %v", err)
		}
		want := filepath.Join(ws, "templates", "press", "index.gohtml")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("nested slug resolves to flat file", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		writeTemplate(t, ws, "templates/press/2024.gohtml", `{{define "body"}}2024{{end}}`)

		got, err := ResolveSlug(ws, "press/2024")
		if err != nil {
			t.Fatalf("ResolveSlug: %v", err)
		}
		want := filepath.Join(ws, "templates", "press", "2024.gohtml")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("trailing slash stripped", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		writeTemplate(t, ws, "templates/about/index.gohtml", `{{define "body"}}about{{end}}`)

		got, err := ResolveSlug(ws, "about/")
		if err != nil {
			t.Fatalf("ResolveSlug: %v", err)
		}
		want := filepath.Join(ws, "templates", "about", "index.gohtml")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("flat file wins over directory index", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		writeTemplate(t, ws, "templates/about.gohtml", `{{define "body"}}flat{{end}}`)
		writeTemplate(t, ws, "templates/about/index.gohtml", `{{define "body"}}index{{end}}`)

		got, err := ResolveSlug(ws, "about")
		if err != nil {
			t.Fatalf("ResolveSlug: %v", err)
		}
		want := filepath.Join(ws, "templates", "about.gohtml")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("missing template returns ErrNotFound", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)

		_, err := ResolveSlug(ws, "missing")
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("slug case preserved", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		writeTemplate(t, ws, "templates/Press/index.gohtml", `{{define "body"}}Press{{end}}`)

		got, err := ResolveSlug(ws, "Press")
		if err != nil {
			t.Fatalf("ResolveSlug: %v", err)
		}
		want := filepath.Join(ws, "templates", "Press", "index.gohtml")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

func writeTemplate(t *testing.T, ws, relPath, content string) {
	t.Helper()
	path := filepath.Join(ws, relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
