package content

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// CollectionPayload is one collection's records for HTTP import (plan 04-04).
type CollectionPayload struct {
	Collection string           `json:"collection"`
	Records    []map[string]any `json:"records"`
}

// UpsertRecord creates or updates a record by stable ID (PUB-04).
func UpsertRecord(app core.App, collectionName string, data map[string]any) error {
	id, _ := data["id"].(string)
	if id == "" {
		return fmt.Errorf("content import: record missing id")
	}

	coll, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return err
	}

	rec, err := app.FindRecordById(collectionName, id)
	if err != nil {
		rec = core.NewRecord(coll)
		rec.Set("id", id)
	}

	for k, v := range data {
		if k == "id" || k == "collectionId" || k == "collectionName" {
			continue
		}
		if coll.Fields.GetByName(k) == nil {
			slog.Warn("content import: unknown field skipped",
				"collection", collectionName, "field", k)
			continue
		}
		rec.Set(k, v)
	}

	if err := app.Save(rec); err != nil {
		return fmt.Errorf("content import: save %s/%s: %w", collectionName, id, err)
	}
	slog.Info("content import: upserted record", "collection", collectionName, "id", id)
	return nil
}

// ImportPayload upserts editorial collections from an in-memory payload (T-04-04).
func ImportPayload(app core.App, wsAbs string, payloads []CollectionPayload) error {
	allowed, err := LoadEditorialCollectionNames(wsAbs)
	if err != nil {
		return err
	}
	allowSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowSet[name] = struct{}{}
	}

	for _, p := range payloads {
		if p.Collection == "" {
			return fmt.Errorf("content import: missing collection name")
		}
		if _, ok := allowSet[p.Collection]; !ok {
			return fmt.Errorf("content import: collection %q is not editorial", p.Collection)
		}
		for _, rec := range p.Records {
			if err := UpsertRecord(app, p.Collection, rec); err != nil {
				return err
			}
		}
	}
	return nil
}

// ImportFiles reads workspace/content/*.json and upserts records (PUB-04).
func ImportFiles(app core.App, wsAbs string) error {
	contentDir := filepath.Join(wsAbs, "content")
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("content import: read content dir: %w", err)
	}

	var payloads []CollectionPayload
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(contentDir, entry.Name())
		if err := ensureUnderWorkspace(wsAbs, path); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("content import: read %s: %w", path, err)
		}

		var file CollectionFile
		if err := json.Unmarshal(data, &file); err != nil {
			return fmt.Errorf("content import: parse %s: %w", path, err)
		}
		if file.Collection == "" {
			return fmt.Errorf("content import: %s missing collection", path)
		}
		payloads = append(payloads, CollectionPayload{
			Collection: file.Collection,
			Records:    file.Records,
		})
	}

	return ImportPayload(app, wsAbs, payloads)
}
