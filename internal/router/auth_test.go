package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func TestLoadAuthFromCookie(t *testing.T) {
	dir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  dir,
		DefaultDev:      true,
		HideStartBanner: true,
	})

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatalf("find superusers: %v", err)
	}

	rec := core.NewRecord(superusers)
	rec.Set("email", "admin@test.local")
	rec.SetPassword("password123456")
	if err := app.Save(rec); err != nil {
		t.Fatalf("save superuser: %v", err)
	}

	token, err := rec.NewAuthToken()
	if err != nil {
		t.Fatalf("new auth token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	recorder := httptest.NewRecorder()

	event := &core.RequestEvent{
		App: app,
	}
	event.Request = req
	event.Response = recorder

	if err := LoadAuthFromCookie(event); err != nil {
		t.Fatalf("LoadAuthFromCookie: %v", err)
	}
	if event.Auth == nil {
		t.Fatal("expected e.Auth populated from cookie")
	}
	if event.Auth.Id != rec.Id {
		t.Fatalf("auth record id %q, want %q", event.Auth.Id, rec.Id)
	}
}

func TestEnrichSSRData_AuthTokenAndHero(t *testing.T) {
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
		t.Fatalf("import: %v", err)
	}

	collection, _ := app.FindCollectionByNameOrId(heroCollection)
	heroRec := core.NewRecord(collection)
	heroRec.Set("id", heroRecordID)
	heroRec.Set("title", "Test Hero")
	heroRec.Set("body", "Test body")
	if err := app.Save(heroRec); err != nil {
		t.Fatalf("save hero: %v", err)
	}

	superusers, _ := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	user := core.NewRecord(superusers)
	user.Set("email", "editor@test.local")
	user.SetPassword("password123456")
	if err := app.Save(user); err != nil {
		t.Fatalf("save user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	event := &core.RequestEvent{App: app, Auth: user}
	event.Request = req
	event.Response = recorder

	t.Run("index includes hero and auth token", func(t *testing.T) {
		data := enrichSSRData(event, "")
		if _, ok := data["AuthToken"]; !ok {
			t.Fatal("expected AuthToken when logged in")
		}
		hero, ok := data["Hero"].(map[string]any)
		if !ok {
			t.Fatal("expected Hero map on index")
		}
		if hero["Title"] != "Test Hero" {
			t.Fatalf("hero title %v, want Test Hero", hero["Title"])
		}
	})

	t.Run("non-index omits hero", func(t *testing.T) {
		data := enrichSSRData(event, "about")
		if _, ok := data["Hero"]; ok {
			t.Fatal("Hero should be omitted on non-index routes")
		}
	})

	t.Run("logged out omits auth token", func(t *testing.T) {
		guest := &core.RequestEvent{App: app}
		guest.Request = req
		guest.Response = recorder
		data := enrichSSRData(guest, "")
		if _, ok := data["AuthToken"]; ok {
			t.Fatal("AuthToken should be absent when logged out")
		}
	})
}

func TestEnrichSSRData_HeroFallback(t *testing.T) {
	dir := t.TempDir()
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  dir,
		DefaultDev:      true,
		HideStartBanner: true,
	})
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	event := &core.RequestEvent{App: app}
	event.Request = req
	event.Response = recorder

	data := enrichSSRData(event, "")
	hero, ok := data["Hero"].(map[string]any)
	if !ok {
		t.Fatal("expected Hero map with fallback")
	}
	if hero["Title"] != "MonMS is running" {
		t.Fatalf("fallback title %v", hero["Title"])
	}
}
