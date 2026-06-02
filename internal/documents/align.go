package documents

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monms/monms/internal/schema"
)

// AlignmentIssueKind classifies a doctree binding mismatch.
type AlignmentIssueKind string

const (
	AlignNewBinding       AlignmentIssueKind = "new_binding"
	AlignCollectionRename AlignmentIssueKind = "collection_rename"
	AlignRootDrift        AlignmentIssueKind = "root_drift"
	AlignStaleRoot        AlignmentIssueKind = "stale_root"
	AlignOrphanSchema     AlignmentIssueKind = "orphan_schema"
)

// DoctreeAlignmentIssue is one actionable alignment problem for operators.
type DoctreeAlignmentIssue struct {
	Kind               AlignmentIssueKind
	Stub               string
	DestRoot           string
	ExpectedCollection string
	ActualCollection   string
	SchemaFile         string
	Message            string
}

// ListDoctreeStubs returns sorted stub names with doctree content or bindings.
func ListDoctreeStubs(siteAbs string) ([]string, error) {
	seen := make(map[string]struct{})

	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return nil, err
	}
	for _, b := range bindings {
		if !schema.IsDoctreeCollection(b.Name) {
			continue
		}
		stub := b.Monms.Doctree
		if stub == "" {
			stub = stubFromDestRoot(b.Monms.Root)
		}
		if stub != "" {
			seen[stub] = struct{}{}
		}
	}

	entries, err := os.ReadDir(siteAbs)
	if err != nil {
		return nil, err
	}
	for _, ent := range entries {
		if !ent.IsDir() || strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		if ent.Name() == "schema" || ent.Name() == "templates" || ent.Name() == "documents" {
			continue
		}
		if err := ValidateDoctreeStub(ent.Name()); err != nil {
			continue
		}
		stubRoot := filepath.Join(siteAbs, ent.Name())
		leaves, err := leafMarkdownDirs(stubRoot)
		if err != nil || len(leaves) == 0 {
			continue
		}
		seen[ent.Name()] = struct{}{}
	}

	stubs := make([]string, 0, len(seen))
	for s := range seen {
		stubs = append(stubs, s)
	}
	sort.Strings(stubs)
	return stubs, nil
}

// AuditDoctreeAlignment compares filesystem discovery to schema for all stubs.
func AuditDoctreeAlignment(siteAbs string) ([]DoctreeAlignmentIssue, error) {
	stubs, err := ListDoctreeStubs(siteAbs)
	if err != nil {
		return nil, err
	}
	var issues []DoctreeAlignmentIssue
	for _, stub := range stubs {
		stubIssues, err := auditStubAlignment(siteAbs, stub)
		if err != nil {
			return nil, err
		}
		issues = append(issues, stubIssues...)
	}
	sort.Slice(issues, func(i, j int) bool {
		a, b := issues[i], issues[j]
		if a.Stub != b.Stub {
			return a.Stub < b.Stub
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.DestRoot < b.DestRoot
	})
	return issues, nil
}

func auditStubAlignment(siteAbs, stub string) ([]DoctreeAlignmentIssue, error) {
	candidates, err := DiscoverLeafCollections(siteAbs, stub)
	if err != nil {
		candidates = nil
	}

	bindings, err := dtBindingsForStub(siteAbs, stub)
	if err != nil {
		return nil, err
	}

	byRoot := make(map[string]schema.CollectionMeta, len(bindings))
	for _, b := range bindings {
		byRoot[b.Monms.Root] = b
	}

	activeRoots := make(map[string]bool, len(candidates))
	var issues []DoctreeAlignmentIssue

	for _, c := range candidates {
		activeRoots[c.DestRoot] = true
		expected := c.Collection
		b, ok := byRoot[c.DestRoot]
		if !ok {
			issues = append(issues, DoctreeAlignmentIssue{
				Kind:               AlignNewBinding,
				Stub:               stub,
				DestRoot:           c.DestRoot,
				ExpectedCollection: expected,
				Message:            "Re-scan and confirm bindings to create schema and sync.",
			})
			continue
		}
		if b.Name != expected {
			issues = append(issues, DoctreeAlignmentIssue{
				Kind:               AlignCollectionRename,
				Stub:               stub,
				DestRoot:           c.DestRoot,
				ExpectedCollection: expected,
				ActualCollection:   b.Name,
				SchemaFile:         b.Name + ".json",
				Message:            "Re-scan and confirm bindings to rename schema (retire old PocketBase collection manually).",
			})
			continue
		}
		if b.Monms.Root != c.DestRoot {
			issues = append(issues, DoctreeAlignmentIssue{
				Kind:               AlignRootDrift,
				Stub:               stub,
				DestRoot:           c.DestRoot,
				ExpectedCollection: expected,
				ActualCollection:   b.Name,
				SchemaFile:         b.Name + ".json",
				Message:            "Re-scan and confirm bindings to update monms.root in schema.",
			})
		}
	}

	for root, b := range byRoot {
		rootPath := filepath.Join(siteAbs, filepath.FromSlash(root))
		_, statErr := os.Stat(rootPath)
		if statErr != nil && os.IsNotExist(statErr) {
			issues = append(issues, DoctreeAlignmentIssue{
				Kind:             AlignOrphanSchema,
				Stub:             stub,
				DestRoot:         root,
				ActualCollection: b.Name,
				SchemaFile:       b.Name + ".json",
				Message:          "Schema binding points at missing directory; remove schema and PocketBase collection manually.",
			})
			continue
		}
		if !activeRoots[root] {
			issues = append(issues, DoctreeAlignmentIssue{
				Kind:             AlignStaleRoot,
				Stub:             stub,
				DestRoot:         root,
				ActualCollection: b.Name,
				SchemaFile:       b.Name + ".json",
				Message:          "Leaf folder removed or moved; retire schema and PocketBase collection manually.",
			})
		}
	}

	return issues, nil
}

func dtBindingsForStub(siteAbs, stub string) ([]schema.CollectionMeta, error) {
	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return nil, err
	}
	prefix := stub + "/"
	var out []schema.CollectionMeta
	for _, b := range bindings {
		if !schema.IsDoctreeCollection(b.Name) {
			continue
		}
		root := b.Monms.Root
		dt := b.Monms.Doctree
		if dt == "" {
			dt = stubFromDestRoot(root)
		}
		if root == stub || strings.HasPrefix(root, prefix) || dt == stub {
			out = append(out, b)
		}
	}
	return out, nil
}
