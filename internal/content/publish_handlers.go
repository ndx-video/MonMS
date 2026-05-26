package content

import (
	_ "embed"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/monms/monms/internal/monmsroutes"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

//go:embed embed/publish.gohtml
var publishPageTemplate string

type publishPageData struct {
	LastPublished   string
	HasChanges      bool
	StatusLabel     string
	Changes         []string
	ChangeGroups    []publishChangeGroup
	ProductionURLSet bool
	SetupMode       bool
	Message         string
	MessageError    bool
}

type publishChangeGroup struct {
	Collection string
	Fields     []publishFieldChange
}

type publishFieldChange struct {
	Field   string
	Summary string
}

func publishPageHandler(deps Deps) func(*core.RequestEvent) error {
	tmpl := template.Must(template.New("publish").Parse(publishPageTemplate))

	return func(e *core.RequestEvent) error {
		cfg, err := LoadMonmsConfig(deps.WsAbs)
		if err != nil {
			return e.InternalServerError("failed to load config", err)
		}

		data, err := buildPublishPageData(e, deps, cfg, "", false)
		if err != nil {
			return e.InternalServerError("publish page", err)
		}

		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.Execute(e.Response, data)
	}
}

func publishDiffHandler(deps Deps) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		diff, err := DiffExport(e.App, deps.WsAbs)
		if err != nil {
			return e.InternalServerError("diff export", err)
		}
		return e.JSON(http.StatusOK, diff)
	}
}

func publishPostHandler(deps Deps) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		cfg, err := LoadMonmsConfig(deps.WsAbs)
		if err != nil {
			return e.InternalServerError("failed to load config", err)
		}

		email := ""
		if e.Auth != nil {
			email = e.Auth.GetString("email")
		}

		if strings.TrimSpace(cfg.ProductionURL) == "" {
			slog.Info("content publish attempt", "email", email, "outcome", "setup_required")
			return e.BadRequestError("production URL not configured", nil)
		}

		if deps.PublishToken == "" {
			slog.Info("content publish attempt", "email", email, "outcome", "token_unset")
			return e.InternalServerError("MONMS_PUBLISH_TOKEN is not configured on staging", nil)
		}

		snap, err := ExportSnapshot(e.App, deps.WsAbs)
		if err != nil {
			return e.InternalServerError("export snapshot", err)
		}

		payloads := snapshotToPayloads(snap)
		if err := PublishToProduction(cfg.ProductionURL, deps.PublishToken, payloads); err != nil {
			slog.Info("content publish attempt", "email", email, "outcome", "failed", "error", err.Error())
			tmpl := template.Must(template.New("publish").Parse(publishPageTemplate))
			data, buildErr := buildPublishPageData(e, deps, cfg, err.Error(), true)
			if buildErr != nil {
				return e.InternalServerError("publish page", buildErr)
			}
			e.Response.WriteHeader(http.StatusBadGateway)
			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return tmpl.Execute(e.Response, data)
		}

		checksum, err := ChecksumExport(payloads)
		if err != nil {
			return e.InternalServerError("checksum", err)
		}

		collections := make([]string, len(payloads))
		for i, p := range payloads {
			collections[i] = p.Collection
		}

		state := PublishState{
			Checksum:    checksum,
			PublishedAt: time.Now().UTC().Format(time.RFC3339),
			Collections: collections,
		}
		if err := WritePublishState(deps.WsAbs, state); err != nil {
			return e.InternalServerError("write publish state", err)
		}

		slog.Info("content publish attempt", "email", email, "outcome", "success")
		tmpl := template.Must(template.New("publish").Parse(publishPageTemplate))
		data, err := buildPublishPageData(e, deps, cfg, "Content published successfully.", false)
		if err != nil {
			return e.InternalServerError("publish page", err)
		}
		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.Execute(e.Response, data)
	}
}

func buildPublishPageData(e *core.RequestEvent, deps Deps, cfg MonmsConfig, message string, messageError bool) (publishPageData, error) {
	productionURLSet := strings.TrimSpace(cfg.ProductionURL) != ""
	data := publishPageData{
		ProductionURLSet: productionURLSet,
		SetupMode:        !productionURLSet,
		Message:          message,
		MessageError:     messageError,
	}

	if !productionURLSet {
		return data, nil
	}

	state, err := ReadPublishState(deps.WsAbs)
	if err != nil {
		return publishPageData{}, err
	}
	if state.PublishedAt != "" {
		data.LastPublished = state.PublishedAt
	}

	diff, err := DiffExport(e.App, deps.WsAbs)
	if err != nil {
		return publishPageData{}, err
	}

	data.HasChanges = diff.HasChanges
	if diff.HasChanges {
		data.StatusLabel = "You have unpublished changes."
		data.Changes = diff.Changes
		data.ChangeGroups = groupDiffChanges(diff.Changes)
	} else {
		data.StatusLabel = "Production matches your staging content."
	}

	return data, nil
}

func snapshotToPayloads(snap []CollectionFile) []CollectionPayload {
	payloads := make([]CollectionPayload, len(snap))
	for i, f := range snap {
		payloads[i] = CollectionPayload{
			Collection: f.Collection,
			Records:    f.Records,
		}
	}
	return payloads
}

func groupDiffChanges(changes []string) []publishChangeGroup {
	byColl := make(map[string][]publishFieldChange)
	var order []string

	for _, line := range changes {
		coll, field, summary, ok := parseDiffLine(line)
		if !ok {
			continue
		}
		if _, exists := byColl[coll]; !exists {
			order = append(order, coll)
		}
		byColl[coll] = append(byColl[coll], publishFieldChange{Field: field, Summary: summary})
	}

	groups := make([]publishChangeGroup, 0, len(order))
	for _, coll := range order {
		groups = append(groups, publishChangeGroup{
			Collection: coll,
			Fields:     byColl[coll],
		})
	}
	return groups
}

func parseDiffLine(line string) (collection, field, summary string, ok bool) {
	// hero_content/homepage/title: "a" → "b"
	parts := strings.SplitN(line, ": ", 2)
	if len(parts) != 2 {
		return "", "", "", false
	}
	path := parts[0]
	summary = parts[1]
	segments := strings.Split(path, "/")
	if len(segments) == 2 {
		return segments[0], segments[1], summary, true
	}
	if len(segments) < 3 {
		return "", "", "", false
	}
	return segments[0], segments[2], summary, true
}

func requirePublisherFromWorkspace(wsAbs string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		cfg, err := LoadMonmsConfig(wsAbs)
		if err != nil {
			return e.InternalServerError("failed to load config", err)
		}
		return RequirePublisher(cfg.PublisherEmails)(e)
	}
}

func bindLoadAuth(loadAuth func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if loadAuth != nil {
			_ = loadAuth(e)
		}
		return e.Next()
	}
}

func registerPublishRoutes(se *core.ServeEvent, deps Deps) {
	publisherBind := requirePublisherFromWorkspace(deps.WsAbs)
	authBind := bindLoadAuth(deps.LoadAuth)

	se.Router.GET(monmsroutes.PublishPath, publishPageHandler(deps)).
		BindFunc(authBind).
		Bind(apis.RequireSuperuserAuth()).
		BindFunc(publisherBind)

	se.Router.GET(monmsroutes.PublishDiffPath, publishDiffHandler(deps)).
		BindFunc(authBind).
		Bind(apis.RequireSuperuserAuth()).
		BindFunc(publisherBind)

	se.Router.POST(monmsroutes.PublishPath, publishPostHandler(deps)).
		BindFunc(authBind).
		Bind(apis.RequireSuperuserAuth()).
		BindFunc(publisherBind)
}
