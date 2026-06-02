package documents

import (
	"fmt"
	"log/slog"

	"github.com/pocketbase/pocketbase/core"
)

// upsertRecord creates or updates a record by stable ID (same semantics as content.UpsertRecord).
func upsertRecord(app core.App, collectionName string, data map[string]any) error {
	id, _ := data["id"].(string)
	if id == "" {
		return fmt.Errorf("documents sync: record missing id")
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
			slog.Warn("documents sync: unknown field skipped",
				"collection", collectionName, "field", k)
			continue
		}
		rec.Set(k, v)
	}

	if err := app.Save(rec); err != nil {
		return fmt.Errorf("documents sync: save %s/%s: %w", collectionName, id, err)
	}
	slog.Info("documents sync: upserted record", "collection", collectionName, "id", id)
	return nil
}
