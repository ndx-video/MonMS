package testutil

import (
	"testing"

	"github.com/monms/monms/internal/workspace"
)

func TestNewEditorialWorkspace(t *testing.T) {
	ws := NewEditorialWorkspace(t)
	if err := workspace.ValidateWorkspace(ws); err != nil {
		t.Fatalf("ValidateWorkspace() = %v, want nil", err)
	}
}
