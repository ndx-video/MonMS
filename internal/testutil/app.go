package testutil

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/authbootstrap"
	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// BootstrapApp starts PocketBase with schema bootstrap only.
func BootstrapApp(t *testing.T, siteAbs string) core.App {
	t.Helper()

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(siteAbs, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})
	authbootstrap.RegisterBootstrapHook(app)
	schema.RegisterBootstrapHook(app, siteAbs)
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	return app
}

// SaveRecord creates or updates a record in collection with the given field map.
func SaveRecord(t *testing.T, app core.App, collection string, data map[string]any) *core.Record {
	t.Helper()

	id, _ := data["id"].(string)
	coll, err := app.FindCollectionByNameOrId(collection)
	if err != nil {
		t.Fatalf("find collection %s: %v", collection, err)
	}

	var rec *core.Record
	if id != "" {
		rec, err = app.FindRecordById(collection, id)
	}
	if err != nil || rec == nil {
		rec = core.NewRecord(coll)
		if id != "" {
			rec.Set("id", id)
		}
	}

	for k, v := range data {
		if k == "id" {
			continue
		}
		rec.Set(k, v)
	}
	if err := app.Save(rec); err != nil {
		t.Fatalf("save record: %v", err)
	}
	return rec
}
