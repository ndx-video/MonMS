package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monms/monms/internal/monmsroutes"
	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/testutil"
	"github.com/monms/monms/internal/site"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/ui"
)

const testPublishToken = "test-publish-token"

func startTestContentServer(t *testing.T, siteAbs, publishToken string, loadAuth func(*core.RequestEvent) error) (*httptest.Server, core.App, func()) {
	t.Helper()

	if err := site.ValidateSite(siteAbs); err != nil {
		t.Fatalf("validate site: %v", err)
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(siteAbs, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})

	schema.RegisterBootstrapHook(app, siteAbs)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		RegisterRoutes(se, Deps{
			SiteAbs:        siteAbs,
			PublishToken: publishToken,
			LoadAuth:     loadAuth,
		})
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

func postImport(t *testing.T, client *http.Client, url, token string, body any) (*http.Response, []byte) {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST import: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, respBody
}

func TestImportAPIUnauthorized(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	ts, app, cleanup := startTestContentServer(t, ws, testPublishToken, nil)
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	url := ts.URL + monmsroutes.ContentImportPath

	t.Run("missing authorization", func(t *testing.T) {
		resp, _ := postImport(t, client, url, "", importRequest{
			Collections: []importCollection{{
				Name: "hero_content",
				Records: []map[string]any{
					{"id": "homepage", "title": "Blocked"},
				},
			}},
		})
		if resp.StatusCode < 400 {
			t.Fatalf("status %d, want >= 400", resp.StatusCode)
		}
	})

	t.Run("wrong token", func(t *testing.T) {
		resp, _ := postImport(t, client, url, "wrong-token", importRequest{
			Collections: []importCollection{{
				Name: "hero_content",
				Records: []map[string]any{
					{"id": "homepage", "title": "Blocked"},
				},
			}},
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status %d, want 401", resp.StatusCode)
		}
	})

	t.Run("valid token upserts record", func(t *testing.T) {
		wantTitle := "Import API Title"
		resp, body := postImport(t, client, url, testPublishToken, importRequest{
			Collections: []importCollection{{
				Name: "hero_content",
				Records: []map[string]any{
					{"id": "homepage", "title": wantTitle, "body": "Updated via API"},
				},
			}},
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status %d, want 200; body: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var report ImportReport
		if err := json.Unmarshal(body, &report); err != nil {
			t.Fatalf("unmarshal report: %v", err)
		}
		if report.Upserted != 1 {
			t.Fatalf("upserted %d, want 1", report.Upserted)
		}

		rec, err := app.FindRecordById("hero_content", "homepage")
		if err != nil {
			t.Fatalf("find record: %v", err)
		}
		if got := rec.GetString("title"); got != wantTitle {
			t.Fatalf("title %q, want %q", got, wantTitle)
		}
	})

	t.Run("non editorial collection rejected", func(t *testing.T) {
		resp, body := postImport(t, client, url, testPublishToken, importRequest{
			Collections: []importCollection{{
				Name: "press_releases",
				Records: []map[string]any{
					{"id": "pr1", "title": "Nope"},
				},
			}},
		})
		if resp.StatusCode < 400 {
			t.Fatalf("status %d, want >= 400; body: %s", resp.StatusCode, body)
		}
	})

	t.Run("system collection rejected", func(t *testing.T) {
		resp, body := postImport(t, client, url, testPublishToken, importRequest{
			Collections: []importCollection{{
				Name: "users",
				Records: []map[string]any{
					{"id": "u1"},
				},
			}},
		})
		if resp.StatusCode < 400 {
			t.Fatalf("status %d, want >= 400; body: %s", resp.StatusCode, body)
		}
	})
}

func TestPublishUIReturns200(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	ts, app, cleanup := startTestContentServer(t, ws, testPublishToken, nil)
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
}

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

func TestPublishUIReturns200WithCookie(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	ts, app, cleanup := startTestContentServer(t, ws, testPublishToken, testLoadAuthFromCookie)
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
	if !strings.Contains(string(body), "Publish to live") {
		t.Fatalf("body missing title, got: %.400s", body)
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

	ts, app, cleanup := startTestContentServer(t, ws, testPublishToken, nil)
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

	state, err := ReadPublishState(ws)
	if err != nil {
		t.Fatalf("read publish state: %v", err)
	}
	if state.Checksum == "" {
		t.Fatal("publish-state checksum empty after successful publish")
	}
	if state.PublishedAt == "" {
		t.Fatal("publish-state publishedAt empty after successful publish")
	}
}

func TestImportAPIFailClosedEmptyToken(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	ts, _, cleanup := startTestContentServer(t, ws, "", nil)
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, _ := postImport(t, client, ts.URL+monmsroutes.ContentImportPath, testPublishToken, importRequest{
		Collections: []importCollection{{
			Name: "hero_content",
			Records: []map[string]any{
				{"id": "homepage", "title": "Should not apply"},
			},
		}},
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401 when publish token unset", resp.StatusCode)
	}
}
