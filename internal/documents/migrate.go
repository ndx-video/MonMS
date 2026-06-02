package documents

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ScanEntry describes one markdown file found during legacy inventory.
type ScanEntry struct {
	Path         string
	RelPath      string
	HasFrontmatter bool
	Meta         map[string]any
	Title        string
}

// ScanTree walks a directory tree and inventories markdown files.
func ScanTree(root string) ([]ScanEntry, error) {
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("documents scan: not a directory: %s", root)
	}

	var entries []ScanEntry
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		doc, err := ParseFile(path)
		entry := ScanEntry{
			Path:    path,
			RelPath: rel,
			Meta:    doc.Meta,
		}
		if err == nil {
			entry.HasFrontmatter = len(doc.Meta) > 0
			if t, ok := doc.Meta["title"].(string); ok {
				entry.Title = t
			}
		}
		if entry.Title == "" {
			base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			entry.Title = strings.ReplaceAll(base, "-", " ")
		}
		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RelPath < entries[j].RelPath
	})
	return entries, nil
}

// BindPlan describes how legacy files map into MonMS documents/.
type BindPlan struct {
	Collection string            `yaml:"collection"`
	SourceRoot string            `yaml:"sourceRoot"`
	DestRoot   string            `yaml:"destRoot"`
	FieldMap   map[string]string `yaml:"fieldMap"`
}

// DefaultPlanFromScan proposes bindings from a top-level folder name per collection.
func DefaultPlanFromScan(sourceRoot string, entries []ScanEntry) []BindPlan {
	if len(entries) == 0 {
		return nil
	}

	byTop := make(map[string]int)
	for _, e := range entries {
		parts := strings.Split(e.RelPath, "/")
		top := parts[0]
		if len(parts) == 1 {
			top = filepath.Base(strings.TrimSuffix(sourceRoot, string(filepath.Separator)))
		}
		byTop[top]++
	}

	if len(byTop) == 1 {
		for top := range byTop {
			collection := singularize(top)
			return []BindPlan{{
				Collection: collection,
				SourceRoot: sourceRoot,
				DestRoot:   "documents/" + collection,
				FieldMap:   map[string]string{"date": "published_at"},
			}}
		}
	}

	var tops []string
	for top := range byTop {
		tops = append(tops, top)
	}
	sort.Strings(tops)

	var plans []BindPlan
	for _, top := range tops {
		collection := singularize(top)
		plans = append(plans, BindPlan{
			Collection: collection,
			SourceRoot: filepath.Join(sourceRoot, top),
			DestRoot:   "documents/" + collection,
			FieldMap:   map[string]string{"date": "published_at"},
		})
	}
	return plans
}

func singularize(name string) string {
	name = strings.TrimSuffix(name, "/")
	if strings.HasSuffix(name, "ies") {
		return strings.TrimSuffix(name, "ies") + "y"
	}
	if strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") {
		return strings.TrimSuffix(name, "s")
	}
	return name
}

// ApplyBind copies files per plan and injects frontmatter.
func ApplyBind(siteAbs string, plan BindPlan, dryRun, force bool) (int, error) {
	entries, err := ScanTree(plan.SourceRoot)
	if err != nil {
		return 0, err
	}

	destRoot := filepath.Join(siteAbs, filepath.FromSlash(plan.DestRoot))
	var bound int

	for _, entry := range entries {
		destPath := filepath.Join(destRoot, entry.RelPath)
		pathKey := strings.TrimSuffix(filepath.ToSlash(entry.RelPath), ".md")

		if dryRun {
			meta, _ := mergeFrontmatterMeta(entry.Meta, plan.Collection, pathKey, entry.Title, force)
			fmt.Printf("  would bind %s -> %s (id=%s)\n", entry.RelPath, destPath, meta["id"])
			bound++
			continue
		}

		doc, err := ParseFile(entry.Path)
		if err != nil {
			return bound, err
		}
		meta, changed := mergeFrontmatterMeta(doc.Meta, plan.Collection, pathKey, entry.Title, force)
		if !changed && frontmatterEqual(doc.Meta, meta) {
			if err := copyFile(entry.Path, destPath); err != nil {
				return bound, err
			}
			bound++
			continue
		}
		if err := WriteFile(destPath, meta, doc.Body); err != nil {
			return bound, err
		}
		bound++
	}

	return bound, nil
}

