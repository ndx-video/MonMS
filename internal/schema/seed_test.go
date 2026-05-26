package schema

import (
	"testing"

	"github.com/pocketbase/pocketbase"
)

func TestSeedHeroHomepage(t *testing.T) {
	dir := t.TempDir()

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  dir,
		DefaultDev:      true,
		HideStartBanner: true,
	})

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	schemaJSON := []byte(`[{
		"name": "hero_content",
		"type": "base",
		"fields": [
			{"name": "id", "type": "text", "required": true, "primaryKey": true, "system": true, "min": 1, "max": 50, "pattern": "^[a-z][a-z0-9_]*$"},
			{"name": "title", "type": "text"},
			{"name": "body", "type": "text"}
		]
	}]`)

	if err := app.ImportCollectionsByMarshaledJSON(schemaJSON, false); err != nil {
		t.Fatalf("import collections: %v", err)
	}

	if err := seedHeroHomepage(app); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	rec, err := app.FindRecordById(heroCollection, heroRecordID)
	if err != nil {
		t.Fatalf("find homepage: %v", err)
	}
	if got := rec.GetString("title"); got != heroSeedTitle {
		t.Fatalf("title %q, want %q", got, heroSeedTitle)
	}
	if got := rec.GetString("body"); got != heroSeedBody {
		t.Fatalf("body %q, want %q", got, heroSeedBody)
	}

	if err := seedHeroHomepage(app); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	list, err := app.FindRecordsByFilter(heroCollection, "id = {:id}", "", 0, 0, map[string]any{"id": heroRecordID})
	if err != nil {
		t.Fatalf("list records: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 homepage record, got %d", len(list))
	}
}

func TestSeedHeroHomepage_NoCollection(t *testing.T) {
	dir := t.TempDir()

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  dir,
		DefaultDev:      true,
		HideStartBanner: true,
	})

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	if err := seedHeroHomepage(app); err != nil {
		t.Fatalf("seed without collection: %v", err)
	}
}
