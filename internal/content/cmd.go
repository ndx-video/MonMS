package content

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/config"
	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ErrPendingChanges is returned by the diff subcommand when editorial changes exist.
var ErrPendingChanges = errors.New("pending editorial content changes")

// RunCLI is the entry point for the `monms content` subcommand (PUB-03, PUB-09).
func RunCLI(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: monms content <export|import|diff|publish> [--workspace PATH]")
	}

	sub := args[0]

	_, wsAbs, err := config.ResolveWorkspace(args, os.Environ())
	if err != nil {
		return err
	}

	switch sub {
	case "export":
		return runExport(wsAbs)
	case "import":
		return runImport(wsAbs)
	case "diff":
		return runDiff(wsAbs)
	case "publish":
		return runPublishCLI(args[1:], wsAbs)
	default:
		return fmt.Errorf("unknown content subcommand %q (want export, import, diff, or publish)", sub)
	}
}

func bootstrapApp(wsAbs string) (core.App, error) {
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(wsAbs, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})
	schema.RegisterBootstrapHook(app, wsAbs)
	if err := app.Bootstrap(); err != nil {
		return nil, fmt.Errorf("content bootstrap: %w", err)
	}
	return app, nil
}

func runExport(wsAbs string) error {
	app, err := bootstrapApp(wsAbs)
	if err != nil {
		return err
	}
	if err := ExportAll(app, wsAbs); err != nil {
		return err
	}
	slog.Info("content export: complete", "workspace", wsAbs)
	return nil
}

func runImport(wsAbs string) error {
	app, err := bootstrapApp(wsAbs)
	if err != nil {
		return err
	}
	if err := ImportFiles(app, wsAbs); err != nil {
		return err
	}
	slog.Info("content import: complete", "workspace", wsAbs)
	return nil
}

func runDiff(wsAbs string) error {
	app, err := bootstrapApp(wsAbs)
	if err != nil {
		return err
	}

	result, err := DiffExport(app, wsAbs)
	if err != nil {
		return err
	}

	if !result.HasChanges {
		fmt.Println("no pending editorial changes")
		return nil
	}

	fmt.Println("pending editorial changes:")
	for _, change := range result.Changes {
		fmt.Println(" ", change)
	}
	if len(result.Changes) == 0 {
		fmt.Println("  editorial content changed since last publish")
	}
	return ErrPendingChanges
}

func runPublishCLI(args []string, wsAbs string) error {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	var toURL string
	fs.StringVar(&toURL, "to", "", "production base URL (required)")
	if err := fs.Parse(stripWorkspaceFlags(args)); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if toURL == "" {
		return fmt.Errorf("content publish: --to production URL is required")
	}

	token := os.Getenv("MONMS_PUBLISH_TOKEN")
	if token == "" {
		return fmt.Errorf("content publish: missing publish token (set MONMS_PUBLISH_TOKEN)")
	}

	app, err := bootstrapApp(wsAbs)
	if err != nil {
		return err
	}

	snap, err := ExportSnapshot(app, wsAbs)
	if err != nil {
		return err
	}

	payloads := make([]CollectionPayload, len(snap))
	for i, f := range snap {
		payloads[i] = CollectionPayload{
			Collection: f.Collection,
			Records:    f.Records,
		}
	}

	return PublishToProduction(toURL, token, payloads)
}

func stripWorkspaceFlags(args []string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--workspace=") {
			continue
		}
		if arg == "--workspace" {
			if i+1 < len(args) {
				i++
			}
			continue
		}
		out = append(out, arg)
	}
	return out
}
