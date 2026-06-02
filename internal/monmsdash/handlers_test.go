package monmsdash_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monms/monms/internal/authbootstrap"
	"github.com/monms/monms/internal/content"
	"github.com/monms/monms/internal/documents"
	"github.com/monms/monms/internal/monmsdash"
	"github.com/monms/monms/internal/monmsroutes"
	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/site"
	"github.com/monms/monms/internal/testutil"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/ui"
)

const testPublishToken = "test-publish-token"

func testLoadAuthFromCookie(e *core.RequestEvent) error {
	if e.Auth != nil {
		return nil
	}
	c, err := e.Request.Cookie("monms_auth")
	if err != nil || c.Value == "" {
		return nil
	}
	record, err := e.App.FindAuthRecordByToken(c.Value, core.TokenTypeAuth)
	if err != nil || record == nil {
		return nil
	}
	e.Auth = record
	return nil
}

func startDashboardServer(t *testing.T, siteAbs, publishToken string, loadAuth func(*core.RequestEvent) error) (*httptest.Server, core.App, func()) {
	t.Helper()

	if err := site.ValidateSite(siteAbs); err != nil {
		t.Fatalf("validate site: %v", err)
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(siteAbs, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})

	authbootstrap.RegisterBootstrapHook(app)
	documents.RegisterBootstrapHook(app, siteAbs)
	schema.RegisterBootstrapHook(app, siteAbs)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		deps := monmsdash.Deps{
			SiteAbs:      siteAbs,
			PublishToken: publishToken,
			LoadAuth:     loadAuth,
		}
		monmsdash.RegisterRoutes(se, deps)
		content.RegisterRoutes(se, monmsdash.PublishDeps(deps))
		return se.Next()
	})

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	if ui.DistDirFS != nil {
		router.GET("/_/{path...}", apis.Static(ui.DistDirFS, false))
	}

	serveEvent := &core.ServeEvent{App: app, Router: router}
	if err := app.OnServe().Trigger(serveEvent, func(e *core.ServeEvent) error {
		return e.Next()
	}); err != nil {
		t.Fatalf("on serve: %v", err)
	}

	mux, err := router.BuildMux()
	if err != nil {
		t.Fatalf("build mux: %v", err)
	}

	ts := httptest.NewServer(mux)
	return ts, app, func() {
		ts.Close()
		_ = app.ResetBootstrapState()
	}
}

func TestDashboardHomeRedirectsWhenUnauthenticated(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, _, cleanup := startDashboardServer(t, ws, testPublishToken, nil)
	defer cleanup()

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(ts.URL + monmsroutes.DashboardHomePath)
	if err != nil {
		t.Fatalf("GET dashboard: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != monmsroutes.AdminPath {
		t.Fatalf("Location %q, want %q", loc, monmsroutes.AdminPath)
	}
}

func TestDashboardStaticAssetsOffline(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, _, cleanup := startDashboardServer(t, ws, testPublishToken, nil)
	defer cleanup()

	for _, path := range []string{
		"/_monms/static/monms-dash.css",
		"/_monms/static/components.css",
		"/_monms/static/alpine.min.js",
		"/_monms/static/htmx.min.js",
		"/_monms/static/js/messages.js",
		"/_monms/static/js/system.js",
		"/_monms/static/fonts/inter-latin.woff2",
	} {
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s status %d, want 200", path, resp.StatusCode)
		}
	}
}

func TestDashboardHomeForSuperuser(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	user := testutil.NewSuperuser(t, app, "consultant@test.local")
	token, err := user.NewAuthToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+monmsroutes.DashboardHomePath, nil)
	req.AddCookie(&http.Cookie{Name: "monms_auth", Value: token})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET home: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200; body: %.300s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "Consultant tools") {
		t.Fatalf("missing consultant section")
	}
	if !strings.Contains(string(body), "monms-msg-strip") {
		t.Fatalf("missing message strip")
	}
	if !strings.Contains(string(body), "consultant@test.local") {
		t.Fatalf("missing user email in sidebar")
	}
}

func TestPublishUIReturns200(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, nil)
	defer cleanup()

	publisher := testutil.NewSuperuser(t, app, "publisher@test.local")
	client := testutil.AuthClient(t, app, publisher)

	resp, err := client.Get(ts.URL + monmsroutes.PublishPath)
	if err != nil {
		t.Fatalf("GET publish: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200; body: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if !strings.Contains(string(body), "Publish to live") {
		t.Fatalf("body missing title, got: %.400s", body)
	}
	if !strings.Contains(string(body), "MonMS Console") {
		t.Fatalf("publish page should use dashboard shell")
	}
}

