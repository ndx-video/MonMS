package documents

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestFindBindingAndListDocuments(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, "schema/articles.json"), articlesSchema)
	testutil.WriteFile(t, filepath.Join(ws, "documents/articles/a.md"), `---
title: A
---
body
`)

	binding, err := FindBinding(ws, "articles")
	if err != nil {
		t.Fatalf("FindBinding: %v", err)
	}
	if binding.Monms.Root != "documents/articles" {
		t.Fatalf("root = %q", binding.Monms.Root)
	}

	entries, err := ListDocuments(ws, binding)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if len(entries) != 1 || entries[0].PathKey != "a" {
		t.Fatalf("entries = %+v", entries)
	}

	path := DocFilePath(ws, binding, "a")
	if filepath.Base(path) != "a.md" {
		t.Fatalf("DocFilePath = %q", path)
	}
}

func TestStableRecordID(t *testing.T) {
	got := StableRecordID("articles", "guides/setup")
	want := "articles--guides--setup"
	if got != want {
		t.Fatalf("StableRecordID = %q, want %q", got, want)
	}
}
