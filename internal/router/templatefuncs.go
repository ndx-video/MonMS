package router

import (
	"fmt"
	"html/template"

	"github.com/monms/monms/internal/documents"
	"github.com/pocketbase/pocketbase/core"
)

// DocumentTemplateFuncs returns template helpers for pulling markdown sections into .gohtml pages.
func DocumentTemplateFuncs(app core.App, siteAbs string) template.FuncMap {
	return template.FuncMap{
		"docSection": func(slug string, level, index int) (template.HTML, error) {
			sec, err := loadSection(app, siteAbs, slug, level, index)
			if err != nil {
				return "", err
			}
			if sec == nil {
				return "", nil
			}
			return documents.RenderSection(sec.Source)
		},
		"docHeading": func(slug string, level, index int) (string, error) {
			sec, err := loadSection(app, siteAbs, slug, level, index)
			if err != nil {
				return "", err
			}
			if sec == nil {
				return "", nil
			}
			return sec.Title, nil
		},
		"docSections": func(slug string) ([]documents.Section, error) {
			body, err := documentBody(app, siteAbs, slug)
			if err != nil {
				return nil, err
			}
			if body == "" {
				return nil, nil
			}
			return documents.ParseSections(body)
		},
	}
}

func loadSection(app core.App, siteAbs, slug string, level, index int) (*documents.Section, error) {
	body, err := documentBody(app, siteAbs, slug)
	if err != nil {
		return nil, err
	}
	if body == "" {
		return nil, nil
	}
	sections, err := documents.ParseSections(body)
	if err != nil {
		return nil, err
	}
	sec, ok := documents.SectionAt(sections, level, index)
	if !ok {
		return nil, nil
	}
	return sec, nil
}

func documentBody(app core.App, siteAbs, slug string) (string, error) {
	match, err := documents.FindBySlug(app, siteAbs, slug)
	if err != nil {
		return "", err
	}
	if match == nil {
		return "", fmt.Errorf("document not found: %s", slug)
	}
	field := match.Binding.Monms.Body
	if field == "" {
		field = "body"
	}
	return match.Record.GetString(field), nil
}
