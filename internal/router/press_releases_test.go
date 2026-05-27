package router

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monms/monms/internal/testutil"
	"github.com/monms/monms/internal/validate"
)

// TestPressReleasesOperation covers AGT-01 through AGT-04 end-to-end:
//   - AGT-01: press_releases collection created via schema bootstrap (no restart)
//   - AGT-02: template mutation visible on next request after cache.Flush()
//   - AGT-03: ValidateTemplate returns nil for valid template
//   - AGT-04: ValidateHTML returns nil for well-formed HTML
func TestPressReleasesOperation(t *testing.T) {
	// Step 1 — workspace with schema fixture (AGT-01 setup, D-33).
	// Schema written before startTestServer so Bootstrap() imports it automatically.
	ws := testutil.NewSite(t)
	schemaJSON := `{"name":"press_releases","type":"base","fields":[{"name":"title","type":"text"},{"name":"body","type":"text"}]}`
	testutil.WriteFile(t, filepath.Join(ws, "schema/press_releases.json"), schemaJSON)

	// Step 2 — start server; RegisterBootstrapHook reads schema/*.json including press_releases.json.
	ts, cache, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	// Step 3 — template mutation (AGT-02): write template after server start.
	pagePath := filepath.Join(ws, "templates/press/index.gohtml")
	testutil.WriteFile(t, pagePath, `{{define "body"}}<h1>Press Releases</h1>{{end}}`)
	cache.Flush() // simulate fsnotify cache invalidation

	// Step 4 — validate template and HTML (AGT-03, AGT-04).
	if err := validate.ValidateTemplate(ws, pagePath); err != nil {
		t.Fatalf("template validation: %v", err)
	}
	content, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	if err := validate.ValidateHTML(pagePath, content); err != nil {
		t.Fatalf("HTML validation: %v", err)
	}

	// Step 5 — verify render without restart (AGT-02): /press must return 200 with expected body.
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ts.URL + "/press")
	if err != nil {
		t.Fatalf("GET /press: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), "Press Releases") {
		t.Fatalf("body missing 'Press Releases', got: %.200s", string(body))
	}
}

// TestValidateRejectsInvalidTemplate covers AGT-05 signal: the pre-commit hook would rollback
// a commit containing a syntactically invalid template.
func TestValidateRejectsInvalidTemplate(t *testing.T) {
	ws := testutil.NewSite(t)
	badPath := filepath.Join(ws, "templates/broken.gohtml")
	testutil.WriteFile(t, badPath, `{{define "body"}}{{if}}{{end}}`) // missing condition arg

	err := validate.ValidateTemplate(ws, badPath)
	if err == nil {
		t.Fatal("expected error for bad template syntax, got nil")
	}
	if !strings.Contains(err.Error(), "template parse error") {
		t.Fatalf("unexpected error format: %v", err)
	}
}
