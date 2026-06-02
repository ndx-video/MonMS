package mcp_test

import (
	"testing"

	"github.com/monms/monms/internal/apikeys"
	"github.com/monms/monms/internal/authbootstrap"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func TestResolveKeyForSuperuser(t *testing.T) {
	dir := t.TempDir()
	siteAbs := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  dir,
		DefaultDev:      true,
		HideStartBanner: true,
	})
	authbootstrap.RegisterBootstrapHook(app)
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	superCol, _ := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	owner := core.NewRecord(superCol)
	owner.Set("email", "mcp@test.local")
	owner.SetPassword("password123456")
	if err := app.Save(owner); err != nil {
		t.Fatalf("save: %v", err)
	}

	secret, _, err := apikeys.Create(app, siteAbs, owner, "mcp-test")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}
	resolved, err := apikeys.Resolve(app, siteAbs, secret)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resolved.Owner.Id != owner.Id {
		t.Fatalf("owner mismatch")
	}
}
