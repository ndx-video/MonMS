package documents

import (
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase/core"
)

// DocumentMatch is a markdown-backed record resolved from a URL slug.
type DocumentMatch struct {
	Collection string
	Record     *core.Record
	Binding    schema.CollectionMeta
}

// FindBySlug looks up a markdown document record by URL path slug.
// slug is the path after the leading slash (e.g. "guides/setup").
func FindBySlug(app core.App, siteAbs, slug string) (*DocumentMatch, error) {
	slug = strings.Trim(slug, "/")
	if slug == "" {
		return nil, nil
	}

	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return nil, err
	}

	for _, binding := range bindings {
		rec, err := findInCollection(app, binding.Name, slug)
		if err != nil {
			return nil, err
		}
		if rec != nil {
			return &DocumentMatch{
				Collection: binding.Name,
				Record:     rec,
				Binding:    binding,
			}, nil
		}
	}
	return nil, nil
}

func findInCollection(app core.App, collection, slug string) (*core.Record, error) {
	records, err := app.FindAllRecords(collection)
	if err != nil {
		return nil, err
	}
	for _, rec := range records {
		if rec.GetString("slug") == slug {
			return rec, nil
		}
	}
	return nil, nil
}

// MarkdownCollectionNames returns names of markdown-backed collections.
func MarkdownCollectionNames(siteAbs string) ([]string, error) {
	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return nil, err
	}
	names := make([]string, len(bindings))
	for i, b := range bindings {
		names[i] = b.Name
	}
	return names, nil
}
