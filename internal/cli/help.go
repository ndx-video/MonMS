package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// MonmsCommands are subcommands handled before PocketBase (serve, superuser, …).
var MonmsCommands = map[string]bool{
	"init":      true,
	"validate":  true,
	"content":   true,
	"documents": true,
	"site":      true,
	"stop":      true,
	"restart":   true,
}

// ParseHelpRequest reports whether args ask for help and which command's help to show.
// Empty command means root help. For PocketBase commands (serve), wantHelp may be true
// with command "serve" — caller should fall through to app.Start().
func ParseHelpRequest(args []string) (command string, wantHelp bool) {
	if len(args) == 0 {
		return "", false
	}
	if args[0] == "help" {
		if len(args) == 1 {
			return "", true
		}
		return args[1], true
	}
	for i, a := range args {
		if isHelpFlag(a) {
			if i == 0 {
				return "", true
			}
			return args[0], true
		}
	}
	return "", false
}

// IsMonmsCommand reports whether name is handled by MonMS before PocketBase.
func IsMonmsCommand(name string) bool {
	return MonmsCommands[name]
}

// PrintHelp writes help for command (empty = root) to stdout.
func PrintHelp(command string) {
	switch command {
	case "":
		printRootHelp(os.Stdout)
	case "init":
		printInitHelp(os.Stdout)
	case "validate":
		printValidateHelp(os.Stdout)
	case "content":
		printContentHelp(os.Stdout)
	case "documents":
		printDocumentsHelp(os.Stdout)
	case "site":
		printSiteHelp(os.Stdout)
	case "stop":
		printStopHelp(os.Stdout)
	case "restart":
		printRestartHelp(os.Stdout)
	default:
		printRootHelp(os.Stdout)
	}
}

func isHelpFlag(arg string) bool {
	switch arg {
	case "-h", "--help", "-help":
		return true
	default:
		return false
	}
}

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, `MonMS — agent-malleable monolithic CMS (Go + PocketBase)

Usage:
  monms [command] [flags]
  monms [flags]                 Start the web server (same as monms serve)

MonMS commands:
  init       Scaffold a new site (templates, schema, assets, git hook)
  validate   Lint site templates (*.gohtml)
  content    Export, import, diff, or publish editorial content (JSON rail)
  documents  Sync, migrate, or diff Git markdown documents
  site       Sync site Git shape (fetch + checkout ref)
  stop       Stop running monms serve processes for this binary
  restart    Stop running instances, then start serve again

Server commands (PocketBase):
  serve      Start the HTTP/HTTPS server (default when no command is given)
  superuser  Manage admin accounts

Configuration:
  -s, --site PATH    Site directory (default: ./site or MONMS_SITE)
  site/.monms/config.json — siteUrl, productionUrl, publisherEmails, allowedHosts, bind, shapeSync

Examples:
  monms init -s ./my-site
  monms validate --site ./site templates/index.gohtml
  monms content export --site ./site
  monms site sync --site ./site --ref v1.2.0
  monms serve --http=127.0.0.1:8090
  monms stop
  monms restart -s ./site

Run "monms <command> --help" for command-specific help.`)
}

func printInitHelp(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  monms init [-s|--site PATH]

Scaffold or refresh a MonMS site at PATH (default: ./site or MONMS_SITE).
On an interactive terminal, also runs the setup wizard (port, allowed hosts,
bind address) and offers to start the server. Idempotent — safe to re-run.

Creates:
  templates/layouts/base.gohtml   Site shell + HTMX inline editing
  templates/index.gohtml          Homepage template
  templates/errors/errors.gohtml  Error pages
  assets/main.css                 Base styles
  schema/hero_content.json        Demo editorial collection
  content/                        Editorial export directory
  .monms/config.json              Staging config (documented; gitignored on commit)
  .monms/config.example.json      Committed config template
  .git/hooks/pre-commit           Validates staged *.gohtml via monms validate

Existing files are never overwritten. Initializes git if .git/ is missing and git
is on PATH. Installs or refreshes the pre-commit hook when safe.

Examples:
  monms init
  monms init --site /var/www/staging`)
}

func printValidateHelp(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  monms validate [-s|--site PATH] [files...]

Parse and lint Go HTML templates using the same rules as production SSR.

With no file arguments, validates staged *.gohtml files from git (pre-commit
mode). If git is unavailable, pass explicit paths.

Examples:
  monms validate
  monms validate templates/index.gohtml
  monms validate --site ./site templates/press/index.gohtml`)
}

