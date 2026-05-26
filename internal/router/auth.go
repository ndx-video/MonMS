package router

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

const authCookieName = "monms_auth"

// RegisterAuthHooks wires HttpOnly cookie auth bridge for browser SSR sessions.
func RegisterAuthHooks(app core.App) {
	app.OnRecordAuthRequest().BindFunc(func(e *core.RecordAuthRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if e.Token == "" {
			return nil
		}
		e.SetCookie(&http.Cookie{
			Name:     authCookieName,
			Value:    e.Token,
			Path:     "/",
			HttpOnly: true,
			Secure:   e.IsTLS(),
			SameSite: http.SameSiteLaxMode,
		})
		return nil
	})
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
		return nil
	}
	e.Auth = record
	return nil
}
