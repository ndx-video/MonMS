package testutil

import (
	"testing"

	"github.com/monms/monms/internal/site"
)

func TestNewEditorialSite(t *testing.T) {
	ws := NewEditorialSite(t)
	if err := site.ValidateSite(ws); err != nil {
		t.Fatalf("ValidateSite() = %v, want nil", err)
	}
}
