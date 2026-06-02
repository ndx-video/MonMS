package documents

import "testing"

func TestInferDoctreeTitleFromH1(t *testing.T) {
	doc := ParsedDocument{Body: "# My Doc\n\nParagraph.\n"}
	if got := inferDoctreeTitle(doc, "intro.md"); got != "My Doc" {
		t.Fatalf("title = %q, want My Doc", got)
	}
}

func TestInferDoctreeTitleFromBasename(t *testing.T) {
	doc := ParsedDocument{Body: "No heading\n"}
	if got := inferDoctreeTitle(doc, "welcome-page.md"); got != "welcome page" {
		t.Fatalf("title = %q, want welcome page", got)
	}
}

func TestMergeDoctreeFrontmatterIdempotentTitle(t *testing.T) {
	meta := map[string]any{
		"title":     "Kept",
		"id":        "dt_guide--page",
		"ts_create": "2024-01-01T00:00:00Z",
		"ts_mod":    "2024-01-02T00:00:00Z",
	}
	doc := ParsedDocument{Body: "# Other\n", Meta: meta}
	out, changed := mergeDoctreeFrontmatter(meta, doc, "dt_guide", "page", "/tmp/page.md", false)
	if changed {
		t.Fatal("expected no change when title and timestamps set")
	}
	if out["title"] != "Kept" {
		t.Fatalf("title = %v", out["title"])
	}
}
