package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/cli/prompt"
	"github.com/monms/monms/internal/config"
)

func TestCheckSiteMissingRoot(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	st, missing := CheckSite(dir)
	if st != StatusMissingRoot || len(missing) == 0 {
		t.Fatalf("CheckSite() = %v, %v", st, missing)
	}
}

func TestCheckSiteIncomplete(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	st, _ := CheckSite(dir)
	if st != StatusIncomplete {
		t.Fatalf("status = %v, want Incomplete", st)
	}
}

func TestEnsureReadyDeclined(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	res := config.SiteResolution{Configured: dir, Absolute: dir, SiteFlagSet: true}
	p := &prompt.Prompter{In: strings.NewReader("n\n"), Out: ioDiscard{}, ForceInteractive: true}
	_, err := EnsureReady(res, p)
	if err != ErrDeclined {
		t.Fatalf("EnsureReady() = %v, want ErrDeclined", err)
	}
}

func TestEnsureReadyNonInteractive(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	res := config.SiteResolution{Configured: "./site", Absolute: dir}
	p := &prompt.Prompter{In: strings.NewReader("y\n"), Out: ioDiscard{}}
	_, err := EnsureReady(res, p)
	if err == nil || !strings.Contains(err.Error(), "site not found") {
		t.Fatalf("EnsureReady() = %v, want non-interactive error", err)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