// DefaultArticlesSchema returns a schema JSON stub for a markdown collection.
func DefaultArticlesSchema(name, root string, fieldMap map[string]string) string {
	fields := `[
    { "name": "id", "type": "text", "required": true, "primaryKey": true, "pattern": "^[a-z][a-z0-9_-]*$", "min": 1, "max": 120 },
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "path", "type": "text" },
    { "name": "folder", "type": "text" },
    { "name": "body", "type": "text" },
    { "name": "status", "type": "select", "values": ["draft", "published"] },
    { "name": "published_at", "type": "text" }
  ]`
	fmMap := ""
	if len(fieldMap) > 0 {
		parts := make([]string, 0, len(fieldMap))
		for k, v := range fieldMap {
			parts = append(parts, fmt.Sprintf("%q: %q", k, v))
		}
		sort.Strings(parts)
		fmMap = ",\n    \"fields\": { " + strings.Join(parts, ", ") + " }"
	}
	return fmt.Sprintf(`{
  "name": %q,
  "type": "base",
  "editorial": true,
  "monms": {
    "source": "markdown",
    "root": %q,
    "slugFrom": "path",
    "idFrom": "frontmatter.id"%s
  },
  "listRule": "",
  "viewRule": "",
  "fields": %s
}`, name, root, fmMap, fields)
}

// DefaultDoctreeFieldMap returns frontmatter → PocketBase field mappings for dt_* collections.
func DefaultDoctreeFieldMap() map[string]string {
	return map[string]string{
		"title":       "title",
		"description": "description",
		"ts_create":   "ts_create",
		"ts_mod":      "ts_mod",
		"date":        "published_at",
	}
}

// DefaultDoctreeCollectionSchema returns schema JSON for a dt_* markdown leaf collection.
func DefaultDoctreeCollectionSchema(name, root, doctreeID string, fieldMap map[string]string) string {
	if fieldMap == nil {
		fieldMap = DefaultDoctreeFieldMap()
	}
	fields := `[
    { "name": "id", "type": "text", "required": true, "primaryKey": true, "pattern": "^[a-z][a-z0-9_-]*$", "min": 1, "max": 120 },
    { "name": "title", "type": "text" },
    { "name": "description", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "path", "type": "text" },
    { "name": "folder", "type": "text" },
    { "name": "body", "type": "text" },
    { "name": "doctree_id", "type": "text" },
    { "name": "leaf_path", "type": "text" },
    { "name": "monms_sync_at", "type": "text" },
    { "name": "ts_create", "type": "text" },
    { "name": "ts_mod", "type": "text" },
    { "name": "status", "type": "select", "values": ["draft", "published"] },
    { "name": "published_at", "type": "text" }
  ]`
	fmMap := ""
	if len(fieldMap) > 0 {
		parts := make([]string, 0, len(fieldMap))
		for k, v := range fieldMap {
			parts = append(parts, fmt.Sprintf("%q: %q", k, v))
		}
		sort.Strings(parts)
		fmMap = ",\n    \"fields\": { " + strings.Join(parts, ", ") + " }"
	}
	return fmt.Sprintf(`{
  "name": %q,
  "type": "base",
  "editorial": true,
  "monms": {
    "source": "markdown",
    "root": %q,
    "doctree": %q,
    "slugFrom": "path",
    "idFrom": "frontmatter.id"%s
  },
  "listRule": "",
  "viewRule": "",
  "fields": %s
}`, name, root, doctreeID, fmMap, fields)
}