func TestPublishUIReturns200WithCookie(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	publisher := testutil.NewSuperuser(t, app, "publisher@test.local")
	token, err := publisher.NewAuthToken()
	if err != nil {
		t.Fatalf("new auth token: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, ts.URL+monmsroutes.PublishPath, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "monms_auth", Value: token})

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET publish with cookie: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200; body: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

func TestPublisherGate(t *testing.T) {
	ws := testutil.NewEditorialSite(t)

	mockProd := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != monmsroutes.ContentImportPath {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"upserted":1,"collections":1}`))
	}))
	defer mockProd.Close()

	cfgPath := filepath.Join(ws, ".monms/config.json")
	testutil.WriteFile(t, cfgPath, fmt.Sprintf(
		`{"productionUrl":%q,"publisherEmails":["publisher@test.local"]}`,
		mockProd.URL,
	))

	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, nil)
	defer cleanup()

	editor := testutil.NewSuperuser(t, app, "editor@test.local")
	editorClient := testutil.AuthClient(t, app, editor)

	postReq, err := http.NewRequest(http.MethodPost, ts.URL+monmsroutes.PublishPath, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	postResp, err := editorClient.Do(postReq)
	if err != nil {
		t.Fatalf("POST publish as editor: %v", err)
	}
	postResp.Body.Close()
	if postResp.StatusCode != http.StatusForbidden {
		t.Fatalf("editor POST status %d, want 403", postResp.StatusCode)
	}

	publisher := testutil.NewSuperuser(t, app, "publisher@test.local")
	pubClient := testutil.AuthClient(t, app, publisher)

	pubReq, err := http.NewRequest(http.MethodPost, ts.URL+monmsroutes.PublishPath, nil)
	if err != nil {
		t.Fatalf("new publisher request: %v", err)
	}
	pubResp, err := pubClient.Do(pubReq)
	if err != nil {
		t.Fatalf("POST publish as publisher: %v", err)
	}
	defer pubResp.Body.Close()

	pubBody, err := io.ReadAll(pubResp.Body)
	if err != nil {
		t.Fatalf("read publisher body: %v", err)
	}
	if pubResp.StatusCode != http.StatusOK {
		t.Fatalf("publisher POST status %d, want 200; body: %s", pubResp.StatusCode, strings.TrimSpace(string(pubBody)))
	}
	if !strings.Contains(string(pubBody), `data-flash="Content published successfully."`) {
		t.Fatalf("publish success should surface in navbar message strip")
	}

	state, err := content.ReadPublishState(ws)
	if err != nil {
		t.Fatalf("read publish state: %v", err)
	}
	if state.Checksum == "" {
		t.Fatal("publish-state checksum empty after successful publish")
	}
}

func TestAPIKeysPageForSuperuser(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	user := testutil.NewSuperuser(t, app, "keys-ui@test.local")
	token, err := user.NewAuthToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+monmsroutes.APIKeysPath, nil)
	req.AddCookie(&http.Cookie{Name: "monms_auth", Value: token})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET api-keys: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200; body: %.200s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "API Keys") {
		t.Fatalf("missing page title")
	}
}

func TestMCPSettingsForbiddenForNonSuperuser(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatalf("users collection: %v", err)
	}
	regular := core.NewRecord(usersCol)
	regular.Set("email", "regular@test.local")
	regular.SetPassword("password123456")
	regular.Set("verified", true)
	if err := app.Save(regular); err != nil {
		t.Fatalf("save user: %v", err)
	}
	token, err := regular.NewAuthToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+monmsroutes.MCPSettingsPath, nil)
	req.AddCookie(&http.Cookie{Name: "monms_auth", Value: token})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET mcp: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status %d, want 403", resp.StatusCode)
	}
}

func TestDoctreesPageListsMarkdownCollection(t *testing.T) {
	ws := testutil.NewSite(t)
	schemaJSON := `{
  "name": "articles",
  "type": "base",
  "editorial": true,
  "monms": { "source": "markdown", "root": "documents/articles" },
  "fields": [
    { "name": "id", "type": "text", "primaryKey": true, "required": true, "pattern": "^[a-z][a-z0-9_-]*$", "min": 1, "max": 120 },
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "path", "type": "text" },
    { "name": "body", "type": "text" }
  ]
}`
	testutil.WriteFile(t, filepath.Join(ws, "schema/articles.json"), schemaJSON)
	testutil.WriteFile(t, filepath.Join(ws, "documents/articles/hello.md"), `---
title: Hello Doc
---
Body
`)

	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	user := testutil.NewSuperuser(t, app, "reader@test.local")
	client := testutil.AuthClient(t, app, user)

	resp, err := client.Get(ts.URL + monmsroutes.DoctreesPath)
	if err != nil {
		t.Fatalf("GET doctrees: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200; body: %.300s", resp.StatusCode, body)
	}
	text := string(body)
	if !strings.Contains(text, "Doctrees") {
		t.Fatalf("missing page title: %.400s", text)
	}
	if !strings.Contains(text, "articles") {
		t.Fatalf("missing collection name: %.400s", text)
	}
	if !strings.Contains(text, "Hello Doc") {
		t.Fatalf("missing document title: %.400s", text)
	}
}

func TestDocumentsPathRedirectsToDoctrees(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	user := testutil.NewSuperuser(t, app, "reader@test.local")
	client := testutil.AuthClient(t, app, user)

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Get(ts.URL + monmsroutes.DocumentsPath)
	if err != nil {
		t.Fatalf("GET legacy documents path: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("status %d, want 301", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if !strings.HasSuffix(loc, monmsroutes.DoctreesPath) {
		t.Fatalf("Location = %q, want suffix %q", loc, monmsroutes.DoctreesPath)
	}
}

func TestDoctreesMigratePublisherGate(t *testing.T) {
	ws := testutil.NewSite(t)
	cfgPath := filepath.Join(ws, ".monms/config.json")
	testutil.WriteFile(t, cfgPath, `{"publisherEmails":["publisher@test.local"]}`)

	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	editor := testutil.NewSuperuser(t, app, "editor@test.local")
	client := testutil.AuthClient(t, app, editor)

	form := url.Values{"source_root": {t.TempDir()}, "doctree_stub": {"guide"}}
	req, err := http.NewRequest(http.MethodPost, ts.URL+monmsroutes.DoctreesPath+"/migrate/copy", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST migrate preview: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("editor POST status %d, want 403", resp.StatusCode)
	}
}

func TestDoctreesPageShowsAlignmentIssues(t *testing.T) {
	site := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(site, "guide", "tutorials", "page.md"), "# Page\n")
	testutil.WriteFile(t, filepath.Join(site, "schema", "dt_tutorial.json"), `{
  "name": "dt_tutorial",
  "type": "base",
  "editorial": true,
  "monms": { "source": "markdown", "root": "guide/tutorials", "doctree": "guide" },
  "fields": [{ "name": "id", "type": "text", "primaryKey": true, "required": true }]
}`)

	ts, app, cleanup := startDashboardServer(t, site, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	user := testutil.NewSuperuser(t, app, "consultant@test.local")
	client := testutil.AuthClient(t, app, user)

	resp, err := client.Get(ts.URL + monmsroutes.DoctreesPath)
	if err != nil {
		t.Fatalf("GET doctrees: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200; body: %.400s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "Doctree alignment needs resolution") {
		t.Fatalf("expected alignment alert on doctrees page")
	}
	if !strings.Contains(string(body), "collection_rename") {
		t.Fatalf("expected collection_rename in alignment table")
	}
}

func TestDoctreesMigrateCopyConfirm(t *testing.T) {
	legacy := t.TempDir()
	testutil.WriteFile(t, filepath.Join(legacy, "welcome.md"), `---
title: Welcome
---
Welcome body
`)

	ws := testutil.NewSite(t)
	cfgPath := filepath.Join(ws, ".monms/config.json")
	testutil.WriteFile(t, cfgPath, `{"publisherEmails":["publisher@test.local"]}`)

	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	publisher := testutil.NewSuperuser(t, app, "publisher@test.local")
	client := testutil.AuthClient(t, app, publisher)

	copyForm := url.Values{"source_root": {legacy}, "doctree_stub": {"guide"}}
	copyReq, err := http.NewRequest(http.MethodPost, ts.URL+monmsroutes.DoctreesPath+"/migrate/copy", strings.NewReader(copyForm.Encode()))
	if err != nil {
		t.Fatalf("copy request: %v", err)
	}
	copyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	copyResp, err := client.Do(copyReq)
	if err != nil {
		t.Fatalf("POST copy: %v", err)
	}
	copyBody, err := io.ReadAll(copyResp.Body)
	copyResp.Body.Close()
	if err != nil {
		t.Fatalf("read copy body: %v", err)
	}
	if copyResp.StatusCode != http.StatusOK {
		t.Fatalf("copy status %d, want 200; body: %.400s", copyResp.StatusCode, copyBody)
	}
	if !strings.Contains(string(copyBody), "Confirm bindings") {
		t.Fatalf("copy should show confirm step: %.300s", copyBody)
	}

	confirmForm := url.Values{"doctree_stub": {"guide"}}
	confirmReq, err := http.NewRequest(http.MethodPost, ts.URL+monmsroutes.DoctreesPath+"/migrate/confirm", strings.NewReader(confirmForm.Encode()))
	if err != nil {
		t.Fatalf("confirm request: %v", err)
	}
	confirmReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	confirmResp, err := client.Do(confirmReq)
	if err != nil {
		t.Fatalf("POST confirm: %v", err)
	}
	confirmBody, err := io.ReadAll(confirmResp.Body)
	confirmResp.Body.Close()
	if err != nil {
		t.Fatalf("read confirm body: %v", err)
	}
	if confirmResp.StatusCode != http.StatusOK {
		t.Fatalf("confirm status %d, want 200; body: %.400s", confirmResp.StatusCode, confirmBody)
	}
	if !strings.Contains(string(confirmBody), `data-flash="Bindings confirmed`) {
		t.Fatalf("confirm success should surface in navbar strip: %.400s", confirmBody)
	}

	migrated := filepath.Join(ws, "guide", "welcome.md")
	if _, err := os.Stat(migrated); err != nil {
		t.Fatalf("migrated file missing: %v", err)
	}
}

func TestDoctreesMigratePruneRemovesStub(t *testing.T) {
	legacy := t.TempDir()
	testutil.WriteFile(t, filepath.Join(legacy, "a.md"), "# A\n")

	ws := testutil.NewSite(t)
	cfgPath := filepath.Join(ws, ".monms/config.json")
	testutil.WriteFile(t, cfgPath, `{"publisherEmails":["publisher@test.local"]}`)

	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	publisher := testutil.NewSuperuser(t, app, "publisher@test.local")
	client := testutil.AuthClient(t, app, publisher)

	copyForm := url.Values{"source_root": {legacy}, "doctree_stub": {"pruneguide"}}
	copyReq, _ := http.NewRequest(http.MethodPost, ts.URL+monmsroutes.DoctreesPath+"/migrate/copy", strings.NewReader(copyForm.Encode()))
	copyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	copyResp, err := client.Do(copyReq)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	copyResp.Body.Close()

	pruneForm := url.Values{"doctree_stub": {"pruneguide"}}
	pruneReq, _ := http.NewRequest(http.MethodPost, ts.URL+monmsroutes.DoctreesPath+"/migrate/prune", strings.NewReader(pruneForm.Encode()))
	pruneReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	pruneResp, err := client.Do(pruneReq)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	pruneResp.Body.Close()

	if _, err := os.Stat(filepath.Join(ws, "pruneguide")); !os.IsNotExist(err) {
		t.Fatal("pruneguide should be removed")
	}
}

func TestSystemPageRedirectsWhenUnauthenticated(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, _, cleanup := startDashboardServer(t, ws, testPublishToken, nil)
	defer cleanup()

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(ts.URL + monmsroutes.SystemPath)
	if err != nil {
		t.Fatalf("GET system: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status %d, want 303", resp.StatusCode)
	}
}

func TestSystemPageForbiddenForNonSuperuser(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatalf("users collection: %v", err)
	}
	regular := core.NewRecord(usersCol)
	regular.Set("email", "regular-sys@test.local")
	regular.SetPassword("password123456")
	regular.Set("verified", true)
	if err := app.Save(regular); err != nil {
		t.Fatalf("save user: %v", err)
	}
	token, err := regular.NewAuthToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+monmsroutes.SystemPath, nil)
	req.AddCookie(&http.Cookie{Name: "monms_auth", Value: token})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET system: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status %d, want 403", resp.StatusCode)
	}
}

