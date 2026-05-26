package content

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/testutil"
	"github.com/pocketbase/pocketbase/core"
)

// TestExportSkipsFileFields verifies MED-01: file blobs are not serialized;
// publishable media uses CDN URLs in text fields (D-55).
func TestExportSkipsFileFields(t *testing.T) {
	ws := testutil.NewWorkspace(t)
	schemaJSON := `{
  "name": "media_items",
  "type": "base",
  "editorial": true,
  "fields": [
    {"name": "id", "type": "text", "required": true, "primaryKey": true, "system": true, "min": 1, "max": 50, "pattern": "^[a-z][a-z0-9_]*$"},
    {"name": "title", "type": "text"},
    {"name": "asset", "type": "file", "maxSelect": 1, "maxSize": 5242880},
    {"name": "cdn_url", "type": "text"}
  ]
}`
	testutil.WriteFile(t, filepath.Join(ws, "schema/media_items.json"), schemaJSON)

	app := bootstrapEditorialApp(t, ws)
	coll, err := app.FindCollectionByNameOrId("media_items")
	if err != nil {
		t.Fatalf("collection: %v", err)
	}
	rec := core.NewRecord(coll)
	rec.Set("id", "item1")
	rec.Set("title", "Logo")
	rec.Set("cdn_url", "https://cdn.example.com/logo.png")
	if err := app.Save(rec); err != nil {
		t.Fatalf("save: %v", err)
	}

	records, err := ExportCollection(app, "media_items")
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records %d", len(records))
	}
	m := records[0]
	if _, ok := m["asset"]; ok {
		t.Fatal("export must not contain file field asset")
	}
	if got, _ := m["cdn_url"].(string); got != "https://cdn.example.com/logo.png" {
		t.Fatalf("cdn_url %q", got)
	}
	if got, _ := m["title"].(string); got != "Logo" {
		t.Fatalf("title %q", got)
	}
}
