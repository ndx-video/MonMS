package content

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/testutil"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func bootstrapEditorialApp(t *testing.T, ws string) core.App {
	t.Helper()

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(ws, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})
	schema.RegisterBootstrapHook(app, ws)
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	return app
}

func TestExportAllWritesHeroContent(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	app := bootstrapEditorialApp(t, ws)

	if err := ExportAll(app, ws); err != nil {
		t.Fatalf("ExportAll: %v", err)
	}

	path := filepath.Join(ws, "content", "hero_content.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}

	var file CollectionFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if file.Collection != "hero_content" {
		t.Fatalf("collection %q, want hero_content", file.Collection)
	}
	if len(file.Records) != 1 {
		t.Fatalf("records len %d, want 1", len(file.Records))
	}
	if id, _ := file.Records[0]["id"].(string); id != "homepage" {
		t.Fatalf("id %q, want homepage", id)
	}
}

func TestImportFilesIdempotent(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	app := bootstrapEditorialApp(t, ws)

	if err := ExportAll(app, ws); err != nil {
		t.Fatalf("export: %v", err)
	}

	updated := filepath.Join(ws, "content", "hero_content.json")
	data, err := os.ReadFile(updated)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var file CollectionFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	file.Records[0]["title"] = "Updated Title"
	merged, _ := json.Marshal(file)
	if err := os.WriteFile(updated, merged, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := ImportFiles(app, ws); err != nil {
		t.Fatalf("first import: %v", err)
	}
	if err := ImportFiles(app, ws); err != nil {
		t.Fatalf("second import: %v", err)
	}

	list, err := app.FindRecordsByFilter("hero_content", "id = {:id}", "", 0, 0, map[string]any{"id": "homepage"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 record, got %d", len(list))
	}
	if got := list[0].GetString("title"); got != "Updated Title" {
		t.Fatalf("title %q, want Updated Title", got)
	}
}

func TestImportPayloadRejectsNonEditorial(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	app := bootstrapEditorialApp(t, ws)

	err := ImportPayload(app, ws, []CollectionPayload{{
		Collection: "users",
		Records:    []map[string]any{{"id": "u1"}},
	}})
	if err == nil || !strings.Contains(err.Error(), "not editorial") {
		t.Fatalf("ImportPayload err = %v, want non-editorial error", err)
	}
}
