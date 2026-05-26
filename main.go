package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/monms/monms/internal/config"
	"github.com/monms/monms/internal/router"
	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/scaffold"
	"github.com/monms/monms/internal/templates"
	"github.com/monms/monms/internal/validate"
	"github.com/monms/monms/internal/workspace"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

var buildMode = "development"

var tplCache = templates.NewCache()

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "init" {
		if err := scaffold.RunInit(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "validate" {
		if err := validate.RunCLI(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	runServe()
}

func runServe() {
	configured, abs, err := config.ResolveWorkspace(os.Args, os.Environ())
	if err != nil {
		fmt.Fprintf(os.Stderr, "workspace: %v\n", err)
		os.Exit(1)
	}

	if err := workspace.ValidateWorkspace(abs); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	slog.Info("workspace configured",
		"path", configured,
		"absolute", abs,
		"mode", buildMode,
	)

	tplCache.SetProductionMode(buildMode == "production")

	if buildMode == "production" {
		if err := templates.StartWatcher(context.Background(), abs, tplCache.Flush); err != nil {
			slog.Error("template watcher failed to start", "err", err)
		}
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: filepath.Join(abs, ".pb_data"),
		DefaultDev:     buildMode != "production",
	})

	schema.RegisterBootstrapHook(app, abs)
	router.RegisterAuthHooks(app)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		router.RegisterRoutes(se, router.Deps{
			WsAbs: abs,
			Cache: tplCache,
			IsDev: buildMode != "production",
		})
		return se.Next()
	})

	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "serve: %v\n", err)
		os.Exit(1)
	}
}
