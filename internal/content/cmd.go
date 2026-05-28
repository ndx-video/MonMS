package content

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/monms/monms/internal/cli"
	"github.com/monms/monms/internal/config"
	"github.com/monms/monms/internal/logging"
	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ErrPendingChanges is returned by the diff subcommand when editorial changes exist.
var ErrPendingChanges = errors.New("pending editorial content changes")

// RunCLI is the entry point for the `monms content` subcommand (PUB-03, PUB-09).
func RunCLI(args []string) error {
	if len(args) == 0 {
		cli.PrintHelp("content")
		return nil
	}

	sub := args[0]

	_, siteAbs, err := config.ResolveSite(args, os.Environ())
	if err != nil {
		return err
	}

	if _, err := logging.Configure(siteAbs); err != nil {
		return err
	}

	switch sub {
	case "export":
		return runExport(siteAbs)
	case "import":
		return runImport(siteAbs)
	case "diff":
		return runDiff(siteAbs)
	case "publish":
		return runPublishCLI(args[1:], siteAbs)
	default:
		return fmt.Errorf("unknown content subcommand %q (want export, import, diff, or publish)", sub)
	}
}

func bootstrapApp(siteAbs string) (core.App, error) {
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(siteAbs, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})
	schema.RegisterBootstrapHook(app, siteAbs)
	if err := app.Bootstrap(); err != nil {
		return nil, fmt.Errorf("content bootstrap: %w", err)
	}
	return app, nil
}

func runExport(siteAbs string) error {
	app, err := bootstrapApp(siteAbs)
	if err != nil {
		return err
	}
	if err := ExportAll(app, siteAbs); err != nil {
		return err
	}
	slog.Info("content export: complete", "site", siteAbs)
	return nil
}

func runImport(siteAbs string) error {
	app, err := bootstrapApp(siteAbs)
	if err != nil {
		return err
	}
	if err := ImportFiles(app, siteAbs); err != nil {
		return err
	}
	slog.Info("content import: complete", "site", siteAbs)
	return nil
}

func runDiff(siteAbs string) error {
	app, err := bootstrapApp(siteAbs)
	if err != nil {
		return err
	}

	result, err := DiffExport(app, siteAbs)
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

func runPublishCLI(args []string, siteAbs string) error {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		if text, ok := cli.ContentSubcommandHelp("publish"); ok {
			fmt.Print(text)
		}
	}
	var toURL string
	fs.StringVar(&toURL, "to", "", "production base URL (required)")
	if err := fs.Parse(config.StripSiteFlags(args)); err != nil {
		if err == flag.ErrHelp {
			fs.Usage()
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

	app, err := bootstrapApp(siteAbs)
	if err != nil {
		return err
	}

	snap, err := ExportSnapshot(app, siteAbs)
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

	if err := PublishToProduction(toURL, token, payloads); err != nil {
		return err
	}

	checksum, err := ChecksumExport(snap)
	if err != nil {
		return err
	}

	collections := make([]string, len(payloads))
	for i, p := range payloads {
		collections[i] = p.Collection
	}

	return WritePublishState(siteAbs, PublishState{
		Checksum:    checksum,
		PublishedAt: time.Now().UTC().Format(time.RFC3339),
		Collections: collections,
	})
}
