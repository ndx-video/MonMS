package documents

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/schema"
)

// DocumentEntry summarizes one markdown file under a collection root.
type DocumentEntry struct {
	PathKey string         `json:"path"`
	RelFile string         `json:"relFile"`
	Title   string         `json:"title"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// FindBinding returns the markdown binding for a collection name.
func FindBinding(siteAbs, collection string) (schema.CollectionMeta, error) {
	bindings, err := schema.LoadMarkdownBindings(SchemaDir(siteAbs))
	if err != nil {
		return schema.CollectionMeta{}, err
	}
	for _, b := range bindings {
		if b.Name == collection {
			return b, nil
		}
	}
	return schema.CollectionMeta{}, fmt.Errorf("documents: collection %q is not markdown-bound", collection)
}

// DocFilePath returns the absolute filesystem path for a document path key.
// pathKey is slash-separated without extension (e.g. "guides/setup").
func DocFilePath(siteAbs string, binding schema.CollectionMeta, pathKey string) string {
	pathKey = strings.Trim(pathKey, "/")
	rel := pathKey
	if !strings.HasSuffix(strings.ToLower(rel), ".md") {
		rel += ".md"
	}
	return filepath.Join(siteAbs, filepath.FromSlash(binding.Monms.Root), filepath.FromSlash(rel))
}

// ListDocuments walks the bound root and returns parsed document summaries.
func ListDocuments(siteAbs string, binding schema.CollectionMeta) ([]DocumentEntry, error) {
	root := filepath.Join(siteAbs, filepath.FromSlash(binding.Monms.Root))
	docs, err := WalkMarkdownFiles(root)
	if err != nil {
		return nil, err
	}

	out := make([]DocumentEntry, 0, len(docs))
	for _, doc := range docs {
		rel, err := filepath.Rel(root, doc.FilePath)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		pathKey := strings.TrimSuffix(rel, filepath.Ext(rel))
		title := ""
		if t, ok := doc.Meta["title"].(string); ok {
			title = t
		}
		if title == "" {
			base := filepath.Base(pathKey)
			title = strings.ReplaceAll(base, "-", " ")
		}
		out = append(out, DocumentEntry{
			PathKey: pathKey,
			RelFile: rel,
			Title:   title,
			Meta:    doc.Meta,
		})
	}
	return out, nil
}

// ReadDocument loads a markdown file by collection and path key.
func ReadDocument(siteAbs, collection, pathKey string) (schema.CollectionMeta, ParsedDocument, error) {
	binding, err := FindBinding(siteAbs, collection)
	if err != nil {
		return schema.CollectionMeta{}, ParsedDocument{}, err
	}
	path := DocFilePath(siteAbs, binding, pathKey)
	doc, err := ParseFile(path)
	if err != nil {
		return binding, ParsedDocument{}, err
	}
	return binding, doc, nil
}
