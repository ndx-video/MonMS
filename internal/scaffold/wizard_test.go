package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/cli/prompt"
)

func TestParseAllowedHosts(t *testing.T) {
	t.Parallel()
	got := parseAllowedHosts("")
	if len(got) != 2 || got[0] != "localhost" || got[1] != "127.0.0.1" {
		t.Fatalf("default = %v", got)
	}
	got = parseAllowedHosts("staging.example.com app.example.com")
	if len(got) != 2 || got[0] != "staging.example.com" {
		t.Fatalf("custom = %v", got)
	}
}

func TestRunSetupWizardDefaults(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".monms"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".monms", "config.json"), []byte(`{
  "_comment": "keep",
  "_fieldDocs": {"bind": "doc"},
  "productionUrl": "https://example.com"
}`), 0o644); err != nil {
		t.Fatal(err)
	}

	in := strings.NewReader("\n\n\n3\n")
	p := &prompt.Prompter{In: in, Out: ioDiscard{}}
	mode, err := RunSetupWizard(dir, p)
	if err != nil {
		t.Fatalf("RunSetupWizard: %v", err)
	}
	if mode != StartNone {
		t.Fatalf("mode = %v, want StartNone", mode)
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".monms", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	var bind ServeBindConfig
	if err := json.Unmarshal(doc["bind"], &bind); err != nil || bind.Host != "0.0.0.0" || bind.Port != "8090" {
		t.Fatalf("bind = %#v, err = %v", bind, err)
	}
	var allowed []string
	if err := json.Unmarshal(doc["allowedHosts"], &allowed); err != nil || len(allowed) != 2 {
		t.Fatalf("allowedHosts = %v, err = %v", allowed, err)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func TestInitAtCreatedPaths(t *testing.T) {
	dir := t.TempDir()
	result, err := InitAt(dir)
	if err != nil {
		t.Fatalf("InitAt: %v", err)
	}
	if len(result.Created) == 0 {
		t.Fatal("expected created paths on fresh init")
	}
	for _, p := range result.Created {
		if !filepath.IsAbs(p) {
			t.Fatalf("expected absolute path, got %q", p)
		}
	}

	result2, err := InitAt(dir)
	if err != nil {
		t.Fatalf("InitAt re-run: %v", err)
	}
	if len(result2.Created) != 0 {
		t.Fatalf("re-init created %v, want none", result2.Created)
	}
}
