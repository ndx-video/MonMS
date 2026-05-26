package router

import (
	"net/http"
	"strings"

	"github.com/monms/monms/internal/monmsroutes"
	"github.com/pocketbase/pocketbase/core"
)

const authCookieName = "monms_auth"

// RegisterAuthHooks wires HttpOnly cookie auth bridge for browser SSR sessions.
func RegisterAuthHooks(app core.App) {
	app.OnRecordAuthRequest().BindFunc(func(e *core.RecordAuthRequestEvent) error {
		if e.Token != "" {
			// Must set before e.Next(): inner handlers write JSON and flush headers.
			setAuthCookie(e.RequestEvent, e.Token)
		}
		return e.Next()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET(monmsroutes.AuthLogoutPath, logoutHandler)
		se.Router.POST(monmsroutes.AuthLogoutPath, logoutHandler)
		se.Router.POST(monmsroutes.AuthSyncPath, syncAuthHandler)
		return se.Next()
	})
}

func syncAuthHandler(e *core.RequestEvent) error {
	token := bearerToken(e.Request.Header.Get("Authorization"))
	if token == "" {
		return e.UnauthorizedError("Missing auth token.", nil)
	}
	record, err := e.App.FindAuthRecordByToken(token, core.TokenTypeAuth)
	if err != nil || record == nil {
		return e.UnauthorizedError("Invalid auth token.", err)
	}
	setAuthCookie(e, token)
	return e.NoContent(http.StatusNoContent)
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}

func setAuthCookie(e *core.RequestEvent, token string) {
	e.SetCookie(&http.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   e.IsTLS(),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAuthCookie(e *core.RequestEvent) {
	e.SetCookie(&http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   e.IsTLS(),
		SameSite: http.SameSiteLaxMode,
	})
}

func logoutHandler(e *core.RequestEvent) error {
	clearAuthCookie(e)
	if e.Request.Method == http.MethodPost {
		return e.NoContent(http.StatusNoContent)
	}
	return e.Redirect(http.StatusSeeOther, safeRedirectPath(e.Request.URL.Query().Get("redirect")))
}

func safeRedirectPath(path string) string {
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return "/"
	}
	return path
}

// LoadAuthFromCookie populates e.Auth from the monms_auth HttpOnly cookie when unset.
func LoadAuthFromCookie(e *core.RequestEvent) error {
	if e.Auth != nil {
		return nil
	}
	c, err := e.Request.Cookie(authCookieName)
	if err != nil || c.Value == "" {
		return nil
	}
	record, err := e.App.FindAuthRecordByToken(c.Value, core.TokenTypeAuth)
	if err != nil || record == nil {
		clearAuthCookie(e)
		return nil
	}
	e.Auth = record
	return nil
}
