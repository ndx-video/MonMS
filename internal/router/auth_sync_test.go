package router

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/monms/monms/internal/monmsroutes"
	"github.com/monms/monms/internal/testutil"
)

func TestSyncAuth_SetsCookieFromBearer(t *testing.T) {
	ws := setupInlineEditWorkspace(t)
	ts, app, _, cleanup := startTestServerWithApp(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	user := testutil.NewSuperuser(t, app, "sync@test.local")
	token, err := user.NewAuthToken()
	if err != nil {
		t.Fatalf("new auth token: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPost, ts.URL+monmsroutes.AuthSyncPath, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST sync-auth: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("sync-auth status %d: %s", resp.StatusCode, b)
	}

	found := false
	for _, c := range resp.Cookies() {
		if c.Name == authCookieName && c.Value == token {
			found = true
		}
	}
	if !found {
		t.Fatalf("sync-auth response missing %s cookie, got %#v", authCookieName, resp.Cookies())
	}
}

func TestSyncAuth_GuestHomepageScriptContainsSync(t *testing.T) {
	ws := setupInlineEditWorkspace(t)
	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if !strings.Contains(bodyStr, monmsroutes.AuthSyncPath) {
		t.Fatal("guest homepage missing auth sync script")
	}
}
