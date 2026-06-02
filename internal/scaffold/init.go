package scaffold

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/cli/prompt"
	"github.com/monms/monms/internal/config"
)

// preCommitHookScript is installed into site/.git/hooks/pre-commit (D-40).
// The monms-validate-hook comment is the idempotency marker — do not remove it.
const preCommitHookScript = `#!/bin/sh
# monms-validate-hook — DO NOT REMOVE THIS COMMENT (idempotency marker)

HOOK_DIR="$(cd "$(dirname "$0")" && pwd)"
SITE_ROOT="$(cd "$HOOK_DIR/../.." && pwd)"

if [ -n "$MONMS_BIN" ]; then
  MONMS="$MONMS_BIN"
elif command -v monms >/dev/null 2>&1; then
  MONMS="monms"
else
  CANDIDATE="$SITE_ROOT/monms"
  if [ -x "$CANDIDATE" ]; then
    MONMS="$CANDIDATE"
  else
    echo "monms: binary not found" >&2
    echo "  Set MONMS_BIN, add monms to PATH, or place binary at $SITE_ROOT/monms" >&2
    exit 1
  fi
fi

STAGED=$(git diff --cached --name-only --diff-filter=ACM | grep '\.gohtml$')
if [ -z "$STAGED" ]; then
  exit 0
fi

if ! echo "$STAGED" | tr '\n' '\0' | xargs -0 "$MONMS" validate -s "$SITE_ROOT"; then
  echo "" >&2
  echo "Pre-commit validation failed. Rolling back site to last stable state..." >&2
  git checkout -- .
  echo "Site restored. Fix the errors above, then re-apply your changes." >&2
  exit 1
fi

exit 0
`

type scaffoldFile struct {
	embedPath string
	destPath  string
}

var scaffoldFiles = []scaffoldFile{
	{"embed/base.gohtml", "templates/layouts/base.gohtml"},
	{"embed/index.gohtml", "templates/index.gohtml"},
	{"embed/doc.gohtml", "templates/doc.gohtml"},
	{"embed/errors.gohtml", "templates/errors/errors.gohtml"},
	{"embed/main.css", "assets/main.css"},
	{"embed/hero_content.json", "schema/hero_content.json"},
	{"embed/monms-config.json", ".monms/config.json"},
	{"embed/monms-config.example.json", ".monms/config.example.json"},
	{"embed/Dockerfile.example", "Dockerfile.example"},
	{"embed/docker-compose.example.yml", "docker-compose.example.yml"},
}

var scaffoldDirs = []string{
	"templates/layouts",
	"templates/fragments",
	"templates/errors",
	"assets",
	"schema",
	".monms",
	"content",
	"documents",
}

// Result lists artifacts created during InitAt.
type Result struct {
	Created []string
}

// CLIOutcome is returned from interactive init.
type CLIOutcome struct {
	SiteAbs   string
	StartMode StartMode
}

// RunInit scaffolds a site at the resolved path (D-05, D-07).
// On an interactive terminal it also runs the setup wizard.
func RunInit(args []string) error {
	_, err := RunInitCLI(args, &prompt.Stdio)
	return err
}

// RunInitCLI scaffolds a site and optionally runs the setup wizard on a TTY.
func RunInitCLI(args []string, p *prompt.Prompter) (CLIOutcome, error) {
	res, err := config.ResolveSiteMeta(args, os.Environ())
	if err != nil {
		return CLIOutcome{}, err
	}

	result, err := InitAt(res.Absolute)
	if err != nil {
		return CLIOutcome{}, err
	}
	printCreatedPaths(result)

	out := CLIOutcome{SiteAbs: res.Absolute, StartMode: StartNone}
	if !p.IsInteractive() {
		slog.Info("site initialized", "path", res.Absolute)
		printInitSummary(res.Absolute)
		return out, nil
	}

	mode, err := RunSetupWizard(res.Absolute, p)
	if err != nil {
		return CLIOutcome{}, err
	}
	out.StartMode = mode
	slog.Info("site initialized", "path", res.Absolute)
	return out, nil
}

// InitAt scaffolds a new site at siteAbs, skipping existing files.
func InitAt(siteAbs string) (*Result, error) {
	result := &Result{}

	if err := os.MkdirAll(siteAbs, 0o755); err != nil {
		return nil, fmt.Errorf("create site root: %w", err)
	}

	for _, dir := range scaffoldDirs {
		created, err := mkdirUnder(siteAbs, dir)
		if err != nil {
			return nil, err
		}
		if created {
			result.Created = append(result.Created, filepath.Join(siteAbs, dir))
		}
	}

	for _, sf := range scaffoldFiles {
		created, err := writeScaffoldFile(siteAbs, sf.embedPath, sf.destPath)
		if err != nil {
			return nil, err
		}
		if created {
			result.Created = append(result.Created, filepath.Join(siteAbs, sf.destPath))
		}
	}

	for _, rel := range []string{"schema/.gitkeep", "templates/fragments/.gitkeep", "content/.gitkeep"} {
		created, err := writeKeepFile(siteAbs, rel)
		if err != nil {
			return nil, err
		}
		if created {
			result.Created = append(result.Created, filepath.Join(siteAbs, rel))
		}
	}

	if created, err := maybeGitInit(siteAbs); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, filepath.Join(siteAbs, ".git"))
	}

	if created, err := installPreCommitHook(siteAbs); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, filepath.Join(siteAbs, ".git", "hooks", "pre-commit"))
	}

	return result, nil
}

