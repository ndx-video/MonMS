package documents

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase/core"
	"gopkg.in/yaml.v2"
)

const scanPreviewMaxPaths = 20

// ScanSummary describes a legacy tree inventory for UI preview.
type ScanSummary struct {
	SourceRoot string
	FileCount  int
	SamplePaths []string
}

// ApplyOptions controls migration bind execution.
type ApplyOptions struct {
	DryRun      bool
	Force       bool
	WriteSchema bool
}

// PlanCollectionResult summarizes one collection in a bind run.
type PlanCollectionResult struct {
	Collection string
	DestRoot   string
	Bound      int
	SchemaWritten bool
}

// ApplyResult summarizes a migration apply run.
type ApplyResult struct {
	Collections        []PlanCollectionResult
	TotalBound         int
	RetiredCollections []string // old dt_* names to remove manually in PocketBase after rename
}

// ResolveSourceRoot cleans and absolutizes a legacy markdown tree path.
func ResolveSourceRoot(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("documents: source path required")
	}
	clean := filepath.Clean(path)
	abs, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("documents: resolve source: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("documents: not a directory: %s", abs)
	}
	return abs, nil
}

// PlanFromSource inventories a legacy tree and proposes bind plans.
func PlanFromSource(sourceRoot string) ([]BindPlan, ScanSummary, error) {
	abs, err := ResolveSourceRoot(sourceRoot)
	if err != nil {
		return nil, ScanSummary{}, err
	}

	entries, err := ScanTree(abs)
	if err != nil {
		return nil, ScanSummary{}, err
	}

	plans := DefaultPlanFromScan(abs, entries)
	summary := ScanSummary{
		SourceRoot: abs,
		FileCount:  len(entries),
	}
	for i, e := range entries {
		if i >= scanPreviewMaxPaths {
			break
		}
		summary.SamplePaths = append(summary.SamplePaths, e.RelPath)
	}
	return plans, summary, nil
}

// MarshalPlansYAML encodes bind plans for CLI or dashboard forms.
func MarshalPlansYAML(plans []BindPlan) (string, error) {
	data, err := yaml.Marshal(plans)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParsePlansYAML decodes bind plans from YAML text.
func ParsePlansYAML(text string) ([]BindPlan, error) {
	var plans []BindPlan
	if err := yaml.Unmarshal([]byte(text), &plans); err != nil {
		return nil, fmt.Errorf("documents: invalid plan YAML: %w", err)
	}
	if len(plans) == 0 {
		return nil, fmt.Errorf("documents: plan YAML is empty")
	}
	return plans, nil
}

// ApplyPlans copies legacy markdown into the site per bind plans.
func ApplyPlans(siteAbs string, plans []BindPlan, opts ApplyOptions) (ApplyResult, error) {
	var result ApplyResult

	for _, plan := range plans {
		col := PlanCollectionResult{
			Collection: plan.Collection,
			DestRoot:   plan.DestRoot,
		}

		if opts.WriteSchema && !opts.DryRun {
			schemaPath := filepath.Join(siteAbs, "schema", plan.Collection+".json")
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				schemaJSON := DefaultArticlesSchema(plan.Collection, plan.DestRoot, plan.FieldMap)
				if err := os.WriteFile(schemaPath, []byte(schemaJSON), 0o644); err != nil {
					return result, err
				}
				col.SchemaWritten = true
			}
		}

		n, err := ApplyBind(siteAbs, plan, opts.DryRun, opts.Force)
		if err != nil {
			return result, err
		}
		col.Bound = n
		result.Collections = append(result.Collections, col)
		result.TotalBound += n
	}
	return result, nil
}

// SyncAfterBind upserts markdown files into PocketBase after a successful bind.
func SyncAfterBind(app core.App, siteAbs string) (SyncResult, error) {
	return SyncAll(app, siteAbs)
}
