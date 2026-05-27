package router

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/monms/monms/internal/monmsroutes"
	"github.com/monms/monms/internal/testutil"
	"github.com/pocketbase/pocketbase/core"
)

func TestLogout_ClearsAuthCookie(t *testing.T) {
	ws := setupInlineEditSite(t)
	ts, app, _, cleanup := startTestServerWithApp(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	user := testutil.NewSuperuser(t, app, "cookie-clear@test.local")
	token, err := user.NewAuthToken()
	if err != nil {
		t.Fatalf("new auth token: %v", err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	loginURL := mustParseURL(t, ts.URL)
	jar.SetCookies(loginURL, []*http.Cookie{{Name: authCookieName, Value: token, Path: "/"}})

	resp, err := client.Get(ts.URL + monmsroutes.AuthLogoutPath + "?redirect=/")
	if err != nil {
		t.Fatalf("GET logout: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("logout status %d, want 303", resp.StatusCode)
	}

	for _, c := range jar.Cookies(loginURL) {
		if c.Name == authCookieName && c.Value != "" {
			t.Fatalf("expected cleared auth cookie, got value=%q", c.Value)
		}
	}
}

func TestLogout_ThenSSRIsGuest(t *testing.T) {
	ws := setupInlineEditSite(t)
	ts, app, _, cleanup := startTestServerWithApp(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	testutil.NewSuperuser(t, app, "guest-after-logout@test.local")

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	authResp, err := client.Post(
		ts.URL+"/api/collections/"+core.CollectionNameSuperusers+"/auth-with-password",
		"application/json",
		strings.NewReader(`{"identity":"guest-after-logout@test.local","password":"password123456"}`),
	)
	if err != nil {
		t.Fatalf("auth POST: %v", err)
	}
	authResp.Body.Close()
	if authResp.StatusCode != http.StatusOK {
		t.Fatalf("auth status %d, want 200", authResp.StatusCode)
	}

	loginResp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / logged in: %v", err)
	}
	loginBody, _ := io.ReadAll(loginResp.Body)
	loginResp.Body.Close()
	if !strings.Contains(string(loginBody), "Live Editor Active") {
		t.Fatal("expected logged-in homepage before logout")
	}

	logoutResp, err := client.Get(ts.URL + monmsroutes.AuthLogoutPath + "?redirect=/")
	if err != nil {
		t.Fatalf("GET logout: %v", err)
	}
	logoutResp.Body.Close()
	if logoutResp.StatusCode != http.StatusSeeOther {
		t.Fatalf("logout status %d, want 303", logoutResp.StatusCode)
	}

	client.CheckRedirect = nil
	guestClient := &http.Client{Timeout: 10 * time.Second, Jar: jar}
	guestResp, err := guestClient.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / after logout: %v", err)
	}
	defer guestResp.Body.Close()
	guestBody, err := io.ReadAll(guestResp.Body)
	if err != nil {
		t.Fatalf("read guest body: %v", err)
	}
	guestStr := string(guestBody)
	if strings.Contains(guestStr, "Live Editor Active") {
		t.Fatal("homepage still shows editor after logout")
	}
	if strings.Contains(guestStr, "contenteditable") {
		t.Fatal("homepage still has contenteditable after logout")
	}
}

func TestLoadAuthFromCookie_InvalidTokenClearsCookie(t *testing.T) {
	ws := setupInlineEditSite(t)
	_, app, _, cleanup := startTestServerWithApp(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: "invalid-token"})
	recorder := httptest.NewRecorder()

	event := &core.RequestEvent{App: app}
	event.Request = req
	event.Response = recorder

	if err := LoadAuthFromCookie(event); err != nil {
		t.Fatalf("LoadAuthFromCookie: %v", err)
	}
	if event.Auth != nil {
		t.Fatal("expected no auth for invalid token")
	}

	setCookies := recorder.Result().Cookies()
	var cleared bool
	for _, c := range setCookies {
		if c.Name == authCookieName && (c.MaxAge < 0 || c.Value == "") {
			cleared = true
		}
	}
	if !cleared {
		t.Fatalf("expected cleared auth cookie, got %#v", setCookies)
	}
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return u
}