func printCreatedPaths(result *Result) {
	if result == nil || len(result.Created) == 0 {
		return
	}
	fmt.Fprintln(os.Stdout, "Created:")
	for _, p := range result.Created {
		fmt.Println(p)
	}
}

func printInitSummary(siteAbs string) {
	fmt.Fprintf(os.Stdout, `
MonMS site ready at %s

Scaffolded:
  templates/          Page shells and layouts (HTMX inline editing)
  assets/             Static CSS and media paths
  schema/             Collection bootstrap JSON
  content/            Editorial export snapshots (monms content export)
  .monms/config.json  Staging publish config — edit publisherEmails and productionUrl

Documentation: docs/README.md in the MonMS engine repository (user guide, operators manual, API reference).

Next steps:
  1. Edit .monms/config.json (_fieldDocs describes each option)
  2. monms serve -s %s
  3. Open http://127.0.0.1:8090/_/ and create a PocketBase admin
  4. Add admin email(s) to publisherEmails in .monms/config.json for Publish to live

Tip: commit config.example.json; keep config.json gitignored with site-specific URLs.
`, siteAbs, siteAbs)
}

func mkdirUnder(siteRoot, rel string) (created bool, err error) {
	dest := filepath.Join(siteRoot, rel)
	if err := ensureUnderSite(siteRoot, dest); err != nil {
		return false, err
	}
	if _, err := os.Stat(dest); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat %s: %w", rel, err)
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return false, fmt.Errorf("mkdir %s: %w", rel, err)
	}
	return true, nil
}

func writeScaffoldFile(siteRoot, embedPath, destRel string) (created bool, err error) {
	dest := filepath.Join(siteRoot, destRel)
	if err := ensureUnderSite(siteRoot, dest); err != nil {
		return false, err
	}

	if _, err := os.Stat(dest); err == nil {
		slog.Info("skip existing scaffold file", "path", destRel)
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat %s: %w", destRel, err)
	}

	data, err := fs.ReadFile(scaffoldFS, embedPath)
	if err != nil {
		return false, fmt.Errorf("read embed %s: %w", embedPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return false, fmt.Errorf("mkdir parent for %s: %w", destRel, err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return false, fmt.Errorf("write %s: %w", destRel, err)
	}
	return true, nil
}

func writeKeepFile(siteRoot, destRel string) (created bool, err error) {
	dest := filepath.Join(siteRoot, destRel)
	if err := ensureUnderSite(siteRoot, dest); err != nil {
		return false, err
	}
	if _, err := os.Stat(dest); err == nil {
		slog.Info("skip existing scaffold file", "path", destRel)
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat %s: %w", destRel, err)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return false, fmt.Errorf("mkdir parent for %s: %w", destRel, err)
	}
	if err := os.WriteFile(dest, nil, 0o644); err != nil {
		return false, fmt.Errorf("write %s: %w", destRel, err)
	}
	return true, nil
}

// ensureUnderSite prevents writes outside the resolved site (T-01-12).
func ensureUnderSite(siteRoot, dest string) error {
	siteRoot = filepath.Clean(siteRoot)
	dest = filepath.Clean(dest)
	rel, err := filepath.Rel(siteRoot, dest)
	if err != nil {
		return fmt.Errorf("resolve path under site: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("refusing write outside site: %s", dest)
	}
	return nil
}

func installPreCommitHook(siteRoot string) (created bool, err error) {
	gitDir := filepath.Join(siteRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		slog.Warn("no .git directory found, skipping pre-commit hook install", "dir", gitDir)
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("stat .git: %w", err)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(hooksDir, 0o755); err != nil {
			return false, fmt.Errorf("mkdir hooks dir: %w", err)
		}
	} else if err != nil {
		return false, fmt.Errorf("stat hooks dir: %w", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	if existing, err := os.ReadFile(hookPath); err == nil {
		if bytes.Contains(existing, []byte("monms-validate-hook")) {
			slog.Info("pre-commit hook already installed, skipping")
			return false, nil
		}
	}

	if err := os.WriteFile(hookPath, []byte(preCommitHookScript), 0o755); err != nil {
		return false, fmt.Errorf("install pre-commit hook: %w", err)
	}
	slog.Info("pre-commit hook installed", "path", hookPath)
	return true, nil
}

func maybeGitInit(siteRoot string) (created bool, err error) {
	gitDir := filepath.Join(siteRoot, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		slog.Info("git repository already exists, skipping git init")
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat .git: %w", err)
	}

	if _, err := exec.LookPath("git"); err != nil {
		slog.Warn("git not found in PATH; skipping git init")
		return false, nil
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = siteRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("git init: %w: %s", err, string(out))
	}
	return true, nil
}
