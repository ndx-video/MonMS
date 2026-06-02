package documents

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Section is one heading-bounded slice of a markdown document.
type Section struct {
	Level  int    // 1–6 (h1–h6)
	Index  int    // 0-based among headings of this level in document order
	Title  string
	Anchor string
	Source string // markdown body until next heading of same or higher level
}

// SectionView is template-friendly section metadata for .Doc.Sections.
type SectionView struct {
	Level  int
	Index  int
	Title  string
	Anchor string
	HTML   template.HTML
}

// ParseSections splits markdown into heading-bounded sections.
func ParseSections(markdown string) ([]Section, error) {
	if strings.TrimSpace(markdown) == "" {
		return nil, nil
	}

	source := []byte(markdown)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	type headingMeta struct {
		level     int
		title     string
		anchor    string
		lineStart int
		lineEnd   int
	}

	var headings []headingMeta
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		title := headingText(h, source)
		anchor := ""
		if id, ok := h.AttributeString("id"); ok {
			if s, ok := id.(string); ok {
				anchor = s
			} else if b, ok := id.([]byte); ok {
				anchor = string(b)
			}
		}

		if h.Lines().Len() == 0 {
			return ast.WalkContinue, nil
		}
		seg := h.Lines().At(0)
		lineStart := lineNumber(source, seg.Start)
		lineEnd := lineNumber(source, seg.Stop-1)

		headings = append(headings, headingMeta{
			level:     h.Level,
			title:     title,
			anchor:    anchor,
			lineStart: lineStart,
			lineEnd:   lineEnd,
		})
		return ast.WalkContinue, nil
	})

	if len(headings) == 0 {
		return nil, nil
	}

	lines := strings.Split(markdown, "\n")
	levelCounts := make(map[int]int)

	sections := make([]Section, 0, len(headings))
	for i, h := range headings {
		contentStart := h.lineEnd + 1
		contentEnd := len(lines)
		for j := i + 1; j < len(headings); j++ {
			if headings[j].level <= h.level {
				contentEnd = headings[j].lineStart
				break
			}
		}

		var body string
		if contentStart < contentEnd {
			body = strings.TrimSpace(strings.Join(lines[contentStart:contentEnd], "\n"))
		}

		idx := levelCounts[h.level]
		levelCounts[h.level]++

		sections = append(sections, Section{
			Level:  h.level,
			Index:  idx,
			Title:  h.title,
			Anchor: h.anchor,
			Source: body,
		})
	}

	return sections, nil
}

// SectionAt returns the section at the given heading level and per-level index.
func SectionAt(sections []Section, level, index int) (*Section, bool) {
	for i := range sections {
		if sections[i].Level == level && sections[i].Index == index {
			return &sections[i], true
		}
	}
	return nil, false
}

// RenderSection renders a section's markdown source to HTML.
func RenderSection(source string) (template.HTML, error) {
	if strings.TrimSpace(source) == "" {
		return "", nil
	}
	return RenderMarkdown(source)
}

// SectionViews builds template views with pre-rendered HTML for each section.
func SectionViews(markdown string) ([]SectionView, error) {
	sections, err := ParseSections(markdown)
	if err != nil {
		return nil, err
	}
	out := make([]SectionView, 0, len(sections))
	for _, sec := range sections {
		html, err := RenderSection(sec.Source)
		if err != nil {
			return nil, fmt.Errorf("documents section render level=%d index=%d: %w", sec.Level, sec.Index, err)
		}
		out = append(out, SectionView{
			Level:  sec.Level,
			Index:  sec.Index,
			Title:  sec.Title,
			Anchor: sec.Anchor,
			HTML:   html,
		})
	}
	return out, nil
}

func headingText(h *ast.Heading, source []byte) string {
	var buf bytes.Buffer
	for child := h.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			buf.Write(n.Segment.Value(source))
		case *ast.String:
			buf.Write(n.Value)
		}
	}
	return strings.TrimSpace(buf.String())
}

func lineNumber(source []byte, offset int) int {
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}
	return bytes.Count(source[:offset], []byte("\n"))
}
