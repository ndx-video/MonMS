package monmsdash

import (
	"net/http"
	"strings"

	"github.com/monms/monms/internal/content"
	"github.com/monms/monms/internal/monmsroutes"
	"github.com/pocketbase/pocketbase/core"
)

// PageData is shared dashboard template context.
type PageData struct {
	ActivePage   string
	Title        string
	UserEmail    string
	IsSuperuser  bool
	IsPublisher  bool
	SiteURL      string
	AdminURL     string
	FlashMessage string
	FlashError   bool
}

func buildPageData(e *core.RequestEvent, siteAbs, activePage, title string) (PageData, error) {
	cfg, err := content.LoadMonmsConfig(siteAbs)
	if err != nil {
		return PageData{}, err
	}

	data := PageData{
		ActivePage: activePage,
		Title:      title,
		SiteURL:    strings.TrimRight(cfg.SiteURL, "/"),
		AdminURL:   monmsroutes.AdminPath,
	}

	if e.Auth != nil {
		data.UserEmail = e.Auth.GetString("email")
		data.IsSuperuser = true
		data.IsPublisher = content.IsPublisher(data.UserEmail, cfg.PublisherEmails)
	}

	if data.SiteURL == "" {
		data.SiteURL = "/"
	}

	return data, nil
}

func bindLoadAuth(loadAuth func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if loadAuth != nil {
			_ = loadAuth(e)
		}
		return e.Next()
	}
}

func requirePublisherFromSite(siteAbs string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		cfg, err := content.LoadMonmsConfig(siteAbs)
		if err != nil {
			return e.InternalServerError("failed to load config", err)
		}
		return content.RequirePublisher(cfg.PublisherEmails)(e)
	}
}

func requireAuthenticatedRedirect() func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return e.Redirect(http.StatusSeeOther, monmsroutes.AdminPath)
		}
		return e.Next()
	}
}

func requireAuthenticated(next func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return e.Redirect(http.StatusSeeOther, monmsroutes.AdminPath)
		}
		return next(e)
	}
}

func registerDashboardRoutes(se *core.ServeEvent, deps Deps, tmpl *templates) {
	authBind := bindLoadAuth(deps.LoadAuth)
	home := requireAuthenticated(homeHandler(deps, tmpl))

	se.Router.GET("/_monms", home).BindFunc(authBind)
	se.Router.GET(monmsroutes.DashboardHomePath, home).BindFunc(authBind)
}

func homeHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data, err := buildPageData(e, deps.SiteAbs, "home", "Dashboard")
		if err != nil {
			return e.InternalServerError("dashboard", err)
		}
		return tmpl.renderPage(e.Response, "home", data)
	}
}

func registerPublishRoutes(se *core.ServeEvent, deps Deps, tmpl *templates) {
	publisherBind := requirePublisherFromSite(deps.SiteAbs)
	authBind := bindLoadAuth(deps.LoadAuth)
	authRedirect := requireAuthenticatedRedirect()

	se.Router.GET(monmsroutes.PublishPath, publishPageHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(publisherBind)

	se.Router.GET(monmsroutes.PublishDiffPath, publishDiffHandler(deps)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(publisherBind)

	se.Router.POST(monmsroutes.PublishPath, publishPostHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(publisherBind)
}
