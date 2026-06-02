package documents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/schema"
)

const defaultBodyField = "body"

// RecordFromDocument maps a parsed markdown file to a PocketBase record payload.
func RecordFromDocument(binding schema.CollectionMeta, doc ParsedDocument, rootAbs string) (map[string]any, error) {
	if binding.Monms.Root == "" {
		return nil, fmt.Errorf("documents: collection %q missing monms.root", binding.Name)
	}

	rel, err := filepath.Rel(rootAbs, doc.FilePath)
	if err != nil {
		return nil, fmt.Errorf("documents: rel path %s: %w", doc.FilePath, err)
	}
	rel = filepath.ToSlash(rel)
	doc.RelPath = rel

	pathKey := strings.TrimSuffix(rel, filepath.Ext(rel))
	slug := pathKey
	if binding.Monms.SlugFrom == "frontmatter.slug" {
		if s, ok := doc.Meta["slug"].(string); ok && s != "" {
			slug = s
		}
	}

	id := ""
	if binding.Monms.IDFrom == "frontmatter.id" {
		if v, ok := doc.Meta["id"].(string); ok {
			id = v
		}
	}
	if id == "" {
		if v, ok := doc.Meta["id"].(string); ok && v != "" {
			id = v
		}
	}
	if id == "" {
		id = recordID(binding.Name, pathKey)
	} else {
		id = sanitizeRecordID(id)
	}

	bodyField := binding.Monms.Body
	if bodyField == "" {
		bodyField = defaultBodyField
	}

	record := map[string]any{
		"id":   id,
		"slug": slug,
		"path": pathKey,
		bodyField: doc.Body,
	}

	if pathKey != "" {
		if i := strings.Index(pathKey, "/"); i >= 0 {
			record["folder"] = pathKey[:i]
		}
	}

	fieldMap := binding.Monms.Fields
	for fmKey, value := range doc.Meta {
		if fmKey == "id" || fmKey == "slug" {
			continue
		}
		pbField := fmKey
		if fieldMap != nil {
			if mapped, ok := fieldMap[fmKey]; ok {
				pbField = mapped
			}
		}
		if pbField == "id" || pbField == "slug" || pbField == "path" || pbField == "folder" {
			continue
		}
		record[pbField] = normalizeFMValue(value)
	}

	return record, nil
}

// StableRecordID builds a PocketBase-safe stable id from collection and path key.
func StableRecordID(collection, pathKey string) string {
	return recordID(collection, pathKey)
}

// recordID builds a PocketBase-safe stable id (slashes are not allowed in PB primary keys).
func recordID(collection, pathKey string) string {
	return collection + "--" + strings.ReplaceAll(pathKey, "/", "--")
}

func sanitizeRecordID(id string) string {
	return strings.ReplaceAll(id, "/", "--")
}

func normalizeFMValue(v any) any {
	switch val := v.(type) {
	case []any:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ",")
	default:
		return v
	}
}

// WalkMarkdownFiles returns parsed documents under root (non-recursive skip of hidden dirs).
func WalkMarkdownFiles(root string) ([]ParsedDocument, error) {
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("documents: root is not a directory: %s", root)
	}

	var docs []ParsedDocument
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "_media" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		doc, err := ParseFile(path)
		if err != nil {
			return err
		}
		docs = append(docs, doc)
		return nil
	})
	return docs, err
}