func printContentHelp(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  monms content <export|import|diff|publish> [-s|--site PATH] [flags]

PB-native editorial content rail — sync PocketBase records marked editorial in schema JSON
without monms.source=markdown. Markdown collections promote via Git (see monms documents).

Subcommands:
  export    Write site/content/{collection}.json from live records
  import    Upsert site/content/*.json into local .pb_data/
  diff      Show pending changes vs last publish; exit 1 if any (for CI)
  publish   Export staging and POST to production (requires --to and MONMS_PUBLISH_TOKEN)

Publish flags:
  --to URL  Production MonMS base URL (required for publish)

Examples:
  monms content export --site ./site
  monms content diff --site ./site
  monms content publish --site ./site --to https://production.example.com

Clients normally publish via the web console at /_monms/publish.`)
}

func printDocumentsHelp(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  monms documents <sync|diff|scan|plan|bind> [-s|--site PATH] [flags]

Markdown document rail — Git-tracked .md files under documents/ sync into PocketBase
for SSR. Collections declare monms.source=markdown in schema JSON.

Subcommands:
  sync      Upsert markdown files into .pb_data/ (also runs on serve bootstrap)
  diff      List PB records with no backing markdown file; exit 1 if any
  scan      Inventory a legacy markdown tree
  plan      Propose collection bindings (--source DIR [--out plan.yaml])
  bind      Apply plan: write schema, copy files, inject frontmatter (--config plan.yaml)

Bind flags:
  --config FILE  Plan YAML from documents plan (required)
  --apply        Write schema and markdown files
  --dry-run      Show planned actions without writing
  --force        Overwrite existing frontmatter id values

Examples:
  monms documents sync --site ./site
  monms documents scan --source ./legacy-docs
  monms documents plan --source ./legacy-docs --out plan.yaml
  monms documents bind --config plan.yaml --apply --site ./site`)
}

func printSiteHelp(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  monms site sync [-s|--site PATH] --ref REF [--remote origin] [--force]

Fetch tags from a Git remote and checkout a shape ref (tag or branch) in the
site repository. Uses fetch + checkout — not git pull.

Flags:
  --ref REF       Git ref to checkout (required)
  --remote NAME   Remote to fetch from (default: origin)
  --force         Checkout even when the worktree has local changes

Examples:
  monms site sync --site ./site --ref v1.2.0
  monms site sync -s /var/www/staging --ref main --remote origin

Optional startup sync: set shapeSync.enabled in site/.monms/config.json.`)
}

func printRestartHelp(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  monms restart [-s|--site PATH] [serve flags]

Stop all running monms processes that use this binary, then start serve with
the same flags (including -d/--daemon for background mode).

Examples:
  monms restart
  monms restart -s ./site --http=127.0.0.1:8090
  monms restart -d`)
}

func printStopHelp(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  monms stop

Send SIGTERM to all running monms processes that use this binary, then SIGKILL
any that remain after a short grace period.

Examples:
  monms stop`)
}

// HasHelpFlag reports whether args contain -h or --help anywhere.
func HasHelpFlag(args []string) bool {
	_, want := ParseHelpRequest(args)
	return want
}

// StripHelpFlags removes help flags from args (for sub-handlers that parse flags).
func StripHelpFlags(args []string) []string {
	var out []string
	for _, a := range args {
		if !isHelpFlag(a) && a != "help" {
			out = append(out, a)
		}
	}
	return out
}

// ContentSubcommandHelp returns help text for a content subcommand, or false.
func ContentSubcommandHelp(sub string) (string, bool) {
	switch strings.ToLower(sub) {
	case "export":
		return "Usage: monms content export [-s|--site PATH]\n\nExport editorial collections to site/content/*.json.\n", true
	case "import":
		return "Usage: monms content import [-s|--site PATH]\n\nImport site/content/*.json into local .pb_data/ (idempotent upsert).\n", true
	case "diff":
		return "Usage: monms content diff [-s|--site PATH]\n\nPrint field-level changes since last publish. Exit 1 if pending.\n", true
	case "publish":
		return "Usage: monms content publish [-s|--site PATH] --to URL\n\nPOST editorial export to production import API. Requires MONMS_PUBLISH_TOKEN.\n", true
	default:
		return "", false
	}
}

// SiteSubcommandHelp returns help text for a site subcommand, or false.
func SiteSubcommandHelp(sub string) (string, bool) {
	switch strings.ToLower(sub) {
	case "sync":
		return "Usage: monms site sync [-s|--site PATH] --ref REF [--remote origin] [--force]\n\nFetch tags and checkout a site shape ref.\n", true
	default:
		return "", false
	}
}

// DocumentsSubcommandHelp returns help text for a documents subcommand, or false.
func DocumentsSubcommandHelp(sub string) (string, bool) {
	switch strings.ToLower(sub) {
	case "sync":
		return "Usage: monms documents sync [-s|--site PATH]\n\nUpsert markdown files from documents/ into PocketBase.\n", true
	case "diff":
		return "Usage: monms documents diff [-s|--site PATH]\n\nList orphan PB records without backing markdown files.\n", true
	case "scan":
		return "Usage: monms documents scan --source DIR\n\nInventory markdown files in a legacy tree.\n", true
	case "plan":
		return "Usage: monms documents plan --source DIR [--out plan.yaml]\n\nPropose collection bindings for migration.\n", true
	case "bind":
		return "Usage: monms documents bind --config plan.yaml [--apply|--dry-run] [-s|--site PATH]\n\nApply migration plan: schema + frontmatter + file copy.\n", true
	default:
		return "", false
	}
}
