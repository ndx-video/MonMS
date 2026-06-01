package monmsdash

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	monmsui "github.com/monms/monms/internal/monmsdash/ui"
	"github.com/pocketbase/pocketbase/core"
)

type templates struct {
	baseSources map[string]string
	funcMap     template.FuncMap
}

func mustLoadTemplates() *templates {
	funcMap := template.FuncMap{
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict: odd argument count")
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict: key must be string")
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
	}

	baseSources := make(map[string]string)

	walk := func(root string) error {
		return fs.WalkDir(monmsui.TemplatesFS, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			if !strings.HasSuffix(path, ".gohtml") {
				return nil
			}
			if strings.HasPrefix(path, "templates/pages/") {
				return nil
			}
			data, err := fs.ReadFile(monmsui.TemplatesFS, path)
			if err != nil {
				return err
			}
			baseSources[path] = string(data)
			return nil
		})
	}

	if err := walk("templates"); err != nil {
		slog.Error("monmsdash: load base templates", "err", err)
		panic(err)
	}

	return &templates{baseSources: baseSources, funcMap: funcMap}
}

func (t *templates) renderPage(w http.ResponseWriter, page string, data any) error {
	tmpl := template.New("base").Funcs(t.funcMap)

	for name, src := range t.baseSources {
		var err error
		tmpl, err = tmpl.New(name).Parse(src)
		if err != nil {
			return err
		}
	}

	pagePath := "templates/pages/" + page + ".gohtml"
	pageSrc, err := fs.ReadFile(monmsui.TemplatesFS, pagePath)
	if err != nil {
		return err
	}
	tmpl, err = tmpl.New(pagePath).Parse(string(pageSrc))
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", data)
}

func staticHandler() func(*core.RequestEvent) error {
	sub, err := fs.Sub(monmsui.StaticFS, "static")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	return func(e *core.RequestEvent) error {
		http.StripPrefix(staticPrefix, fileServer).ServeHTTP(e.Response, e.Request)
		return nil
	}
}
