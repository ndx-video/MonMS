package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestValidateSiteValidFixture(t *testing.T) {
	ws := testutil.NewSite(t)
	if err := ValidateSite(ws); err != nil {
		t.Fatalf("ValidateSite() = %v, want nil", err)
	}
}

func TestValidateSiteMissingBaseLayout(t *testing.T) {
	ws := testutil.NewSite(t)
	base := filepath.Join(ws, "templates/layouts/base.gohtml")
	if err := os.Remove(base); err != nil {
		t.Fatalf("remove base layout: %v", err)
	}

	err := ValidateSite(ws)
	if err == nil {
		t.Fatal("ValidateSite() = nil, want error")
	}
	if !strings.Contains(err.Error(), "monms init") {
		t.Errorf("error = %q, want substring %q", err.Error(), "monms init")
	}
}
