package router

import (
	"html/template"
	"log/slog"
	"path/filepath"

	"github.com/monms/monms/internal/documents"
	"github.com/monms/monms/internal/templates"
	"github.com/pocketbase/pocketbase/core"
)

// tryRenderDocument renders a markdown-backed document when no .gohtml template exists.
func tryRenderDocument(e *core.RequestEvent, siteAbs string, cache *templates.TemplateCache, slug string, isDev bool) (bool, error) {
	match, err := documents.FindBySlug(e.App, siteAbs, slug)
	if err != nil {
		return false, err
	}
	if match == nil {
		return false, nil
	}

	bodyField := match.Binding.Monms.Body
	if bodyField == "" {
		bodyField = "body"
	}
	rawBody := match.Record.GetString(bodyField)
	htmlBody, err := documents.RenderMarkdown(rawBody)
	if err != nil {
		slog.Warn("document render: markdown failed", "slug", slug, "error", err)
		htmlBody = template.HTML(template.HTMLEscapeString(rawBody))
	}

	sections, err := documents.SectionViews(rawBody)
	if err != nil {
		slog.Warn("document render: sections failed", "slug", slug, "error", err)
		sections = nil
	}

	docData := map[string]any{
		"Title":       match.Record.GetString("title"),
		"Slug":        match.Record.GetString("slug"),
		"Collection":  match.Collection,
		"PublishedAt": match.Record.GetString("published_at"),
		"HTMLBody":    htmlBody,
		"Sections":    sections,
	}

	layoutPath := filepath.Join(siteAbs, "templates", "layouts", "base.gohtml")
	docPath := filepath.Join(siteAbs, "templates", "doc.gohtml")

	loader := func() (*template.Template, error) {
		return template.New("base").
			Funcs(DocumentTemplateFuncs(e.App, siteAbs)).
			ParseFiles(layoutPath, docPath)
	}

	cacheKey := "doc:" + slug
	tmpl, err := cache.Get(cacheKey, loader)
	if err != nil {
		if isDev {
			return false, err
		}
		return false, nil
	}

	data := enrichSSRData(e, siteAbs, slug, isDev)
	data["Doc"] = docData

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	if execErr := tmpl.ExecuteTemplate(e.Response, "base", data); execErr != nil {
		return false, execErr
	}
	return true, nil
}
