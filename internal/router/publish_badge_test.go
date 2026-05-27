package router

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/monmsroutes"
	"github.com/monms/monms/internal/testutil"
)

func setupPublishBadgeSite(t *testing.T) string {
	t.Helper()
	ws := setupInlineEditSite(t)
	if err := os.MkdirAll(filepath.Join(ws, ".monms"), 0o755); err != nil {
		t.Fatalf("mkdir .monms: %v", err)
	}
	testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"),
		`{"productionUrl":"http://127.0.0.1:1","publisherEmails":["publisher@test.local"]}`)
	return ws
}

func TestEditorBadge_PublishLinkForPublisher(t *testing.T) {
	ws := setupPublishBadgeSite(t)
	ts, app, _, cleanup := startTestServerWithApp(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	publisher := testutil.NewSuperuser(t, app, "publisher@test.local")
	client := testutil.AuthClient(t, app, publisher)

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

	if !strings.Contains(bodyStr, `href="`+monmsroutes.PublishPath+`"`) {
		t.Fatalf("publisher body missing publish link, got: %.400s", bodyStr)
	}
	if !strings.Contains(bodyStr, "Publish to live") {
		t.Fatalf("publisher body missing Publish to live label")
	}
}

func TestEditorBadge_NoPublishLinkForEditor(t *testing.T) {
	ws := setupPublishBadgeSite(t)
	ts, app, _, cleanup := startTestServerWithApp(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	editor := testutil.NewSuperuser(t, app, "editor@test.local")
	client := testutil.AuthClient(t, app, editor)

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

	if strings.Contains(bodyStr, `href="`+monmsroutes.PublishPath+`"`) {
		t.Fatalf("editor body must not contain publish link, got: %.400s", bodyStr)
	}
}
