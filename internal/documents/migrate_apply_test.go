package documents

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestPlanFromSourceAndApplyPlansDryRun(t *testing.T) {
	legacy := t.TempDir()
	testutil.WriteFile(t, filepath.Join(legacy, "post-one.md"), `---
title: Post One
---
Body one
`)
	testutil.WriteFile(t, filepath.Join(legacy, "post-two.md"), `---
title: Post Two
---
Body two
`)

	plans, summary, err := PlanFromSource(legacy)
	if err != nil {
		t.Fatalf("PlanFromSource: %v", err)
	}
	if summary.FileCount != 2 {
		t.Fatalf("file count = %d, want 2", summary.FileCount)
	}
	if len(plans) != 1 {
		t.Fatalf("plans = %d, want 1", len(plans))
	}

	site := testutil.NewSite(t)
	result, err := ApplyPlans(site, plans, ApplyOptions{DryRun: true})
	if err != nil {
		t.Fatalf("ApplyPlans dry-run: %v", err)
	}
	if result.TotalBound != 2 {
		t.Fatalf("total bound = %d, want 2", result.TotalBound)
	}
	if _, err := os.Stat(filepath.Join(site, "documents", plans[0].Collection, "post-one.md")); !os.IsNotExist(err) {
		t.Fatal("dry-run should not write markdown files")
	}
}

func TestResolveSourceRootRejectsMissing(t *testing.T) {
	_, err := ResolveSourceRoot(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestParsePlansYAMLRejectsEmpty(t *testing.T) {
	_, err := ParsePlansYAML("")
	if err == nil {
		t.Fatal("expected error for empty plan")
	}
}
