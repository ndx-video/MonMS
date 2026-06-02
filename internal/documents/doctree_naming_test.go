package documents

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestDoctreeCollectionName(t *testing.T) {
	tests := []struct {
		stub, leaf, want string
	}{
		{"guide", ".", "dt_guide"},
		{"guide", "tutorials", "dt_guide_tutorials"},
		{"guide", "tutorials/advanced", "dt_guide_tutorials_advanced"},
		{"guide", "parent-only/child", "dt_guide_parent_only_child"},
	}
	for _, tc := range tests {
		got := DoctreeCollectionName(tc.stub, tc.leaf)
		if got != tc.want {
			t.Errorf("DoctreeCollectionName(%q, %q) = %q, want %q", tc.stub, tc.leaf, got, tc.want)
		}
	}
}

func TestDiscoverPathDrivenNamesNoCollision(t *testing.T) {
	site := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(site, "guide", "parent", "child", "a.md"), "# A\n")
	testutil.WriteFile(t, filepath.Join(site, "guide", "other", "child", "b.md"), "# B\n")

	candidates, err := DiscoverLeafCollections(site, "guide")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(candidates) != 2 {
		t.Fatalf("candidates = %d, want 2", len(candidates))
	}
	names := map[string]bool{}
	for _, c := range candidates {
		if names[c.Collection] {
			t.Fatalf("duplicate collection name %q", c.Collection)
		}
		names[c.Collection] = true
		if !strings.HasPrefix(c.Collection, "dt_guide_") {
			t.Fatalf("collection = %q, want dt_guide_* prefix", c.Collection)
		}
	}
	if !names["dt_guide_parent_child"] || !names["dt_guide_other_child"] {
		t.Fatalf("expected dt_guide_parent_child and dt_guide_other_child, got %v", names)
	}
}
