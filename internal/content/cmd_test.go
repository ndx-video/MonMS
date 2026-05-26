package content

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/monmsroutes"
	"github.com/monms/monms/internal/testutil"
)

func TestContentExportCLI(t *testing.T) {
	ws := testutil.NewEditorialWorkspace(t)

	if err := RunCLI([]string{"export", "--workspace", ws}); err != nil {
		t.Fatalf("RunCLI export: %v", err)
	}

	path := filepath.Join(ws, "content", "hero_content.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("export file missing: %v", err)
	}
}

func TestContentExportCLIWithShortWorkspaceFlag(t *testing.T) {
	ws := testutil.NewEditorialWorkspace(t)

	if err := RunCLI([]string{"export", "-w", ws}); err != nil {
		t.Fatalf("RunCLI export -w: %v", err)
	}

	path := filepath.Join(ws, "content", "hero_content.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("export file missing: %v", err)
	}
}

func TestContentImportCLIIdempotent(t *testing.T) {
	ws := testutil.NewEditorialWorkspace(t)

	if err := RunCLI([]string{"export", "--workspace", ws}); err != nil {
		t.Fatalf("export: %v", err)
	}
	if err := RunCLI([]string{"import", "--workspace", ws}); err != nil {
		t.Fatalf("first import: %v", err)
	}
	if err := RunCLI([]string{"import", "--workspace", ws}); err != nil {
		t.Fatalf("second import: %v", err)
	}
}

func TestContentDiffCLI(t *testing.T) {
	ws := testutil.NewEditorialWorkspace(t)

	if err := RunCLI([]string{"export", "--workspace", ws}); err != nil {
		t.Fatalf("export: %v", err)
	}

	app, err := bootstrapApp(ws)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	rec, err := app.FindRecordById("hero_content", "homepage")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	rec.Set("title", "CLI Diff Title")
	if err := app.Save(rec); err != nil {
		t.Fatalf("save: %v", err)
	}

	err = RunCLI([]string{"diff", "--workspace", ws})
	if !errors.Is(err, ErrPendingChanges) {
		t.Fatalf("diff err = %v, want ErrPendingChanges", err)
	}
}

func TestContentPublishCLI(t *testing.T) {
	const wantToken = "test-publish-token"
	var gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != monmsroutes.ContentImportPath {
			t.Errorf("path %q, want %s", r.URL.Path, monmsroutes.ContentImportPath)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method %q, want POST", r.Method)
		}
		gotAuth = r.Header.Get("Authorization")
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Errorf("Content-Type %q, want application/json", ct)
		}
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			t.Error("empty request body")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ws := testutil.NewEditorialWorkspace(t)
	t.Setenv("MONMS_PUBLISH_TOKEN", wantToken)

	if err := RunCLI([]string{"publish", "--workspace", ws, "--to", srv.URL}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	if gotAuth != "Bearer "+wantToken {
		t.Fatalf("Authorization %q, want Bearer %s", gotAuth, wantToken)
	}

	state, err := ReadPublishState(ws)
	if err != nil {
		t.Fatalf("read publish state: %v", err)
	}
	if state.Checksum == "" {
		t.Fatal("publish state checksum empty after CLI publish")
	}
	if state.PublishedAt == "" {
		t.Fatal("publish state PublishedAt empty after CLI publish")
	}
}

func TestContentPublishCLIMissingToken(t *testing.T) {
	ws := testutil.NewEditorialWorkspace(t)
	t.Setenv("MONMS_PUBLISH_TOKEN", "")

	err := RunCLI([]string{"publish", "--workspace", ws, "--to", "http://127.0.0.1:1"})
	if err == nil {
		t.Fatal("publish without token: want error")
	}
	if !strings.Contains(err.Error(), "publish token") {
		t.Fatalf("error %q, want publish token message", err)
	}
}
