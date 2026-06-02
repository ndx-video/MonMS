package documents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestValidateDoctreeStub(t *testing.T) {
	if err := ValidateDoctreeStub("guide"); err != nil {
		t.Fatalf("guide: %v", err)
	}
	if err := ValidateDoctreeStub("documents"); err == nil {
		t.Fatal("expected reject documents")
	}
}

func TestDiscoverUsesPathDrivenName(t *testing.T) {
	site := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(site, "guide", "tutorials", "page.md"), "# Hi\n")
	candidates, err := DiscoverLeafCollections(site, "guide")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("candidates = %+v", candidates)
	}
	if candidates[0].Collection != "dt_guide_tutorials" {
		t.Fatalf("collection = %q, want dt_guide_tutorials", candidates[0].Collection)
	}
}

func TestCopyAndDiscoverLeafOnly(t *testing.T) {
	legacy := t.TempDir()
	testutil.WriteFile(t, filepath.Join(legacy, "parent-only", "child", "leaf.md"), "# Child\n")
	testutil.WriteFile(t, filepath.Join(legacy, "parent-only", "sibling.md"), "# Sibling\n")

	site := testutil.NewSite(t)
	result, err := CopySourceTree(site, legacy, "guide")
	if err != nil {
		t.Fatalf("CopySourceTree: %v", err)
	}
	if result.FilesCopied != 2 {
		t.Fatalf("copied = %d, want 2", result.FilesCopied)
	}

	candidates, err := DiscoverLeafCollections(site, "guide")
	if err != nil {
		t.Fatalf("DiscoverLeafCollections: %v", err)
	}
	// parent-only has sibling.md directly — leaf
	// parent-only/child has leaf.md directly — leaf
	// parent-only alone has no direct md
	if len(candidates) != 2 {
		t.Fatalf("candidates = %d, want 2; %+v", len(candidates), candidates)
	}
}

func TestEnsureFrontmatterIdempotent(t *testing.T) {
	site := testutil.NewSite(t)
	root := filepath.Join(site, "guide", "docs")
	testutil.WriteFile(t, filepath.Join(root, "page.md"), "# Hi\n")

	candidates, err := DiscoverLeafCollections(site, "guide")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("candidates = %+v", candidates)
	}

	_, err = FinalizeBindings(site, candidates, FinalizeOptions{})
	if err != nil {
		t.Fatalf("finalize 1: %v", err)
	}
	data1, err := os.ReadFile(filepath.Join(root, "page.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	_, err = FinalizeBindings(site, candidates, FinalizeOptions{})
	if err != nil {
		t.Fatalf("finalize 2: %v", err)
	}
	data2, err := os.ReadFile(filepath.Join(root, "page.md"))
	if err != nil {
		t.Fatalf("read2: %v", err)
	}
	if string(data1) != string(data2) {
		t.Fatalf("second finalize changed file:\n%s\nvs\n%s", data1, data2)
	}
}

func TestEnsureFrontmatterFixesWrongID(t *testing.T) {
	site := testutil.NewSite(t)
	root := filepath.Join(site, "guide")
	testutil.WriteFile(t, filepath.Join(root, "a.md"), `---
id: wrong--id
title: A
---
Body
`)

	candidates, err := DiscoverLeafCollections(site, "guide")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	_, err = FinalizeBindings(site, candidates, FinalizeOptions{})
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}
	doc, err := ParseFile(filepath.Join(root, "a.md"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := recordID(candidates[0].Collection, "a")
	if doc.Meta["id"] != want {
		t.Fatalf("id = %v, want %s", doc.Meta["id"], want)
	}
}

func TestPruneCopiedTree(t *testing.T) {
	site := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(site, "guide", "x.md"), "# X\n")
	if err := PruneCopiedTree(site, "guide"); err != nil {
		t.Fatalf("prune: %v", err)
	}
	if _, err := os.Stat(filepath.Join(site, "guide")); !os.IsNotExist(err) {
		t.Fatal("guide should be removed")
	}
}

func TestFinalizeInjectsTitleFromH1(t *testing.T) {
	site := testutil.NewSite(t)
	root := filepath.Join(site, "guide")
	testutil.WriteFile(t, filepath.Join(root, "intro.md"), "# My Doc\n\nBody.\n")

	candidates, err := DiscoverLeafCollections(site, "guide")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	_, err = FinalizeBindings(site, candidates, FinalizeOptions{})
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}
	doc, err := ParseFile(filepath.Join(root, "intro.md"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc.Meta["title"] != "My Doc" {
		t.Fatalf("title = %v, want My Doc", doc.Meta["title"])
	}
	if _, ok := doc.Meta["ts_mod"].(string); !ok || doc.Meta["ts_mod"] == "" {
		t.Fatalf("expected ts_mod in frontmatter")
	}
}

func TestFinalizeWritesDtSchema(t *testing.T) {
	site := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(site, "guide", "x.md"), "# X\n")
	candidates, err := DiscoverLeafCollections(site, "guide")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	_, err = FinalizeBindings(site, candidates, FinalizeOptions{})
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}
	schemaPath := filepath.Join(site, "schema", candidates[0].Collection+".json")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	if !strings.Contains(string(data), `"doctree_id"`) {
		t.Fatalf("schema missing doctree_id: %s", data)
	}
	if !strings.Contains(string(data), `"doctree": "guide"`) {
		t.Fatalf("schema missing monms.doctree: %s", data)
	}
}

func TestRescanDiscoversNewLeafFolder(t *testing.T) {
	site := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(site, "guide", "a.md"), "# A\n")
	c1, err := DiscoverLeafCollections(site, "guide")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(c1) != 1 {
		t.Fatalf("want 1 candidate, got %d", len(c1))
	}

	testutil.WriteFile(t, filepath.Join(site, "guide", "extra", "b.md"), "# B\n")
	c2, err := DiscoverLeafCollections(site, "guide")
	if err != nil {
		t.Fatalf("discover2: %v", err)
	}
	if len(c2) != 2 {
		t.Fatalf("want 2 candidates after new folder, got %d", len(c2))
	}
}
