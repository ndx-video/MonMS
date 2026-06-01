package monmsdash

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/monms/monms/internal/content"
	"github.com/pocketbase/pocketbase/core"
)

type publishPageData struct {
	PageData
	LastPublished    string
	HasChanges       bool
	StatusLabel      string
	Changes          []string
	ChangeGroups     []publishChangeGroup
	ProductionURLSet bool
	SetupMode        bool
	Message          string
	MessageError     bool
}

type publishChangeGroup struct {
	Collection string
	Fields     []publishFieldChange
}

type publishFieldChange struct {
	Field   string
	Summary string
}

func publishPageHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		cfg, err := content.LoadMonmsConfig(deps.SiteAbs)
		if err != nil {
			return e.InternalServerError("failed to load config", err)
		}

		base, err := buildPageData(e, deps.SiteAbs, "publish", "Publish to live")
		if err != nil {
			return e.InternalServerError("dashboard", err)
		}

		data, err := buildPublishPageData(e, deps, cfg, base, "", false)
		if err != nil {
			return e.InternalServerError("publish page", err)
		}

		return tmpl.renderPage(e.Response, "publish", data)
	}
}

func publishDiffHandler(deps Deps) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		diff, err := content.DiffExport(e.App, deps.SiteAbs)
		if err != nil {
			return e.InternalServerError("diff export", err)
		}
		return e.JSON(http.StatusOK, diff)
	}
}

func publishPostHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		cfg, err := content.LoadMonmsConfig(deps.SiteAbs)
		if err != nil {
			return e.InternalServerError("failed to load config", err)
		}

		base, err := buildPageData(e, deps.SiteAbs, "publish", "Publish to live")
		if err != nil {
			return e.InternalServerError("dashboard", err)
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

		snap, err := content.ExportSnapshot(e.App, deps.SiteAbs)
		if err != nil {
			return e.InternalServerError("export snapshot", err)
		}

		payloads := snapshotToPayloads(snap)
		if err := content.PublishToProduction(cfg.ProductionURL, deps.PublishToken, payloads); err != nil {
			slog.Info("content publish attempt", "email", email, "outcome", "failed", "error", err.Error())
			data, buildErr := buildPublishPageData(e, deps, cfg, base, err.Error(), true)
			if buildErr != nil {
				return e.InternalServerError("publish page", buildErr)
			}
			e.Response.WriteHeader(http.StatusBadGateway)
			return tmpl.renderPage(e.Response, "publish", data)
		}

		checksum, err := content.ChecksumExport(payloads)
		if err != nil {
			return e.InternalServerError("checksum", err)
		}

		collections := make([]string, len(payloads))
		for i, p := range payloads {
			collections[i] = p.Collection
		}

		state := content.PublishState{
			Checksum:    checksum,
			PublishedAt: time.Now().UTC().Format(time.RFC3339),
			Collections: collections,
		}
		if err := content.WritePublishState(deps.SiteAbs, state); err != nil {
			return e.InternalServerError("write publish state", err)
		}

		slog.Info("content publish attempt", "email", email, "outcome", "success")
		data, err := buildPublishPageData(e, deps, cfg, base, "Content published successfully.", false)
		if err != nil {
			return e.InternalServerError("publish page", err)
		}
		return tmpl.renderPage(e.Response, "publish", data)
	}
}

func buildPublishPageData(e *core.RequestEvent, deps Deps, cfg content.MonmsConfig, base PageData, message string, messageError bool) (publishPageData, error) {
	productionURLSet := strings.TrimSpace(cfg.ProductionURL) != ""
	data := publishPageData{
		PageData:         base,
		ProductionURLSet: productionURLSet,
		SetupMode:        !productionURLSet,
		Message:          message,
		MessageError:     messageError,
	}

	if message != "" {
		data.FlashMessage = message
		data.FlashError = messageError
	}

	if !productionURLSet {
		return data, nil
	}

	state, err := content.ReadPublishState(deps.SiteAbs)
	if err != nil {
		return publishPageData{}, err
	}
	if state.PublishedAt != "" {
		data.LastPublished = state.PublishedAt
	}

	diff, err := content.DiffExport(e.App, deps.SiteAbs)
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

func snapshotToPayloads(snap []content.CollectionFile) []content.CollectionPayload {
	payloads := make([]content.CollectionPayload, len(snap))
	for i, f := range snap {
		payloads[i] = content.CollectionPayload{
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
