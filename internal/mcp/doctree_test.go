package mcp

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/monms/monms/internal/apikeys"
	"github.com/monms/monms/internal/authbootstrap"
	"github.com/monms/monms/internal/documents"
	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/testutil"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const testArticlesSchema = `{
  "name": "articles",
  "type": "base",
  "editorial": true,
  "monms": {
    "source": "markdown",
    "root": "documents/articles"
  },
  "fields": [
    { "name": "id", "type": "text", "primaryKey": true, "required": true, "pattern": "^[a-z][a-z0-9_-]*$", "min": 1, "max": 120 },
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "path", "type": "text" },
    { "name": "folder", "type": "text" },
    { "name": "body", "type": "text" }
  ]
}`

func testDoctreeDeps(t *testing.T) (Deps, context.Context) {
	t.Helper()
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, "schema/articles.json"), testArticlesSchema)
	testutil.WriteFile(t, filepath.Join(ws, "documents/articles/guides/setup.md"), `---
id: articles--guides--setup
title: Setup Guide
---
## Intro

Hello world.
`)

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(ws, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})
	authbootstrap.RegisterBootstrapHook(app)
	documents.RegisterBootstrapHook(app, ws)
	schema.RegisterBootstrapHook(app, ws)
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	superCol, _ := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	owner := core.NewRecord(superCol)
	owner.Set("email", "doctree@test.local")
	owner.SetPassword("password123456")
	if err := app.Save(owner); err != nil {
		t.Fatalf("save owner: %v", err)
	}

	siteAbs := ws
	secret, _, err := apikeys.Create(app, siteAbs, owner, "doctree-test")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}
	resolved, err := apikeys.Resolve(app, siteAbs, secret)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	token, err := resolved.Owner.NewAuthToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	deps := Deps{App: app, SiteAbs: ws}
	ctx := withSession(context.Background(), &session{Resolved: resolved, OwnerToken: token})
	return deps, ctx
}

func TestDoctreeBindings(t *testing.T) {
	deps, ctx := testDoctreeDeps(t)
	res, err := doctreeBindingsHandler(deps)(ctx, mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if res.IsError {
		t.Fatalf("error result: %+v", res)
	}
	if len(res.Content) == 0 || !containsText(res, "articles") {
		t.Fatalf("expected articles binding, got %+v", res)
	}
}

func TestDoctreeForest(t *testing.T) {
	deps, ctx := testDoctreeDeps(t)
	res, err := doctreeForestHandler(deps)(ctx, mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if res.IsError {
		t.Fatalf("error result: %+v", res)
	}
	if !containsText(res, "guides/setup") {
		t.Fatalf("expected guides/setup in forest, got %+v", res)
	}
}

func TestDoctreeWriteAndGet(t *testing.T) {
	deps, ctx := testDoctreeDeps(t)

	writeReq := mcp.CallToolRequest{}
	writeReq.Params.Arguments = map[string]any{
		"collection": "articles",
		"path":       "guides/new-page",
		"body":       "# New\n\nContent here.",
		"meta": map[string]any{
			"title": "New Page",
		},
	}
	res, err := doctreeWriteHandler(deps)(ctx, writeReq)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if res.IsError {
		t.Fatalf("write error: %+v", res)
	}

	getReq := mcp.CallToolRequest{}
	getReq.Params.Arguments = map[string]any{
		"slug": "guides/new-page",
	}
	res, err = doctreeGetHandler(deps)(ctx, getReq)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if res.IsError {
		t.Fatalf("get error: %+v", res)
	}
	if !containsText(res, "New Page") || !containsText(res, "Content here") {
		t.Fatalf("unexpected get body: %+v", res)
	}
}

func TestDoctreeSections(t *testing.T) {
	deps, ctx := testDoctreeDeps(t)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":  "guides/setup",
		"level": float64(2),
		"index": float64(0),
	}
	res, err := doctreeSectionsHandler(deps)(ctx, req)
	if err != nil {
		t.Fatalf("sections: %v", err)
	}
	if res.IsError {
		t.Fatalf("sections error: %+v", res)
	}
	if !containsText(res, "Intro") {
		t.Fatalf("expected Intro section, got %+v", res)
	}
}

func containsText(res *mcp.CallToolResult, sub string) bool {
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok && contains(tc.Text, sub) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && stringIndex(s, sub) >= 0)
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
