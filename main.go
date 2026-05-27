package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/monms/monms/internal/cli"
	"github.com/monms/monms/internal/cli/prompt"
	"github.com/monms/monms/internal/config"
	"github.com/monms/monms/internal/content"
	"github.com/monms/monms/internal/daemon"
	"github.com/monms/monms/internal/router"
	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/scaffold"
	"github.com/monms/monms/internal/stop"
	"github.com/monms/monms/internal/templates"
	"github.com/monms/monms/internal/validate"
	"github.com/monms/monms/internal/site"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

var buildMode = "development"

var tplCache = templates.NewCache()

func main() {
	args := os.Args[1:]

	if len(args) >= 2 && args[0] == "content" {
		if _, wantHelp := cli.ParseHelpRequest(args[1:]); wantHelp {
			if text, ok := cli.ContentSubcommandHelp(args[1]); ok {
				fmt.Print(text)
				return
			}
			cli.PrintHelp("content")
			return
		}
	}

	if len(args) >= 2 && args[0] == "site" {
		if _, wantHelp := cli.ParseHelpRequest(args[1:]); wantHelp {
			if text, ok := cli.SiteSubcommandHelp(args[1]); ok {
				fmt.Print(text)
				return
			}
			cli.PrintHelp("site")
			return
		}
	}

	if cmd, wantHelp := cli.ParseHelpRequest(args); wantHelp {
		if cmd == "" || cli.IsMonmsCommand(cmd) {
			cli.PrintHelp(cmd)
			return
		}
	}

	if len(args) >= 1 {
		switch args[0] {
		case "init":
			outcome, err := scaffold.RunInitCLI(args[1:], &prompt.Stdio)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			switch outcome.StartMode {
			case scaffold.StartBackground:
				if err := startDaemon(outcome.SiteAbs, []string{"serve", "-d"}); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			case scaffold.StartForeground:
				if err := os.Setenv("MONMS_SITE", outcome.SiteAbs); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				runServeAt(outcome.SiteAbs)
			}
			return
		case "validate":
			if err := validate.RunCLI(args[1:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "content":
			if err := content.RunCLI(args[1:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "stop":
			if err := stop.RunCLI(args[1:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "site":
			if err := site.RunCLI(args[1:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
	}

	runServe()
}

func runServe() {
	res, err := config.ResolveSiteMeta(os.Args[1:], os.Environ())
	if err != nil {
		fmt.Fprintf(os.Stderr, "site: %v\n", err)
		os.Exit(1)
	}

	outcome, err := site.EnsureReady(res, &prompt.Stdio)
	if err != nil {
		if errors.Is(err, site.ErrDeclined) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch outcome.StartMode {
	case scaffold.StartNone:
		if outcome.Scaffolded {
			return
		}
	case scaffold.StartBackground:
		if err := startDaemon(outcome.SiteAbs, os.Args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	runServeAt(outcome.SiteAbs)
}

func startDaemon(siteAbs string, args []string) error {
	slog.Info("starting daemon",
		"path", siteAbs,
		"mode", buildMode,
	)
	return daemon.Start(siteAbs, args)
}

func runServeAt(abs string) {
	args := os.Args[1:]

	if daemon.ShouldDetach(args) {
		if err := startDaemon(abs, args); err != nil {
			fmt.Fprintf(os.Stderr, "daemon: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := content.ApplyShapeSyncFromSite(abs); err != nil {
		fmt.Fprintf(os.Stderr, "site sync: %v\n", err)
		os.Exit(1)
	}

	if err := os.Setenv("MONMS_SITE", abs); err != nil {
		fmt.Fprintf(os.Stderr, "site env: %v\n", err)
		os.Exit(1)
	}
	os.Args = append([]string{os.Args[0]}, config.StripSiteFlags(os.Args[1:])...)
	os.Args = append([]string{os.Args[0]}, cli.EnsureServeSubcommand(os.Args[1:])...)

	serveArgs, err := content.ApplyServeConfigFromSite(abs, os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "monms config: %v\n", err)
		os.Exit(1)
	}
	os.Args = append([]string{os.Args[0]}, serveArgs...)

	configured := abs
	if res, err := config.ResolveSiteMeta(os.Args[1:], os.Environ()); err == nil {
		configured = res.Configured
	}

	slog.Info("site configured",
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
		router.RegisterAdminUIExtension(se)
		content.RegisterRoutes(se, content.Deps{
			SiteAbs:      abs,
			PublishToken: os.Getenv("MONMS_PUBLISH_TOKEN"),
			LoadAuth:     router.LoadAuthFromCookie,
		})
		router.RegisterRoutes(se, router.Deps{
			SiteAbs: abs,
			Cache:   tplCache,
			IsDev:   buildMode != "production",
		})
		return se.Next()
	})

	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "serve: %v\n", err)
		os.Exit(1)
	}
}
