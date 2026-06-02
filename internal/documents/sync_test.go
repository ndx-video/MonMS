package documents

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/testutil"
	"github.com/pocketbase/pocketbase/core"
)

const articlesSchema = `{
  "name": "articles",
  "type": "base",
  "editorial": true,
  "monms": {
    "source": "markdown",
    "root": "documents/articles"
  },
  "fields": [
    { "name": "id", "type": "text", "primaryKey": true, "required": true, "pattern": "^[a-z][a-z0-9_-]*$", "min": 1, "max": 120 },
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "path", "type": "text" },
    { "name": "folder", "type": "text" },
    { "name": "body", "type": "text" }
  ]
}`

func bootstrapWithDocuments(t *testing.T, ws string) core.App {
	t.Helper()
	app := testutil.BootstrapApp(t, ws)
	RegisterBootstrapHook(app, ws)
	if _, err := SyncAll(app, ws); err != nil {
		t.Fatalf("SyncAll: %v", err)
	}
	return app
}

func TestSyncAll(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, "schema/articles.json"), articlesSchema)
	testutil.WriteFile(t, filepath.Join(ws, "documents/articles/hello.md"), `---
title: Hello World
---
# Hello

First post.
`)

	app := bootstrapWithDocuments(t, ws)

	rec, err := app.FindRecordById("articles", "articles--hello")
	if err != nil {
		t.Fatalf("FindRecordById: %v", err)
	}
	if rec.GetString("title") != "Hello World" {
		t.Fatalf("title = %q", rec.GetString("title"))
	}
	if rec.GetString("body") == "" {
		t.Fatal("body empty")
	}
}

func TestDiffOrphans(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, "schema/articles.json"), articlesSchema)
	testutil.WriteFile(t, filepath.Join(ws, "documents/articles/kept.md"), `---
title: Kept
---
Body
`)

	app := bootstrapWithDocuments(t, ws)

	orphan := testutil.SaveRecord(t, app, "articles", map[string]any{
		"id":    "articles--orphan",
		"title": "Orphan",
		"slug":  "orphan",
	})

	diff, err := DiffOrphans(app, ws)
	if err != nil {
		t.Fatalf("DiffOrphans: %v", err)
	}
	if len(diff.Orphans) != 1 {
		t.Fatalf("orphans = %d, want 1", len(diff.Orphans))
	}
	if diff.Orphans[0].ID != orphan.Id {
		t.Fatalf("orphan id = %q", diff.Orphans[0].ID)
	}
}

func TestLoadMarkdownBindingsIntegration(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, "schema/articles.json"), articlesSchema)

	bindings, err := schema.LoadMarkdownBindings(filepath.Join(ws, "schema"))
	if err != nil {
		t.Fatalf("LoadMarkdownBindings: %v", err)
	}
	if len(bindings) != 1 || bindings[0].Name != "articles" {
		t.Fatalf("bindings = %+v", bindings)
	}
}
