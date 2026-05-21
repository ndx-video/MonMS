package router

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
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

func startTestServer(t *testing.T, wsAbs string) (*httptest.Server, func()) {
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
	return ts, func() {
		ts.Close()
		_ = app.ResetBootstrapState()
	}
}

func TestServeStarts(t *testing.T) {
	ws := testutil.NewWorkspace(t)
	ts, cleanup := startTestServer(t, ws)
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET /api/health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		t.Fatalf("expected non-5xx from /api/health, got %d", resp.StatusCode)
	}
}

func TestAdminDashboard(t *testing.T) {
	ws := testutil.NewWorkspace(t)
	ts, cleanup := startTestServer(t, ws)
	defer cleanup()

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(ts.URL + "/_/")
	if err != nil {
		t.Fatalf("GET /_/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		t.Fatalf("expected admin dashboard 200 or 302, got %d", resp.StatusCode)
	}
}

func TestAssetsHandler(t *testing.T) {
	t.Skip("implemented in plan 01-04")
}

func Test404NoPanic(t *testing.T) {
	t.Skip("implemented in plan 01-04")
}

func TestHomepageSSR(t *testing.T) {
	t.Skip("implemented in plan 01-04")
}
