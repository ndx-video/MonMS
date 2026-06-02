package apikeys_test

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/apikeys"
	"github.com/monms/monms/internal/authbootstrap"
	"github.com/monms/monms/internal/content"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func bootstrapApp(t *testing.T) core.App {
	t.Helper()
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
	return app
}

func TestGenerateResolveRevoke(t *testing.T) {
	app := bootstrapApp(t)
	siteAbs := t.TempDir()

	superCol, _ := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	owner := core.NewRecord(superCol)
	owner.Set("email", "keys@test.local")
	owner.SetPassword("password123456")
	if err := app.Save(owner); err != nil {
		t.Fatalf("save superuser: %v", err)
	}

	secret, rec, err := apikeys.Create(app, siteAbs, owner, "agent")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if secret == "" || rec == nil {
		t.Fatal("expected secret and record")
	}

	resolved, err := apikeys.Resolve(app, siteAbs, secret)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resolved.Owner.Id != owner.Id {
		t.Fatalf("owner id %q, want %q", resolved.Owner.Id, owner.Id)
	}

	list, err := apikeys.ListForOwner(app, owner)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v len=%d", err, len(list))
	}

	if err := apikeys.Revoke(app, owner, rec.Id); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := apikeys.Resolve(app, siteAbs, secret); err == nil {
		t.Fatal("expected resolve failure after revoke")
	}
}

func TestCanManageKeys(t *testing.T) {
	app := bootstrapApp(t)
	cfg := content.MonmsConfig{MCP: content.MCPConfig{AllowNonSuperuserKeys: true}}

	superCol, _ := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	super := core.NewRecord(superCol)
	if !apikeys.CanManageKeys(super, content.MonmsConfig{}) {
		t.Fatal("superuser should manage keys")
	}

	usersCol, _ := app.FindCollectionByNameOrId(apikeys.CollectionUsers)
	user := core.NewRecord(usersCol)
	if !apikeys.CanManageKeys(user, cfg) {
		t.Fatal("user should manage keys when allowed")
	}
	if apikeys.CanManageKeys(user, content.MonmsConfig{}) {
		t.Fatal("user should not manage keys when disallowed")
	}
}
