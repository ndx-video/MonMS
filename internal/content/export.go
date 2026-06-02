package content

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// CollectionFile is the on-disk export shape per specs/staging.md §5.1 (PUB-02).
type CollectionFile struct {
	Collection string           `json:"collection"`
	Records    []map[string]any `json:"records"`
}

// ExportCollection exports one editorial collection to a map slice.
func ExportCollection(app core.App, name string) ([]map[string]any, error) {
	coll, err := app.FindCollectionByNameOrId(name)
	if err != nil {
		return nil, fmt.Errorf("content export: collection %s: %w", name, err)
	}

	records, err := app.FindAllRecords(name)
	if err != nil {
		return nil, fmt.Errorf("content export: list %s: %w", name, err)
	}

	out := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		m := rec.PublicExport()
		delete(m, "collectionId")
		delete(m, "collectionName")
		delete(m, "expand")

		for k := range m {
			if k == "id" {
				continue
			}
			field := coll.Fields.GetByName(k)
			if field == nil {
				continue
			}
			if field.Type() == core.FieldTypeFile {
				delete(m, k)
				slog.Warn("content export: file field skipped",
					"collection", name, "field", k)
			}
		}
		out = append(out, m)
	}

	sort.Slice(out, func(i, j int) bool {
		idi, _ := out[i]["id"].(string)
		idj, _ := out[j]["id"].(string)
		return idi < idj
	})

	return out, nil
}

// ExportAll writes site/content/{collection}.json for each PB-native editorial collection (PUB-02).
func ExportAll(app core.App, siteAbs string) error {
	names, err := LoadPBNativeEditorialCollectionNames(siteAbs)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}

	contentDir := filepath.Join(siteAbs, "content")
	if err := os.MkdirAll(contentDir, 0o755); err != nil {
		return fmt.Errorf("content export: mkdir content: %w", err)
	}

	sort.Strings(names)
	for _, name := range names {
		records, err := ExportCollection(app, name)
		if err != nil {
			return err
		}

		payload := CollectionFile{Collection: name, Records: records}
		merged, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("content export: marshal %s: %w", name, err)
		}

		dest := filepath.Join(contentDir, name+".json")
		if err := ensureUnderSite(siteAbs, dest); err != nil {
			return err
		}
		if err := os.WriteFile(dest, merged, 0o644); err != nil {
			return fmt.Errorf("content export: write %s: %w", dest, err)
		}
		slog.Info("content export: wrote file", "collection", name, "records", len(records))
	}
	return nil
}

// ExportSnapshot builds in-memory editorial payloads for checksum/diff (sorted).
// Markdown-backed collections are excluded — they promote via Git structure rail.
func ExportSnapshot(app core.App, siteAbs string) ([]CollectionFile, error) {
	names, err := LoadPBNativeEditorialCollectionNames(siteAbs)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)

	var files []CollectionFile
	for _, name := range names {
		records, err := ExportCollection(app, name)
		if err != nil {
			return nil, err
		}
		files = append(files, CollectionFile{Collection: name, Records: records})
	}
	return files, nil
}

// LoadContentFiles reads site/content/*.json as baseline snapshots.
func LoadContentFiles(siteAbs string) ([]CollectionFile, error) {
	contentDir := filepath.Join(siteAbs, "content")
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []CollectionFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(contentDir, entry.Name())
		if err := ensureUnderSite(siteAbs, path); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("content export: read %s: %w", path, err)
		}
		var file CollectionFile
		if err := json.Unmarshal(data, &file); err != nil {
			return nil, fmt.Errorf("content export: parse %s: %w", path, err)
		}
		if file.Collection != "" {
			files = append(files, file)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Collection < files[j].Collection
	})
	return files, nil
}
