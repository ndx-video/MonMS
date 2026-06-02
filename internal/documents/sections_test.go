package documents

import (
	"strings"
	"testing"
)

const sectionsFixture = `# Page Title

Intro paragraph before first h2.

## First Section

Content of first section.

### Sub section

Nested content.

## Second Section

Second section body.
`

func TestParseSections(t *testing.T) {
	sections, err := ParseSections(sectionsFixture)
	if err != nil {
		t.Fatalf("ParseSections: %v", err)
	}

	h2s := filterLevel(sections, 2)
	if len(h2s) != 2 {
		t.Fatalf("h2 count = %d, want 2", len(h2s))
	}
	if h2s[0].Index != 0 || h2s[0].Title != "First Section" {
		t.Fatalf("first h2 = %+v", h2s[0])
	}
	if h2s[1].Index != 1 || h2s[1].Title != "Second Section" {
		t.Fatalf("second h2 = %+v", h2s[1])
	}
	if !strings.Contains(h2s[0].Source, "Content of first section") {
		t.Fatalf("first h2 source = %q", h2s[0].Source)
	}
	if strings.Contains(h2s[0].Source, "Second Section") {
		t.Fatalf("first h2 should not include second section")
	}

	h3s := filterLevel(sections, 3)
	if len(h3s) != 1 || h3s[0].Title != "Sub section" {
		t.Fatalf("h3 = %+v", h3s)
	}
}

func TestSectionAt(t *testing.T) {
	sections, err := ParseSections(sectionsFixture)
	if err != nil {
		t.Fatalf("ParseSections: %v", err)
	}

	sec, ok := SectionAt(sections, 2, 1)
	if !ok || sec.Title != "Second Section" {
		t.Fatalf("SectionAt(2,1) = %+v ok=%v", sec, ok)
	}
	if _, ok := SectionAt(sections, 2, 99); ok {
		t.Fatal("expected missing section")
	}
}

func TestRenderSection(t *testing.T) {
	sections, err := ParseSections("## Hello\n\n**bold** text")
	if err != nil {
		t.Fatalf("ParseSections: %v", err)
	}
	html, err := RenderSection(sections[0].Source)
	if err != nil {
		t.Fatalf("RenderSection: %v", err)
	}
	if !strings.Contains(string(html), "<strong>bold</strong>") {
		t.Fatalf("html = %q", string(html))
	}
}

func TestSectionViews(t *testing.T) {
	views, err := SectionViews("## Title\n\nParagraph.")
	if err != nil {
		t.Fatalf("SectionViews: %v", err)
	}
	if len(views) != 1 || views[0].Title != "Title" {
		t.Fatalf("views = %+v", views)
	}
	if views[0].HTML == "" {
		t.Fatal("expected rendered HTML")
	}
}

func filterLevel(sections []Section, level int) []Section {
	var out []Section
	for _, s := range sections {
		if s.Level == level {
			out = append(out, s)
		}
	}
	return out
}