func TestSystemPageForSuperuser(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	user := testutil.NewSuperuser(t, app, "sys-ui@test.local")
	token, err := user.NewAuthToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+monmsroutes.SystemPath, nil)
	req.AddCookie(&http.Cookie{Name: "monms_auth", Value: token})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET system: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200; body: %.300s", resp.StatusCode, body)
	}
	text := string(body)
	for _, want := range []string{"System", "Build mode", "Restart MonMS", "Shutdown MonMS", filepath.Base(ws)} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in body", want)
		}
	}
}

func TestSystemRestartPostSurfacesFlash(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, app, cleanup := startDashboardServer(t, ws, testPublishToken, testLoadAuthFromCookie)
	defer cleanup()

	restore := monmsdash.SetLifecycleHooksForTest(
		func(monmsdash.Deps) error { return nil },
		func() error { return nil },
	)
	defer restore()

	user := testutil.NewSuperuser(t, app, "sys-restart@test.local")
	token, err := user.NewAuthToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, ts.URL+monmsroutes.SystemRestartPath, nil)
	req.AddCookie(&http.Cookie{Name: "monms_auth", Value: token})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST restart: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200; body: %.200s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `data-flash="Restarting MonMS…"`) {
		t.Fatal("expected restart flash in navbar strip")
	}
}
