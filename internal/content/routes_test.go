package content

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/testutil"
	"github.com/monms/monms/internal/workspace"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/ui"
)

const testPublishToken = "test-publish-token"

func startTestContentServer(t *testing.T, wsAbs, publishToken string) (*httptest.Server, core.App, func()) {
	t.Helper()

	if err := workspace.ValidateWorkspace(wsAbs); err != nil {
		t.Fatalf("validate workspace: %v", err)
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(wsAbs, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})

	schema.RegisterBootstrapHook(app, wsAbs)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		RegisterRoutes(se, Deps{
			WsAbs:        wsAbs,
			PublishToken: publishToken,
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
	ws := testutil.NewEditorialWorkspace(t)
	ts, app, cleanup := startTestContentServer(t, ws, testPublishToken)
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	url := ts.URL + "/api/monms/content/import"

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

func TestImportAPIFailClosedEmptyToken(t *testing.T) {
	ws := testutil.NewEditorialWorkspace(t)
	ts, _, cleanup := startTestContentServer(t, ws, "")
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, _ := postImport(t, client, ts.URL+"/api/monms/content/import", testPublishToken, importRequest{
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
