package monmsdash

import (
	"github.com/monms/monms/internal/documents"
	"github.com/monms/monms/internal/monmsroutes"
	"github.com/pocketbase/pocketbase/core"
)

type documentsPageData struct {
	PageData
	Forest documents.Forest
}

func registerDocumentsRoutes(se *core.ServeEvent, deps Deps, tmpl *templates) {
	authBind := bindLoadAuth(deps.LoadAuth)
	authRedirect := requireAuthenticatedRedirect()

	se.Router.GET(monmsroutes.DocumentsPath, documentsPageHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect)
}

func documentsPageHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		base, err := buildPageData(e, deps.SiteAbs, "documents", "Documents")
		if err != nil {
			return e.InternalServerError("dashboard", err)
		}

		forest, err := documents.BuildForest(e.App, deps.SiteAbs)
		if err != nil {
			return e.InternalServerError("documents forest", err)
		}

		data := documentsPageData{
			PageData: base,
			Forest:   forest,
		}
		return tmpl.renderPage(e.Response, "documents", data)
	}
}
