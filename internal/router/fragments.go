package router

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/monms/monms/internal/templates"
	"github.com/pocketbase/pocketbase/core"
)

// FragmentsHandler renders HTMX partials from templates/fragments/{name}.gohtml (D-15).
func FragmentsHandler(wsAbs string, cache *templates.TemplateCache) func(*core.RequestEvent) error {
	fragDir := filepath.Join(wsAbs, "templates", "fragments")
	return func(e *core.RequestEvent) error {
		name := e.Request.PathValue("name")
		fragPath := filepath.Join(fragDir, name+".gohtml")
		if _, err := os.Stat(fragPath); err != nil {
			if os.IsNotExist(err) {
				return e.HTML(http.StatusNotFound, "Not Found")
			}
			return err
		}

		loader := func() (*template.Template, error) {
			return template.ParseFiles(fragPath)
		}
		tmpl, err := cache.Get("fragment:"+name, loader)
		if err != nil {
			return e.HTML(http.StatusInternalServerError, "Internal Server Error")
		}

		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.Execute(e.Response, nil)
	}
}
