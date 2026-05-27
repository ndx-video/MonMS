package router

import (
	"errors"
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/content"
	"github.com/monms/monms/internal/templates"
	"github.com/pocketbase/pocketbase/core"
)

// Deps holds dependencies for route registration.
type Deps struct {
	SiteAbs string
	Cache *templates.TemplateCache
	IsDev bool
}

// RegisterRoutes wires assets, fragments, then SSR catch-all (D-14).
func RegisterRoutes(se *core.ServeEvent, deps Deps) {
	se.Router.GET("/assets/{path...}", AssetsHandler(deps.SiteAbs))
	se.Router.GET("/fragments/{name}", withAuthCookie(FragmentsHandler(deps.SiteAbs, deps.Cache)))
	se.Router.GET("/{slug...}", withAuthCookie(SSRHandler(deps.SiteAbs, deps.Cache, deps.IsDev)))
}

func withAuthCookie(next func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		_ = LoadAuthFromCookie(e)
		return next(e)
	}
}

// SSRHandler renders full pages via base layout and template cache.
func SSRHandler(siteAbs string, cache *templates.TemplateCache, isDev bool) func(*core.RequestEvent) error {
	layoutPath := filepath.Join(siteAbs, "templates", "layouts", "base.gohtml")
	return func(e *core.RequestEvent) error {
		slug := e.Request.PathValue("slug")
		if isReservedSlug(slug) {
			return renderErrorPage(e, siteAbs, cache, isDev, http.StatusNotFound,
				fmt.Sprintf("Page not found: %s", e.Request.URL.Path), nil)
		}

		pagePath, err := templates.ResolveSlug(siteAbs, slug)
		if errors.Is(err, templates.ErrNotFound) {
			return renderErrorPage(e, siteAbs, cache, isDev, http.StatusNotFound,
				fmt.Sprintf("Page not found: %s", e.Request.URL.Path), nil)
		}
		if err != nil {
			return err
		}

		cacheKey := slug
		if cacheKey == "" {
			cacheKey = "index"
		}

		loader := func() (*template.Template, error) {
			return template.ParseFiles(layoutPath, pagePath)
		}

		tmpl, err := cache.Get(cacheKey, loader)
		if err != nil {
			return renderErrorPage(e, siteAbs, cache, isDev, http.StatusInternalServerError, "", err)
		}

		data := enrichSSRData(e, siteAbs, slug, isDev)
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.ExecuteTemplate(e.Response, "base", data)
	}
}

func ssrData(e *core.RequestEvent, slug string) map[string]any {
	return map[string]any{
		"IsLoggedIn": e.Auth != nil,
		"User":       e.Auth,
		"Slug":       slug,
		"Path":       e.Request.URL.Path,
	}
}

func enrichSSRData(e *core.RequestEvent, siteAbs, slug string, isDev bool) map[string]any {
	data := ssrData(e, slug)
	data["IsDev"] = isDev

	if e.Auth != nil {
		if token, err := e.Auth.NewAuthToken(); err == nil && token != "" {
			data["AuthToken"] = token
		}
		if cfg, err := content.LoadMonmsConfig(siteAbs); err == nil {
			data["IsPublisher"] = content.IsPublisher(e.Auth.GetString("email"), cfg.PublisherEmails)
		}
	}

	if slug == "" || slug == "index" {
		data["Hero"] = loadHero(e.App)
	}

	return data
}

func loadHero(app core.App) map[string]any {
	fallback := map[string]any{
		"Title": "MonMS is running",
		"Body":  "Your site is live. Templates load from disk — no build step required.",
		"ID":    heroRecordID,
	}

	rec, err := app.FindRecordById(heroCollection, heroRecordID)
	if err != nil || rec == nil {
		slog.Warn("hero load: homepage record missing, using fallback", "error", err)
		return fallback
	}

	return map[string]any{
		"Title": rec.GetString("title"),
		"Body":  rec.GetString("body"),
		"ID":    rec.Id,
	}
}

const heroCollection = "hero_content"
const heroRecordID = "homepage"

func isReservedSlug(slug string) bool {
	if slug == "" {
		return false
	}
	first := strings.Split(slug, "/")[0]
	switch first {
	case "api", "assets", "_", "_monms":
		return true
	default:
		return false
	}
}

func renderErrorPage(e *core.RequestEvent, siteAbs string, cache *templates.TemplateCache, isDev bool, code int, message string, parseErr error) error {
	if code == http.StatusInternalServerError {
		if !isDev {
			message = "Internal server error"
		} else if parseErr != nil {
			message = parseErr.Error()
		}
	}

	data := map[string]any{
		"Code":       code,
		"Message":    message,
		"Path":       e.Request.URL.Path,
		"IsLoggedIn": false,
		"Slug":       "",
	}

	layoutPath := filepath.Join(siteAbs, "templates", "layouts", "base.gohtml")
	errorsPath := filepath.Join(siteAbs, "templates", "errors", "errors.gohtml")

	loader := func() (*template.Template, error) {
		return template.ParseFiles(layoutPath, errorsPath)
	}

	tmpl, err := cache.Get(fmt.Sprintf("error:%d", code), loader)
	if err != nil {
		body := fallbackErrorHTML(code, e.Request.URL.Path, message)
		return e.HTML(code, body)
	}

	e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	e.Response.WriteHeader(code)
	return tmpl.ExecuteTemplate(e.Response, "base", data)
}

func fallbackErrorHTML(code int, path, message string) string {
	escPath := html.EscapeString(path)
	escMsg := html.EscapeString(message)
	if code == http.StatusNotFound {
		return fmt.Sprintf("<h1>404 Not Found</h1><p>Page not found: %s</p><a href=\"/\">Go home</a>", escPath)
	}
	return fmt.Sprintf("<h1>500 Internal Server Error</h1><p>%s</p>", escMsg)
}
