package monmsdash

import (
	"strings"

	"github.com/monms/monms/internal/apikeys"
	"github.com/monms/monms/internal/content"
	"github.com/monms/monms/internal/monmsroutes"
	"github.com/pocketbase/pocketbase/core"
)

type apiKeysPageData struct {
	PageData
	Keys           []apiKeyRow
	CanCreate      bool
	NewKeySecret   string
	PolicyDisabled bool
}

type apiKeyRow struct {
	ID         string
	Name       string
	Prefix     string
	Created    string
	LastUsedAt string
}

func registerAPIKeysRoutes(se *core.ServeEvent, deps Deps, tmpl *templates) {
	authBind := bindLoadAuth(deps.LoadAuth)
	authRedirect := requireAuthenticatedRedirect()
	keysGate := requireCanManageAPIKeys(deps.SiteAbs)

	se.Router.GET(monmsroutes.APIKeysPath, apiKeysGetHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(keysGate)

	se.Router.POST(monmsroutes.APIKeysPath, apiKeysPostHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(keysGate)

	se.Router.POST(monmsroutes.APIKeysRevokePath, apiKeysRevokeHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(keysGate)
}

func apiKeysGetHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data, err := buildAPIKeysPageData(e, deps, "", false)
		if err != nil {
			return e.InternalServerError("api keys", err)
		}
		return tmpl.renderPage(e.Response, "apikeys", data)
	}
}

func apiKeysPostHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if err := e.Request.ParseForm(); err != nil {
			return e.BadRequestError("invalid form", err)
		}
		cfg, err := content.LoadMonmsConfig(deps.SiteAbs)
		if err != nil {
			return e.InternalServerError("config", err)
		}
		if !apikeys.CanManageKeys(e.Auth, cfg) {
			return e.ForbiddenError("api key creation not allowed", nil)
		}

		name := strings.TrimSpace(e.Request.FormValue("name"))
		secret, _, err := apikeys.Create(e.App, deps.SiteAbs, e.Auth, name)
		if err != nil {
			data, derr := buildAPIKeysPageData(e, deps, err.Error(), true)
			if derr != nil {
				return e.InternalServerError("api keys", derr)
			}
			return tmpl.renderPage(e.Response, "apikeys", data)
		}

		data, err := buildAPIKeysPageData(e, deps, secret, false)
		if err != nil {
			return e.InternalServerError("api keys", err)
		}
		data.FlashMessage = "API key created. Copy it now — it will not be shown again."
		return tmpl.renderPage(e.Response, "apikeys", data)
	}
}

func apiKeysRevokeHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if err := e.Request.ParseForm(); err != nil {
			return e.BadRequestError("invalid form", err)
		}
		cfg, err := content.LoadMonmsConfig(deps.SiteAbs)
		if err != nil {
			return e.InternalServerError("config", err)
		}
		if !apikeys.CanManageKeys(e.Auth, cfg) {
			return e.ForbiddenError("api key revoke not allowed", nil)
		}

		id := strings.TrimSpace(e.Request.FormValue("id"))
		if err := apikeys.Revoke(e.App, e.Auth, id); err != nil {
			data, derr := buildAPIKeysPageData(e, deps, err.Error(), true)
			if derr != nil {
				return e.InternalServerError("api keys", derr)
			}
			data.FlashError = true
			return tmpl.renderPage(e.Response, "apikeys", data)
		}

		data, err := buildAPIKeysPageData(e, deps, "", false)
		if err != nil {
			return e.InternalServerError("api keys", err)
		}
		data.FlashMessage = "API key revoked."
		return tmpl.renderPage(e.Response, "apikeys", data)
	}
}

func buildAPIKeysPageData(e *core.RequestEvent, deps Deps, flash string, flashErr bool) (apiKeysPageData, error) {
	cfg, err := content.LoadMonmsConfig(deps.SiteAbs)
	if err != nil {
		return apiKeysPageData{}, err
	}

	base, err := buildPageData(e, deps.SiteAbs, "apikeys", "API Keys")
	if err != nil {
		return apiKeysPageData{}, err
	}

	canCreate := apikeys.CanManageKeys(e.Auth, cfg)
	records, err := apikeys.ListForOwner(e.App, e.Auth)
	if err != nil {
		return apiKeysPageData{}, err
	}

	rows := make([]apiKeyRow, 0, len(records))
	for _, rec := range records {
		rows = append(rows, apiKeyRow{
			ID:         rec.Id,
			Name:       rec.GetString("name"),
			Prefix:     rec.GetString("prefix") + "…",
			Created:    rec.GetDateTime("created").String(),
			LastUsedAt: rec.GetString("lastUsedAt"),
		})
	}

	data := apiKeysPageData{
		PageData:       base,
		Keys:           rows,
		CanCreate:      canCreate,
		PolicyDisabled: !canCreate && e.Auth != nil && !apikeys.IsSuperuserAuth(e.Auth),
	}
	if flash != "" && !flashErr && strings.HasPrefix(flash, "monms_") {
		data.NewKeySecret = flash
	} else if flash != "" {
		data.FlashMessage = flash
		data.FlashError = flashErr
	}
	return data, nil
}
