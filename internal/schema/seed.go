package schema

import (
	"github.com/monms/monms/internal/logging"
	"github.com/pocketbase/pocketbase/core"
)

const (
	heroCollection = "hero_content"
	heroRecordID   = "homepage"
)

const (
	heroSeedTitle = "Welcome to MonMS"
	heroSeedBody  = "This headline and paragraph are stored in the hero_content collection. Sign in via Admin, then click here to edit in place — changes save when you click away."
)

// seedHeroHomepage creates the homepage hero record idempotently after schema import.
func seedHeroHomepage(app core.App) error {
	if _, err := app.FindCollectionByNameOrId(heroCollection); err != nil {
		return nil
	}

	if _, err := app.FindRecordById(heroCollection, heroRecordID); err == nil {
		return nil
	}

	collection, err := app.FindCollectionByNameOrId(heroCollection)
	if err != nil {
		return nil
	}

	record := core.NewRecord(collection)
	record.Set("id", heroRecordID)
	record.Set("title", heroSeedTitle)
	record.Set("body", heroSeedBody)

	if err := app.Save(record); err != nil {
		return err
	}

	logging.Schema("schema seed: created hero homepage record", "id", heroRecordID)
	return nil
}
