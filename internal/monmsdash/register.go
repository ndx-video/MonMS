package monmsdash

import (
	"github.com/monms/monms/internal/content"
	"github.com/pocketbase/pocketbase/core"
)

const staticPrefix = "/_monms/static/"

// Deps configures the MonMS operator dashboard.
type Deps struct {
	SiteAbs      string
	PublishToken string
	LoadAuth     func(*core.RequestEvent) error
}

// RegisterRoutes wires the /_monms/* dashboard (HTML, static assets, publish console).
func RegisterRoutes(se *core.ServeEvent, deps Deps) {
	tmpl := mustLoadTemplates()

	se.Router.GET(staticPrefix+"{path...}", staticHandler())

	registerDashboardRoutes(se, deps, tmpl)
	registerDocumentsRoutes(se, deps, tmpl)
	registerPublishRoutes(se, deps, tmpl)
	registerAPIKeysRoutes(se, deps, tmpl)
	registerMCPSettingsRoutes(se, deps, tmpl)
}

// PublishDeps exposes publish business logic to tests without rendering HTML.
func PublishDeps(deps Deps) content.Deps {
	return content.Deps{
		SiteAbs:      deps.SiteAbs,
		PublishToken: deps.PublishToken,
		LoadAuth:     deps.LoadAuth,
	}
}
