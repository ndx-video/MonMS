package documents

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/monms/monms/internal/cli"
	"github.com/monms/monms/internal/config"
	"github.com/monms/monms/internal/logging"
	"github.com/monms/monms/internal/authbootstrap"
	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"gopkg.in/yaml.v2"
)

// RunCLI is the entry point for `monms documents`.
func RunCLI(args []string) error {
	if len(args) == 0 {
		cli.PrintHelp("documents")
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
	case "sync":
		return runSync(siteAbs)
	case "diff":
		return runDiff(siteAbs)
	case "scan":
		return runScan(args[1:])
	case "plan":
		return runPlan(args[1:])
	case "bind":
		return runBind(args[1:], siteAbs)
	default:
		return fmt.Errorf("unknown documents subcommand %q (want sync, diff, scan, plan, or bind)", sub)
	}
}

func bootstrapApp(siteAbs string) (core.App, error) {
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  filepath.Join(siteAbs, ".pb_data"),
		DefaultDev:      true,
		HideStartBanner: true,
	})
	RegisterBootstrapHook(app, siteAbs)
	authbootstrap.RegisterBootstrapHook(app)
	schema.RegisterBootstrapHook(app, siteAbs)
	if err := app.Bootstrap(); err != nil {
		return nil, fmt.Errorf("documents bootstrap: %w", err)
	}
	return app, nil
}

func runSync(siteAbs string) error {
	app, err := bootstrapApp(siteAbs)
	if err != nil {
		return err
	}
	result, err := SyncAll(app, siteAbs)
	if err != nil {
		return err
	}
	if len(result.Errors) > 0 {
		for _, msg := range result.Errors {
			fmt.Fprintln(os.Stderr, "warning:", msg)
		}
	}
	fmt.Printf("documents sync: %d collection(s), %d record(s) upserted\n",
		result.Collections, result.Upserted)
	slog.Info("documents sync: complete", "site", siteAbs, "upserted", result.Upserted)
	return nil
}

func runDiff(siteAbs string) error {
	app, err := bootstrapApp(siteAbs)
	if err != nil {
		return err
	}
	diff, err := DiffOrphans(app, siteAbs)
	if err != nil {
		return err
	}
	if len(diff.Orphans) == 0 {
		fmt.Println("no orphan records (all PB records have backing markdown files)")
		return nil
	}
	fmt.Println("orphan records (PB without markdown file):")
	for _, o := range diff.Orphans {
		fmt.Printf("  %s/%s", o.Collection, o.ID)
		if o.Slug != "" {
			fmt.Printf(" (slug=%s)", o.Slug)
		}
		fmt.Println()
	}
	return errors.New("orphan markdown records found")
}

func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var source string
	fs.StringVar(&source, "source", "", "legacy markdown tree root (required)")
	if err := fs.Parse(config.StripSiteFlags(args)); err != nil {
		return err
	}
	if source == "" && fs.NArg() > 0 {
		source = fs.Arg(0)
	}
	if source == "" {
		return fmt.Errorf("documents scan: source path required")
	}

	entries, err := ScanTree(source)
	if err != nil {
		return err
	}
	fmt.Printf("found %d markdown file(s) under %s\n", len(entries), source)
	for _, e := range entries {
		fm := "no"
		if e.HasFrontmatter {
			fm = "yes"
		}
		fmt.Printf("  %s  title=%q  frontmatter=%s\n", e.RelPath, e.Title, fm)
	}
	return nil
}

func runPlan(args []string) error {
	fs := flag.NewFlagSet("plan", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var source, out string
	fs.StringVar(&source, "source", "", "legacy markdown tree root (required)")
	fs.StringVar(&out, "out", "", "write plan YAML to file (optional)")
	if err := fs.Parse(config.StripSiteFlags(args)); err != nil {
		return err
	}
	if source == "" && fs.NArg() > 0 {
		source = fs.Arg(0)
	}
	if source == "" {
		return fmt.Errorf("documents plan: --source required")
	}

	entries, err := ScanTree(source)
	if err != nil {
		return err
	}
	plans := DefaultPlanFromScan(source, entries)
	if len(plans) == 0 {
		fmt.Println("no markdown files found")
		return nil
	}

	data, err := yaml.Marshal(plans)
	if err != nil {
		return err
	}
	if out != "" {
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote plan to %s (%d binding(s))\n", out, len(plans))
		return nil
	}
	fmt.Print(string(data))
	return nil
}

func runBind(args []string, siteAbs string) error {
	fs := flag.NewFlagSet("bind", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		if text, ok := cli.DocumentsSubcommandHelp("bind"); ok {
			fmt.Print(text)
		}
	}
	var configPath string
	var apply, force, dryRun bool
	fs.StringVar(&configPath, "config", "", "plan YAML from documents plan (required)")
	fs.BoolVar(&apply, "apply", false, "write schema and markdown files")
	fs.BoolVar(&force, "force", false, "overwrite existing frontmatter id")
	fs.BoolVar(&dryRun, "dry-run", false, "show actions without writing")
	if err := fs.Parse(config.StripSiteFlags(args)); err != nil {
		if err == flag.ErrHelp {
			fs.Usage()
			return nil
		}
		return err
	}
	if configPath == "" {
		return fmt.Errorf("documents bind: --config plan YAML required")
	}
	if !apply && !dryRun {
		return fmt.Errorf("documents bind: specify --apply or --dry-run")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var plans []BindPlan
	if err := yaml.Unmarshal(data, &plans); err != nil {
		return err
	}

	for _, plan := range plans {
		if apply && !dryRun {
			schemaPath := filepath.Join(siteAbs, "schema", plan.Collection+".json")
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				schemaJSON := DefaultArticlesSchema(plan.Collection, plan.DestRoot, plan.FieldMap)
				if err := os.WriteFile(schemaPath, []byte(schemaJSON), 0o644); err != nil {
					return err
				}
				fmt.Printf("wrote schema %s\n", schemaPath)
			}
		}

		n, err := ApplyBind(siteAbs, plan, dryRun, force)
		if err != nil {
			return err
		}
		action := "would bind"
		if apply && !dryRun {
			action = "bound"
		}
		fmt.Printf("%s %d file(s) for collection %q -> %s\n", action, n, plan.Collection, plan.DestRoot)
	}

	if apply && !dryRun {
		app, err := bootstrapApp(siteAbs)
		if err != nil {
			return err
		}
		result, err := SyncAll(app, siteAbs)
		if err != nil {
			return err
		}
		fmt.Printf("documents sync: %d record(s) upserted\n", result.Upserted)
	}
	return nil
}
