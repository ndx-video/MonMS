package documents

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestBuildForest(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, "schema/articles.json"), articlesSchema)
	testutil.WriteFile(t, filepath.Join(ws, "documents/articles/hello.md"), `---
title: Hello
---
Body
`)
	testutil.WriteFile(t, filepath.Join(ws, "documents/articles/guides/setup.md"), `---
title: Setup Guide
---
Body
`)

	app := bootstrapWithDocuments(t, ws)

	forest, err := BuildForest(app, ws)
	if err != nil {
		t.Fatalf("BuildForest: %v", err)
	}
	if len(forest.Collections) != 1 {
		t.Fatalf("collections = %d, want 1", len(forest.Collections))
	}

	tree := forest.Collections[0]
	if tree.Name != "articles" {
		t.Fatalf("name = %q", tree.Name)
	}
	if len(tree.Folders) != 1 || tree.Folders[0].Name != "guides" {
		t.Fatalf("folders = %+v", tree.Folders)
	}
	if len(tree.Folders[0].Docs) != 1 || tree.Folders[0].Docs[0].Slug != "guides/setup" {
		t.Fatalf("guide docs = %+v", tree.Folders[0].Docs)
	}
	if len(tree.Orphans) != 1 || tree.Orphans[0].Slug != "hello" {
		t.Fatalf("orphans = %+v", tree.Orphans)
	}
}
