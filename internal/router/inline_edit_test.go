package router

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monms/monms/internal/testutil"
	"github.com/pocketbase/pocketbase/core"
)

const heroContentSchema = `{
  "name": "hero_content",
  "type": "base",
  "listRule": "",
  "viewRule": "",
  "updateRule": "@request.auth.id != ''",
  "createRule": "@request.auth.id != ''",
  "deleteRule": "@request.auth.id != ''",
  "fields": [
    {
      "name": "id",
      "type": "text",
      "required": true,
      "primaryKey": true,
      "system": true,
      "min": 1,
      "max": 50,
      "pattern": "^[a-z][a-z0-9_]*$"
    },
    {"name": "title", "type": "text"},
    {"name": "body", "type": "text"}
  ]
}`

func setupInlineEditWorkspace(t *testing.T) string {
	t.Helper()

	ws := testutil.NewWorkspace(t)
	testutil.WriteFile(t, filepath.Join(ws, "schema/hero_content.json"), heroContentSchema)

	// Copy Phase 3 templates from repo workspace fixtures.
	repoRoot := filepath.Join("..", "..")
	baseSrc := filepath.Join(repoRoot, "workspace/templates/layouts/base.gohtml")
	indexSrc := filepath.Join(repoRoot, "workspace/templates/index.gohtml")
	cssSrc := filepath.Join(repoRoot, "workspace/assets/main.css")

	copyFile(t, baseSrc, filepath.Join(ws, "templates/layouts/base.gohtml"))
	copyFile(t, indexSrc, filepath.Join(ws, "templates/index.gohtml"))
	copyFile(t, cssSrc, filepath.Join(ws, "assets/main.css"))

	return ws
}

func copyFile(t *testing.T, src, dest string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	testutil.WriteFile(t, dest, string(data))
}

func TestInlineEdit_UnauthenticatedHidesEdit(t *testing.T) {
	ws := setupInlineEditWorkspace(t)
	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	bodyStr := string(body)

	for _, forbidden := range []string{"contenteditable", "Live Editor Active", "hx-patch", "document.cookie", "monms_auth", "id=\"monms-sign-out\""} {
		if strings.Contains(bodyStr, forbidden) {
			t.Fatalf("guest body must not contain %q", forbidden)
		}
	}
	if !strings.Contains(bodyStr, "Welcome to MonMS") {
		t.Fatalf("guest body missing seeded hero title, got: %.300s", bodyStr)
	}
}

func TestInlineEdit_AuthenticatedShowsBadge(t *testing.T) {
	ws := setupInlineEditWorkspace(t)
	ts, app, _, cleanup := startTestServerWithApp(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	user := testutil.NewSuperuser(t, app, "editor@test.local")
	client := testutil.AuthClient(t, app, user)
	client.Timeout = 10 * time.Second

	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "Live Editor Active") {
		t.Fatal("authenticated body missing Live Editor Active badge")
	}
	if !strings.Contains(bodyStr, "contenteditable") {
		t.Fatal("authenticated body missing contenteditable")
	}
	if !strings.Contains(bodyStr, "hero_content/records/homepage") {
		t.Fatal("authenticated body missing hx-patch target")
	}
	if !strings.Contains(bodyStr, `hx-swap="none"`) {
		t.Fatal("authenticated body missing hx-swap=\"none\" (prevents JSON response replacing edited text)")
	}
	if !strings.Contains(bodyStr, "editor-save-error") {
		t.Fatal("dev mode authenticated body missing editor-save-error banner")
	}
	if !strings.Contains(bodyStr, "Authorization") || !strings.Contains(bodyStr, "Bearer") {
		t.Fatal("authenticated body missing Bearer auth script")
	}
	if strings.Contains(bodyStr, "document.cookie") {
		t.Fatal("body must not reference document.cookie (SEC-04)")
	}
	if !strings.Contains(bodyStr, "monms-sign-out") {
		t.Fatal("authenticated body missing Sign out link")
	}
}

func TestAuthCookieBridge_LoginThenSSR(t *testing.T) {
	ws := setupInlineEditWorkspace(t)
	ts, app, _, cleanup := startTestServerWithApp(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	testutil.NewSuperuser(t, app, "bridge@test.local")

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
	}

	authResp, err := client.Post(
		ts.URL+"/api/collections/"+core.CollectionNameSuperusers+"/auth-with-password",
		"application/json",
		strings.NewReader(`{"identity":"bridge@test.local","password":"password123456"}`),
	)
	if err != nil {
		t.Fatalf("auth POST: %v", err)
	}
	defer authResp.Body.Close()
	if authResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(authResp.Body)
		t.Fatalf("auth status %d: %s", authResp.StatusCode, b)
	}

	foundCookie := false
	for _, raw := range authResp.Header.Values("Set-Cookie") {
		if strings.HasPrefix(raw, authCookieName+"=") {
			foundCookie = true
			break
		}
	}
	if !foundCookie {
		t.Fatalf("auth response missing %s Set-Cookie: %v", authCookieName, authResp.Header.Values("Set-Cookie"))
	}

	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "Live Editor Active") {
		t.Fatalf("expected logged-in homepage after cookie bridge, got: %.400s", bodyStr)
	}
}

func TestHeroContent_AuthenticatedPatchPersists(t *testing.T) {
	ws := setupInlineEditWorkspace(t)
	ts, app, _, cleanup := startTestServerWithApp(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	user := testutil.NewSuperuser(t, app, "patch@test.local")
	client := testutil.AuthClient(t, app, user)
	client.Timeout = 10 * time.Second

	req, err := http.NewRequest(http.MethodPatch, ts.URL+"/api/collections/hero_content/records/homepage", strings.NewReader(`{"title":"Saved via PATCH"}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("PATCH: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("PATCH status %d: %s", resp.StatusCode, b)
	}

	rec, err := app.FindRecordById(heroCollection, heroRecordID)
	if err != nil {
		t.Fatalf("find record: %v", err)
	}
	if rec.GetString("title") != "Saved via PATCH" {
		t.Fatalf("title %q, want Saved via PATCH", rec.GetString("title"))
	}
}

func TestHeroContent_GuestPutForbidden(t *testing.T) {
	ws := setupInlineEditWorkspace(t)
	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPatch, ts.URL+"/api/collections/hero_content/records/homepage", strings.NewReader(`{"title":"Hacked"}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("PATCH: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 400 {
		t.Fatalf("guest PATCH status %d, want >= 400", resp.StatusCode)
	}
}
