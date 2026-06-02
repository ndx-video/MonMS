package documents

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/schema"
)

func TestParseFileWithFrontmatter(t *testing.T) {
	content := `---
title: Setup Guide
status: published
tags:
  - guides
  - onboarding
---
# Hello

Body text.
`
	doc, err := ParseBytes("setup.md", []byte(content))
	if err != nil {
		t.Fatalf("ParseBytes: %v", err)
	}
	if doc.Meta["title"] != "Setup Guide" {
		t.Fatalf("title = %v", doc.Meta["title"])
	}
	if !strings.Contains(doc.Body, "# Hello") {
		t.Fatalf("body = %q", doc.Body)
	}
}

func TestRecordFromDocument(t *testing.T) {
	binding := schema.CollectionMeta{
		Name: "articles",
		Monms: schema.MonmsBinding{
			Root:   "documents/articles",
			Fields: map[string]string{"date": "published_at"},
		},
	}

	doc := ParsedDocument{
		FilePath: filepath.Join("documents", "articles", "guides", "setup.md"),
		Meta: map[string]any{
			"title": "Setup Guide",
			"date":  "2024-03-15",
		},
		Body: "# Setup\n\nSteps here.",
	}

	record, err := RecordFromDocument(binding, doc, filepath.Join("documents", "articles"))
	if err != nil {
		t.Fatalf("RecordFromDocument: %v", err)
	}
	if record["id"] != "articles--guides--setup" {
		t.Fatalf("id = %v", record["id"])
	}
	if record["slug"] != "guides/setup" {
		t.Fatalf("slug = %v", record["slug"])
	}
	if record["folder"] != "guides" {
		t.Fatalf("folder = %v", record["folder"])
	}
	if record["title"] != "Setup Guide" {
		t.Fatalf("title = %v", record["title"])
	}
	if record["published_at"] != "2024-03-15" {
		t.Fatalf("published_at = %v", record["published_at"])
	}
	if record["body"] != "# Setup\n\nSteps here." {
		t.Fatalf("body = %v", record["body"])
	}
}

func TestWriteFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.md"
	meta := map[string]any{
		"id":    "articles/test",
		"title": "Test",
	}
	body := "# Title\n\nContent."

	if err := WriteFile(path, meta, body); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	doc, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if doc.Meta["id"] != "articles/test" {
		t.Fatalf("id = %v", doc.Meta["id"])
	}
	if !bytes.Contains([]byte(doc.Body), []byte("Content.")) {
		t.Fatalf("body = %q", doc.Body)
	}
}

func TestRecordFromDocumentExplicitID(t *testing.T) {
	binding := schema.CollectionMeta{
		Name: "articles",
		Monms: schema.MonmsBinding{
			Root:   "documents/articles",
			IDFrom: "frontmatter.id",
		},
	}
	doc := ParsedDocument{
		FilePath: "documents/articles/page.md",
		Meta:     map[string]any{"id": "custom-id", "title": "Page"},
		Body:     "Body",
	}
	record, err := RecordFromDocument(binding, doc, filepath.Join("documents", "articles"))
	if err != nil {
		t.Fatalf("RecordFromDocument: %v", err)
	}
	if record["id"] != "custom-id" {
		t.Fatalf("id = %v", record["id"])
	}
}
