package content

import (
	"fmt"
	"net/http"

	"github.com/monms/monms/internal/monmsroutes"
	"github.com/pocketbase/pocketbase/core"
)

const maxRecordsPerCollection = 1000

var deniedSystemCollections = map[string]struct{}{
	"_superusers": {},
	"users":       {},
}

// Deps configures content HTTP routes.
type Deps struct {
	SiteAbs        string
	PublishToken string
	// LoadAuth hydrates e.Auth from browser session (e.g. monms_auth cookie) before RequireSuperuserAuth.
	LoadAuth func(*core.RequestEvent) error
}

// ImportReport summarizes a production import request.
type ImportReport struct {
	Upserted    int      `json:"upserted"`
	Collections int      `json:"collections"`
	Warnings    []string `json:"warnings,omitempty"`
}

type importRequest struct {
	Collections []importCollection `json:"collections"`
}

type importCollection struct {
	Name    string           `json:"name"`
	Records []map[string]any `json:"records"`
}

// RegisterRoutes wires MonMS content JSON API routes on the PocketBase router.
// Operator HTML (dashboard, publish console) lives in internal/monmsdash.
func RegisterRoutes(se *core.ServeEvent, deps Deps) {
	se.Router.POST(monmsroutes.ContentImportPath, importHandler(deps)).
		BindFunc(RequirePublishToken(deps.PublishToken))
}

func importHandler(deps Deps) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var body importRequest
		if err := e.BindBody(&body); err != nil {
			return e.BadRequestError("invalid JSON", err)
		}

		payloads, err := payloadsFromImportRequest(body)
		if err != nil {
			return e.BadRequestError("import failed", err)
		}

		if err := ImportPayload(e.App, deps.SiteAbs, payloads); err != nil {
			return e.BadRequestError("import failed", err)
		}

		report := ImportReport{
			Upserted:    countRecords(payloads),
			Collections: len(payloads),
		}
		return e.JSON(http.StatusOK, report)
	}
}

func payloadsFromImportRequest(body importRequest) ([]CollectionPayload, error) {
	if len(body.Collections) == 0 {
		return nil, nil
	}

	payloads := make([]CollectionPayload, 0, len(body.Collections))
	for _, coll := range body.Collections {
		name := coll.Name
		if name == "" {
			return nil, fmt.Errorf("content import: missing collection name")
		}
		if _, denied := deniedSystemCollections[name]; denied {
			return nil, fmt.Errorf("content import: collection %q is not allowed", name)
		}
		if len(coll.Records) > maxRecordsPerCollection {
			return nil, fmt.Errorf("content import: collection %q exceeds %d records", name, maxRecordsPerCollection)
		}
		payloads = append(payloads, CollectionPayload{
			Collection: name,
			Records:    coll.Records,
		})
	}
	return payloads, nil
}

func countRecords(payloads []CollectionPayload) int {
	n := 0
	for _, p := range payloads {
		n += len(p.Records)
	}
	return n
}
