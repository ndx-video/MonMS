package authbootstrap_test

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/authbootstrap"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func TestBootstrapCreatesCollections(t *testing.T) {
	dir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(dir, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})
	authbootstrap.RegisterBootstrapHook(app)
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	if _, err := app.FindCollectionByNameOrId(authbootstrap.CollectionUsers); err != nil {
		t.Fatalf("users collection: %v", err)
	}
	if _, err := app.FindCollectionByNameOrId(authbootstrap.CollectionAPIKeys); err != nil {
		t.Fatalf("api keys collection: %v", err)
	}
}

func TestAPIKeyRequiresSingleOwner(t *testing.T) {
	dir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(dir, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})
	authbootstrap.RegisterBootstrapHook(app)
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	superCol, _ := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	super := core.NewRecord(superCol)
	super.Set("email", "a@test.local")
	super.SetPassword("password123456")
	if err := app.Save(super); err != nil {
		t.Fatalf("save superuser: %v", err)
	}

	keyCol, _ := app.FindCollectionByNameOrId(authbootstrap.CollectionAPIKeys)
	rec := core.NewRecord(keyCol)
	rec.Set("name", "bad")
	rec.Set("prefix", "monms_ab")
	rec.Set("secretHash", "00")
	rec.Set("superuser", super.Id)
	rec.Set("user", super.Id)
	if err := app.Save(rec); err == nil {
		t.Fatal("expected validation error for dual owner")
	}
}
