package monmsdash

import (
	"time"

	"github.com/monms/monms/internal/monmsroutes"
	"github.com/monms/monms/internal/restart"
	"github.com/monms/monms/internal/stop"
	"github.com/monms/monms/internal/systemstatus"
	"github.com/pocketbase/pocketbase/core"
)

const lifecycleDelay = 200 * time.Millisecond

type systemPageData struct {
	PageData
	Status systemstatus.View
}

var (
	restartServeFn  = defaultRestartServe
	shutdownServeFn = defaultShutdownServe
)

// SetLifecycleHooksForTest swaps lifecycle handlers for tests. Returns restore.
func SetLifecycleHooksForTest(restart func(Deps) error, shutdown func() error) func() {
	prevR, prevS := restartServeFn, shutdownServeFn
	if restart != nil {
		restartServeFn = restart
	}
	if shutdown != nil {
		shutdownServeFn = shutdown
	}
	return func() {
		restartServeFn = prevR
		shutdownServeFn = prevS
	}
}

func defaultRestartServe(deps Deps) error {
	if restart.ShouldRestartDetached(deps.SiteAbs) {
		return restart.RestartDetached(deps.SiteAbs, deps.ServeArgs)
	}
	return restart.ExecCurrentProcess()
}

func defaultShutdownServe() error {
	return stop.ShutdownAll()
}

func registerSystemRoutes(se *core.ServeEvent, deps Deps, tmpl *templates) {
	authBind := bindLoadAuth(deps.LoadAuth)
	authRedirect := requireAuthenticatedRedirect()
	super := requireSuperuser()

	se.Router.GET(monmsroutes.SystemPath, systemGetHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(super)

	se.Router.POST(monmsroutes.SystemRestartPath, systemRestartHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(super)

	se.Router.POST(monmsroutes.SystemShutdownPath, systemShutdownHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(super)
}

func systemGetHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data, err := buildSystemPageData(e, deps, "", false)
		if err != nil {
			return e.InternalServerError("system", err)
		}
		return tmpl.renderPage(e.Response, "system", data)
	}
}

func systemRestartHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data, err := buildSystemPageData(e, deps, "Restarting MonMS…", false)
		if err != nil {
			return e.InternalServerError("system", err)
		}
		if data.Status.LifecycleSupported {
			scheduleLifecycle(func() { _ = restartServeFn(deps) })
		}
		return tmpl.renderPage(e.Response, "system", data)
	}
}

func systemShutdownHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data, err := buildSystemPageData(e, deps, "Shutting down MonMS…", false)
		if err != nil {
			return e.InternalServerError("system", err)
		}
		if data.Status.LifecycleSupported {
			scheduleLifecycle(func() { _ = shutdownServeFn() })
		}
		return tmpl.renderPage(e.Response, "system", data)
	}
}

func buildSystemPageData(e *core.RequestEvent, deps Deps, flash string, flashErr bool) (systemPageData, error) {
	base, err := buildPageData(e, deps.SiteAbs, "system", "System")
	if err != nil {
		return systemPageData{}, err
	}
	if flash != "" {
		base.FlashMessage = flash
		base.FlashError = flashErr
	}

	status, err := systemstatus.Snapshot(
		deps.SiteAbs,
		deps.ServeArgs,
		deps.BuildMode,
		deps.PublishToken != "",
	)
	if err != nil {
		return systemPageData{}, err
	}

	return systemPageData{
		PageData: base,
		Status:   status,
	}, nil
}

func scheduleLifecycle(fn func()) {
	go func() {
		time.Sleep(lifecycleDelay)
		fn()
	}()
}
