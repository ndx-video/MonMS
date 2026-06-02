package monmsdash

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/monms/monms/internal/documents"
	"github.com/monms/monms/internal/monmsroutes"
	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase/core"
)

const (
	doctreesMigrateCopyPath    = monmsroutes.DoctreesPath + "/migrate/copy"
	doctreesMigrateRescanPath  = monmsroutes.DoctreesPath + "/migrate/rescan"
	doctreesMigrateConfirmPath = monmsroutes.DoctreesPath + "/migrate/confirm"
	doctreesMigrateCancelPath  = monmsroutes.DoctreesPath + "/migrate/cancel"
	doctreesMigratePrunePath   = monmsroutes.DoctreesPath + "/migrate/prune"
)

type doctreesPageData struct {
	PageData
	Forest           documents.Forest
	CanMigrate       bool
	MigrateStep      string
	SourceRoot       string
	DoctreeStub      string
	BindingCandidates  []documents.BindingCandidate
	StaleBindings      []string
	AlignmentIssues    []documents.DoctreeAlignmentIssue
	FilesCopied        int
	Force              bool
}

func registerDoctreesRoutes(se *core.ServeEvent, deps Deps, tmpl *templates) {
	authBind := bindLoadAuth(deps.LoadAuth)
	authRedirect := requireAuthenticatedRedirect()
	publisherBind := requirePublisherFromSite(deps.SiteAbs)

	se.Router.GET(monmsroutes.DoctreesPath, doctreesPageHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect)

	se.Router.GET(monmsroutes.DocumentsPath, documentsRedirectHandler()).
		BindFunc(authBind).
		BindFunc(authRedirect)

	se.Router.POST(doctreesMigrateCopyPath, doctreesMigrateCopyHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(publisherBind)

	se.Router.POST(doctreesMigrateRescanPath, doctreesMigrateRescanHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(publisherBind)

	se.Router.POST(doctreesMigrateConfirmPath, doctreesMigrateConfirmHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(publisherBind)

	se.Router.POST(doctreesMigrateCancelPath, doctreesMigrateCancelHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(publisherBind)

	se.Router.POST(doctreesMigratePrunePath, doctreesMigratePruneHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(publisherBind)
}

func documentsRedirectHandler() func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		return e.Redirect(http.StatusMovedPermanently, monmsroutes.DoctreesPath)
	}
}

func doctreesPageHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data, err := buildDoctreesPageData(e, deps, doctreesFormState{})
		if err != nil {
			return e.InternalServerError("doctrees", err)
		}
		return tmpl.renderPage(e.Response, "doctrees", data)
	}
}

type doctreesFormState struct {
	SourceRoot        string
	DoctreeStub       string
	MigrateStep       string
	Candidates        []documents.BindingCandidate
	StaleBindings     []string
	FilesCopied       int
	Force             bool
	FlashMessage      string
	FlashError        bool
}

func buildDoctreesPageData(e *core.RequestEvent, deps Deps, form doctreesFormState) (doctreesPageData, error) {
	base, err := buildPageData(e, deps.SiteAbs, "doctrees", "Doctrees")
	if err != nil {
		return doctreesPageData{}, err
	}
	if form.FlashMessage != "" {
		base.FlashMessage = form.FlashMessage
		base.FlashError = form.FlashError
	}

	forest, err := documents.BuildForest(e.App, deps.SiteAbs)
	if err != nil {
		return doctreesPageData{}, err
	}

	alignmentIssues, err := documents.AuditDoctreeAlignment(deps.SiteAbs)
	if err != nil {
		return doctreesPageData{}, err
	}

	step := form.MigrateStep
	if step == "" {
		step = "start"
	}

	stub := form.DoctreeStub
	source := form.SourceRoot
	candidates := form.Candidates
	stale := form.StaleBindings
	filesCopied := form.FilesCopied

	if step == "start" {
		if pending, ok, err := documents.ReadPendingMigrate(deps.SiteAbs); err == nil && ok {
			stub = pending.Stub
			source = pending.SourceRoot
			filesCopied = pending.FilesCopied
			step = "confirm"
			candidates, stale, err = discoverWithStale(deps.SiteAbs, stub)
			if err != nil {
				return doctreesPageData{}, err
			}
		}
	}

	return doctreesPageData{
		PageData:          base,
		Forest:            forest,
		CanMigrate:        base.IsPublisher,
		MigrateStep:       step,
		SourceRoot:        source,
		DoctreeStub:       stub,
		BindingCandidates: candidates,
		StaleBindings:     stale,
		AlignmentIssues:   alignmentIssues,
		FilesCopied:       filesCopied,
		Force:             form.Force,
	}, nil
}

func discoverWithStale(siteAbs, stub string) ([]documents.BindingCandidate, []string, error) {
	candidates, err := documents.DiscoverLeafCollections(siteAbs, stub)
	if err != nil {
		return nil, nil, err
	}
	stale, err := documents.StaleBindingsUnderStub(siteAbs, stub, candidates)
	if err != nil {
		return nil, nil, err
	}
	return candidates, stale, nil
}

func doctreesMigrateCopyHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if err := e.Request.ParseForm(); err != nil {
			return e.BadRequestError("invalid form", err)
		}
		source := strings.TrimSpace(e.Request.FormValue("source_root"))
		stub := strings.TrimSpace(e.Request.FormValue("doctree_stub"))
		state := doctreesFormState{SourceRoot: source, DoctreeStub: stub}

		if source == "" {
			state.FlashMessage = "Source path is required for Migrate."
			state.FlashError = true
			return renderDoctrees(e, deps, tmpl, state)
		}

		copyResult, err := documents.CopySourceTree(deps.SiteAbs, source, stub)
		if err != nil {
			state.FlashMessage = err.Error()
			state.FlashError = true
			return renderDoctrees(e, deps, tmpl, state)
		}

		if err := documents.WritePendingMigrate(deps.SiteAbs, documents.PendingMigrate{
			Stub:        copyResult.Stub,
			SourceRoot:  copyResult.SourceRoot,
			FilesCopied: copyResult.FilesCopied,
			CopiedAt:    copyResult.CopiedAt,
		}); err != nil {
			return e.InternalServerError("pending migrate", err)
		}

		return finishDiscoverStep(e, deps, tmpl, copyResult.Stub, copyResult.SourceRoot, copyResult.FilesCopied,
			fmt.Sprintf("Copied %d file(s). Review bindings below.", copyResult.FilesCopied))
	}
}

func doctreesMigrateRescanHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if err := e.Request.ParseForm(); err != nil {
			return e.BadRequestError("invalid form", err)
		}
		stub := strings.TrimSpace(e.Request.FormValue("doctree_stub"))
		source := strings.TrimSpace(e.Request.FormValue("source_root"))

		if err := documents.ValidateDoctreeStub(stub); err != nil {
			state := doctreesFormState{DoctreeStub: stub, SourceRoot: source, FlashMessage: err.Error(), FlashError: true}
			return renderDoctrees(e, deps, tmpl, state)
		}

		pending, ok, _ := documents.ReadPendingMigrate(deps.SiteAbs)
		if !ok {
			pending = documents.PendingMigrate{Stub: stub, SourceRoot: source}
		} else {
			pending.Stub = stub
			if source != "" {
				pending.SourceRoot = source
			}
		}
		if err := documents.WritePendingMigrate(deps.SiteAbs, pending); err != nil {
			return e.InternalServerError("pending migrate", err)
		}

		filesCopied := pending.FilesCopied
		return finishDiscoverStep(e, deps, tmpl, stub, pending.SourceRoot, filesCopied, "Binding list refreshed.")
	}
}

func finishDiscoverStep(e *core.RequestEvent, deps Deps, tmpl *templates, stub, source string, filesCopied int, flash string) error {
	candidates, stale, err := discoverWithStale(deps.SiteAbs, stub)
	state := doctreesFormState{
		MigrateStep:   "confirm",
		DoctreeStub:   stub,
		SourceRoot:    source,
		FilesCopied:   filesCopied,
		Candidates:    candidates,
		StaleBindings: stale,
		FlashMessage:  flash,
	}
	if err != nil {
		state.MigrateStep = "start"
		state.FlashMessage = err.Error()
		state.FlashError = true
	}
	if len(candidates) == 0 && err == nil {
		state.FlashMessage = "No leaf markdown folders found under " + stub + "/."
		state.FlashError = true
	}
	return renderDoctrees(e, deps, tmpl, state)
}

func doctreesMigrateConfirmHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if err := e.Request.ParseForm(); err != nil {
			return e.BadRequestError("invalid form", err)
		}

		stub := strings.TrimSpace(e.Request.FormValue("doctree_stub"))
		force := e.Request.FormValue("force") == "on"

		pending, ok, err := documents.ReadPendingMigrate(deps.SiteAbs)
		if err != nil || !ok {
			state := doctreesFormState{
				FlashMessage: "No pending migration. Run Migrate or Re-scan first.",
				FlashError:   true,
			}
			return renderDoctrees(e, deps, tmpl, state)
		}
		if stub == "" {
			stub = pending.Stub
		}

		candidates, _, err := discoverWithStale(deps.SiteAbs, stub)
		if err != nil {
			state := doctreesFormState{FlashMessage: err.Error(), FlashError: true}
			return renderDoctrees(e, deps, tmpl, state)
		}
		if len(candidates) == 0 {
			state := doctreesFormState{FlashMessage: "No bindings to confirm.", FlashError: true}
			return renderDoctrees(e, deps, tmpl, state)
		}

		result, err := documents.FinalizeBindings(deps.SiteAbs, candidates, documents.FinalizeOptions{Force: force})
		if err != nil {
			state := doctreesFormState{
				MigrateStep:  "confirm",
				DoctreeStub:  stub,
				SourceRoot:   pending.SourceRoot,
				Candidates:   candidates,
				Force:        force,
				FlashMessage: err.Error(),
				FlashError:   true,
			}
			return renderDoctrees(e, deps, tmpl, state)
		}

		if err := schema.ImportSiteCollections(e.App, deps.SiteAbs); err != nil {
			state := doctreesFormState{
				MigrateStep:  "confirm",
				DoctreeStub:  stub,
				Candidates:   candidates,
				FlashMessage: "Schema import failed: " + err.Error(),
				FlashError:   true,
			}
			return renderDoctrees(e, deps, tmpl, state)
		}

		dtResults, syncErr := documents.SyncAfterBindDt(e.App, deps.SiteAbs)
		if syncErr != nil {
			state := doctreesFormState{
				MigrateStep:  "confirm",
				DoctreeStub:  stub,
				Candidates:   candidates,
				FlashMessage: "Sync failed: " + syncErr.Error(),
				FlashError:   true,
			}
			return renderDoctrees(e, deps, tmpl, state)
		}

		if err := documents.UpsertDtTreeFromConfirm(e.App, stub, pending.SourceRoot, candidates); err != nil {
			state := doctreesFormState{
				MigrateStep:  "confirm",
				DoctreeStub:  stub,
				Candidates:   candidates,
				FlashMessage: "dt_trees registry failed: " + err.Error(),
				FlashError:   true,
			}
			return renderDoctrees(e, deps, tmpl, state)
		}

		_ = documents.ClearPendingMigrate(deps.SiteAbs)

		var recCreated, recUpdated, filesCreated, filesUpdated int
		for _, r := range dtResults {
			recCreated += r.RecordsCreated
			recUpdated += r.RecordsUpdated
			filesCreated += r.FilesCreated
			filesUpdated += r.FilesUpdated
		}

		msg := fmt.Sprintf(
			"Bindings confirmed: %d file(s) updated across %d collection(s); dt sync: %d record(s) created, %d updated, %d file(s) created, %d updated.",
			result.TotalBound, len(result.Collections), recCreated, recUpdated, filesCreated, filesUpdated,
		)
		if len(result.RetiredCollections) > 0 {
			msg += " Retire old PocketBase collections in admin: " + strings.Join(result.RetiredCollections, ", ") + "."
		}

		state := doctreesFormState{
			MigrateStep:  "start",
			DoctreeStub:  stub,
			FlashMessage: msg,
		}
		return renderDoctrees(e, deps, tmpl, state)
	}
}

func doctreesMigrateCancelHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		pending, ok, _ := documents.ReadPendingMigrate(deps.SiteAbs)
		_ = documents.ClearPendingMigrate(deps.SiteAbs)
		stub := ""
		if ok {
			stub = pending.Stub
		}
		msg := "Migration cancelled."
		if stub != "" {
			msg = fmt.Sprintf("Migration cancelled. Copied files kept under %s/.", stub)
		}
		state := doctreesFormState{
			MigrateStep:  "start",
			DoctreeStub:  stub,
			FlashMessage: msg,
		}
		return renderDoctrees(e, deps, tmpl, state)
	}
}

func doctreesMigratePruneHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if err := e.Request.ParseForm(); err != nil {
			return e.BadRequestError("invalid form", err)
		}
		stub := strings.TrimSpace(e.Request.FormValue("doctree_stub"))
		pending, ok, _ := documents.ReadPendingMigrate(deps.SiteAbs)
		if stub == "" && ok {
			stub = pending.Stub
		}
		if stub != "" {
			if err := documents.PruneCopiedTree(deps.SiteAbs, stub); err != nil {
				state := doctreesFormState{FlashMessage: err.Error(), FlashError: true}
				return renderDoctrees(e, deps, tmpl, state)
			}
		}
		_ = documents.ClearPendingMigrate(deps.SiteAbs)
		state := doctreesFormState{
			MigrateStep:  "start",
			FlashMessage: "Migration cancelled; copied tree removed.",
		}
		return renderDoctrees(e, deps, tmpl, state)
	}
}

func renderDoctrees(e *core.RequestEvent, deps Deps, tmpl *templates, state doctreesFormState) error {
	data, err := buildDoctreesPageData(e, deps, state)
	if err != nil {
		return e.InternalServerError("doctrees", err)
	}
	return tmpl.renderPage(e.Response, "doctrees", data)
}
