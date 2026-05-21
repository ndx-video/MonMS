package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestValidateWorkspaceValidFixture(t *testing.T) {
	ws := testutil.NewWorkspace(t)
	if err := ValidateWorkspace(ws); err != nil {
		t.Fatalf("ValidateWorkspace() = %v, want nil", err)
	}
}

func TestValidateWorkspaceMissingBaseLayout(t *testing.T) {
	ws := testutil.NewWorkspace(t)
	base := filepath.Join(ws, "templates/layouts/base.gohtml")
	if err := os.Remove(base); err != nil {
		t.Fatalf("remove base layout: %v", err)
	}

	err := ValidateWorkspace(ws)
	if err == nil {
		t.Fatal("ValidateWorkspace() = nil, want error")
	}
	if !strings.Contains(err.Error(), "monms init") {
		t.Errorf("error = %q, want substring %q", err.Error(), "monms init")
	}
}
