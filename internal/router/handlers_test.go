package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/templates"
	"github.com/monms/monms/internal/testutil"
	"github.com/monms/monms/internal/site"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/ui"
)

type testServerOpts struct {
	isDev          bool
	productionMode bool
}

func startTestServer(t *testing.T, siteAbs string, opts testServerOpts) (*httptest.Server, *templates.TemplateCache, func()) {
	ts, _, cache, cleanup := startTestServerWithApp(t, siteAbs, opts)
	return ts, cache, cleanup
}

func startTestServerWithApp(t *testing.T, siteAbs string, opts testServerOpts) (*httptest.Server, core.App, *templates.TemplateCache, func()) {
	t.Helper()

	if err := site.ValidateSite(siteAbs); err != nil {
		t.Fatalf("validate site: %v", err)
	}

	cache := templates.NewCache()
	cache.SetProductionMode(opts.productionMode)

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(siteAbs, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})

	schema.RegisterBootstrapHook(app, siteAbs)
	RegisterAuthHooks(app)

	deps := Deps{SiteAbs: siteAbs, Cache: cache, IsDev: opts.isDev}

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		RegisterAdminUIExtension(se)
		RegisterRoutes(se, deps)
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
	return ts, app, cache, func() {
		ts.Close()
		_ = app.ResetBootstrapState()
	}
}

func TestServeStarts(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
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
	ws := testutil.NewSite(t)
	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
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

func TestAdminUIExtension_ViewSiteLink(t *testing.T) {
	if ui.DistDirFS == nil {
		t.Skip("admin UI not bundled")
	}

	ws := testutil.NewSite(t)
	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ts.URL + "/_/extensions.js")
	if err != nil {
		t.Fatalf("GET /_/extensions.js: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("extensions.js status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "View site") {
		t.Fatalf("extensions.js missing View site link: %q", bodyStr)
	}
	if !strings.Contains(bodyStr, `href: "/"`) {
		t.Fatalf("extensions.js missing home href: %q", bodyStr)
	}
}

func TestAssetsHandler(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("/assets/main.css", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/assets/main.css")
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status %d, want 200", resp.StatusCode)
		}
		if !strings.Contains(resp.Header.Get("Content-Type"), "text/css") {
			t.Fatalf("Content-Type %q, want text/css", resp.Header.Get("Content-Type"))
		}
	})

	t.Run("/assets/missing.css", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/assets/missing.css")
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status %d, want 404", resp.StatusCode)
		}
	})

}

func Test404NoPanic(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ts.URL + "/does-not-exist")
	if err != nil {
		t.Fatalf("GET /does-not-exist: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d, want 404", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "Page not found: /does-not-exist") {
		t.Fatalf("body missing path message, got: %s", bodyStr)
	}
}

func TestHomepageSSR(t *testing.T) {
	ws := testutil.NewSite(t)
	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "MonMS is running") {
		t.Fatalf("body missing hero title, got: %s", bodyStr)
	}
}

func TestFragmentPartial(t *testing.T) {
	ws := testutil.NewSite(t)
	fragPath := filepath.Join(ws, "templates/fragments", "nav.gohtml")
	testutil.WriteFile(t, fragPath, `<nav class="fragment-nav">Nav partial</nav>`)

	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ts.URL + "/fragments/nav")
	if err != nil {
		t.Fatalf("GET /fragments/nav: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	bodyStr := string(body)
	if strings.Contains(bodyStr, "<!DOCTYPE") {
		t.Fatal("fragment response must not include DOCTYPE from base layout")
	}
	if !strings.Contains(bodyStr, "Nav partial") {
		t.Fatalf("body missing fragment content, got: %s", bodyStr)
	}
}

func TestProduction500Generic(t *testing.T) {
	ws := testutil.NewSite(t)
	badPage := filepath.Join(ws, "templates", "broken.gohtml")
	testutil.WriteFile(t, badPage, `{{define "body"}}{{if}}{{end}}`)

	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: false, productionMode: true})
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ts.URL + "/broken")
	if err != nil {
		t.Fatalf("GET /broken: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	bodyStr := string(body)
	if strings.Contains(bodyStr, "can't evaluate") || strings.Contains(bodyStr, "BadField") {
		t.Fatalf("production 500 must not leak parse error, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "Internal server error") && !strings.Contains(bodyStr, "Something went wrong") {
		t.Fatalf("production 500 missing generic message, got: %s", bodyStr)
	}
}
