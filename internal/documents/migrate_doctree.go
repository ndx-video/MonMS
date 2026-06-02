package documents

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/monms/monms/internal/schema"
)

const (
	pendingMigrateRel     = ".monms/doctree-migrate-pending.json"
	samplePathsPerBinding = 5
)

var doctreeStubPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// BindingCandidate is a proposed markdown collection for a leaf folder under a doctree stub.
type BindingCandidate struct {
	Collection  string
	DestRoot    string
	DoctreeID   string
	FileCount   int
	SamplePaths []string
	Status      string // "new" or "existing"
}

// CopyTreeResult summarizes a source tree copy into the site.
type CopyTreeResult struct {
	Stub        string
	SourceRoot  string
	FilesCopied int
	CopiedAt    time.Time
}

// FrontmatterOptions controls idempotent frontmatter updates.
type FrontmatterOptions struct {
	Force bool
}

// PendingMigrate tracks an in-progress doctree migration.
type PendingMigrate struct {
	Stub        string    `json:"stub"`
	SourceRoot  string    `json:"sourceRoot,omitempty"`
	FilesCopied int       `json:"filesCopied"`
	CopiedAt    time.Time `json:"copiedAt"`
}

// FinalizeOptions controls binding confirmation.
type FinalizeOptions struct {
	Force bool
}

// ValidateDoctreeStub ensures the stub is a safe single path segment.
func ValidateDoctreeStub(stub string) error {
	stub = strings.TrimSpace(stub)
	if stub == "" {
		return fmt.Errorf("documents: doctree stub required")
	}
	if stub == "documents" || stub == "." || stub == ".." {
		return fmt.Errorf("documents: invalid doctree stub %q", stub)
	}
	if !doctreeStubPattern.MatchString(stub) {
		return fmt.Errorf("documents: stub must match %s", doctreeStubPattern.String())
	}
	return nil
}

// CopySourceTree copies all regular files from sourceAbs into siteAbs/stub.
func CopySourceTree(siteAbs, sourceAbs, stub string) (CopyTreeResult, error) {
	if err := ValidateDoctreeStub(stub); err != nil {
		return CopyTreeResult{}, err
	}
	sourceAbs, err := ResolveSourceRoot(sourceAbs)
	if err != nil {
		return CopyTreeResult{}, err
	}

	destRoot := filepath.Join(siteAbs, stub)
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return CopyTreeResult{}, err
	}

	var copied int
	err = filepath.WalkDir(sourceAbs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		rel, err := filepath.Rel(sourceAbs, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destRoot, rel)
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}
		if err := copyFile(path, destPath); err != nil {
			return err
		}
		copied++
		return nil
	})
	if err != nil {
		return CopyTreeResult{}, err
	}

	return CopyTreeResult{
		Stub:        stub,
		SourceRoot:  sourceAbs,
		FilesCopied: copied,
		CopiedAt:    time.Now().UTC(),
	}, nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// DiscoverLeafCollections finds binding candidates under siteAbs/stub (leaf-md-folder rule).
func DiscoverLeafCollections(siteAbs, stub string) ([]BindingCandidate, error) {
	if err := ValidateDoctreeStub(stub); err != nil {
		return nil, err
	}

	stubRoot := filepath.Join(siteAbs, stub)
	if _, err := os.Stat(stubRoot); err != nil {
		return nil, fmt.Errorf("documents: doctree %q not found under site", stub)
	}

	existingRoots, err := existingMarkdownRoots(siteAbs)
	if err != nil {
		return nil, err
	}

	leafDirs, err := leafMarkdownDirs(stubRoot)
	if err != nil {
		return nil, err
	}

	names := assignCollectionNames(leafDirs, stub)
	var candidates []BindingCandidate
	for i, dir := range leafDirs {
		destRoot := filepath.ToSlash(filepath.Join(stub, dir))
		if dir == "." {
			destRoot = stub
		}
		mdFiles, err := markdownFilesForBinding(siteAbs, destRoot, leafDirs, stub)
		if err != nil {
			return nil, err
		}
		if len(mdFiles) == 0 {
			continue
		}

		status := "new"
		if existingRoots[destRoot] {
			status = "existing"
		}

		c := BindingCandidate{
			Collection: names[i],
			DestRoot:   destRoot,
			DoctreeID:  stub,
			FileCount:  len(mdFiles),
			Status:     status,
		}
		for j, p := range mdFiles {
			if j >= samplePathsPerBinding {
				break
			}
			c.SamplePaths = append(c.SamplePaths, p)
		}
		candidates = append(candidates, c)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].DestRoot < candidates[j].DestRoot
	})
	return candidates, nil
}

// StaleBindingsUnderStub returns schema monms.root paths under stub not in current discovery.
func StaleBindingsUnderStub(siteAbs, stub string, current []BindingCandidate) ([]string, error) {
	existingRoots, err := existingMarkdownRoots(siteAbs)
	if err != nil {
		return nil, err
	}
	active := make(map[string]bool, len(current))
	for _, c := range current {
		active[c.DestRoot] = true
	}
	prefix := stub + "/"
	var stale []string
	for root := range existingRoots {
		if root == stub || strings.HasPrefix(root, prefix) {
			if !active[root] {
				stale = append(stale, root)
			}
		}
	}
	sort.Strings(stale)
	return stale, nil
}

func leafMarkdownDirs(stubRoot string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(stubRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") && path != stubRoot {
			return filepath.SkipDir
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		hasDirectMD := false
		for _, ent := range entries {
			if ent.IsDir() {
				continue
			}
			if strings.HasSuffix(strings.ToLower(ent.Name()), ".md") {
				hasDirectMD = true
				break
			}
		}
		if !hasDirectMD {
			return nil
		}
		rel, err := filepath.Rel(stubRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			rel = "."
		} else {
			rel = filepath.ToSlash(rel)
		}
		dirs = append(dirs, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(dirs, func(i, j int) bool {
		if dirs[i] == "." {
			return true
		}
		if dirs[j] == "." {
			return false
		}
		return len(dirs[i]) < len(dirs[j])
	})
	return dirs, nil
}

func assignCollectionNames(leafDirs []string, stub string) []string {
	used := make(map[string]int)
	names := make([]string, len(leafDirs))
	for i, dir := range leafDirs {
		name := DoctreeCollectionName(stub, dir)
		if n, ok := used[name]; ok {
			used[name] = n + 1
			name = fmt.Sprintf("%s_%d", name, n+1)
		} else {
			used[name] = 1
		}
		names[i] = name
	}
	return names
}

func markdownFilesForBinding(siteAbs, destRoot string, leafDirs []string, stub string) ([]string, error) {
	rootAbs := filepath.Join(siteAbs, filepath.FromSlash(destRoot))
	var paths []string
	err := filepath.WalkDir(rootAbs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != rootAbs {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		relToSite, err := filepath.Rel(siteAbs, path)
		if err != nil {
			return err
		}
		relToSite = filepath.ToSlash(relToSite)
		dirRel := filepath.ToSlash(filepath.Dir(relToSite))
		if ownedByNestedLeaf(dirRel, destRoot, leafDirs, stub) {
			return nil
		}
		relToRoot, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(relToRoot))
		return nil
	})
	return paths, err
}

func ownedByNestedLeaf(fileDirRel, destRoot string, leafDirs []string, stub string) bool {
	for _, leaf := range leafDirs {
		leafDest := stub
		if leaf != "." {
			leafDest = filepath.ToSlash(filepath.Join(stub, leaf))
		}
		if leafDest == destRoot {
			continue
		}
		if strings.HasPrefix(leafDest, destRoot+"/") && (fileDirRel == leafDest || strings.HasPrefix(fileDirRel, leafDest+"/")) {
			return true
		}
	}
	return false
}

func existingMarkdownRoots(siteAbs string) (map[string]bool, error) {
	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(bindings))
	for _, b := range bindings {
		out[b.Monms.Root] = true
	}
	return out, nil
}

// EnsureFrontmatter updates frontmatter on all markdown files owned by a binding (idempotent).
func EnsureFrontmatter(siteAbs string, candidate BindingCandidate, all []BindingCandidate, opts FrontmatterOptions) (int, error) {
	leafDirs, err := leafDirsFromCandidates(all)
	if err != nil {
		return 0, err
	}
	stub := stubFromCandidates(all)
	rootAbs := filepath.Join(siteAbs, filepath.FromSlash(candidate.DestRoot))
	mdFiles, err := markdownFilesForBinding(siteAbs, candidate.DestRoot, leafDirs, stub)
	if err != nil {
		return 0, err
	}
	var updated int
	for _, rel := range mdFiles {
		path := filepath.Join(rootAbs, filepath.FromSlash(rel))
		n, err := ensureFrontmatterFile(path, candidate.Collection, rel, opts.Force)
		if err != nil {
			return updated, err
		}
		updated += n
	}
	return updated, nil
}

func stubFromDestRoot(destRoot string) string {
	parts := strings.Split(filepath.ToSlash(destRoot), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func ensureFrontmatterFile(path, collection, relToRoot string, force bool) (int, error) {
	doc, err := ParseFile(path)
	if err != nil {
		return 0, err
	}
	pathKey := strings.TrimSuffix(filepath.ToSlash(relToRoot), ".md")
	newMeta, changed := mergeDoctreeFrontmatter(doc.Meta, doc, collection, pathKey, path, force)
	if !changed && frontmatterEqual(doc.Meta, newMeta) {
		return 0, nil
	}
	if err := WriteFile(path, newMeta, doc.Body); err != nil {
		return 0, err
	}
	return 1, nil
}

func docTitle(doc ParsedDocument, path string) string {
	if t, ok := doc.Meta["title"].(string); ok && strings.TrimSpace(t) != "" {
		return t
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return strings.ReplaceAll(base, "-", " ")
}

func mergeFrontmatterMeta(meta map[string]any, collection, pathKey, title string, force bool) (map[string]any, bool) {
	targetID := recordID(collection, pathKey)
	out := make(map[string]any, len(meta)+2)
	for k, v := range meta {
		out[k] = v
	}
	changed := false

	currentID, _ := out["id"].(string)
	currentID = sanitizeRecordID(currentID)
	if force || currentID == "" || currentID != targetID {
		if currentID != targetID {
			changed = true
		}
		out["id"] = targetID
	}

	if t, ok := out["title"].(string); !ok || strings.TrimSpace(t) == "" {
		out["title"] = title
		changed = true
	}

	return out, changed
}

func frontmatterEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}
		if fmt.Sprint(va) != fmt.Sprint(vb) {
			return false
		}
	}
	return true
}

// FinalizeBindings writes schema stubs and idempotent frontmatter for all candidates.
func FinalizeBindings(siteAbs string, candidates []BindingCandidate, opts FinalizeOptions) (ApplyResult, error) {
	leafDirs, err := leafDirsFromCandidates(candidates)
	if err != nil {
		return ApplyResult{}, err
	}
	stub := stubFromCandidates(candidates)

	var result ApplyResult
	for _, c := range candidates {
		col := PlanCollectionResult{
			Collection: c.Collection,
			DestRoot:   c.DestRoot,
		}

		doctreeID := c.DoctreeID
		if doctreeID == "" {
			doctreeID = stubFromDestRoot(c.DestRoot)
		}
		renamedFrom, schemaWritten, err := ensureDoctreeSchema(siteAbs, c, doctreeID)
		if err != nil {
			return result, err
		}
		if schemaWritten {
			col.SchemaWritten = true
		}
		if renamedFrom != "" {
			result.RetiredCollections = append(result.RetiredCollections, renamedFrom)
		}

		forceFM := opts.Force || renamedFrom != ""

		mdFiles, err := markdownFilesForBinding(siteAbs, c.DestRoot, leafDirs, stub)
		if err != nil {
			return result, err
		}
		rootAbs := filepath.Join(siteAbs, filepath.FromSlash(c.DestRoot))
		for _, rel := range mdFiles {
			path := filepath.Join(rootAbs, filepath.FromSlash(rel))
			n, err := ensureFrontmatterFile(path, c.Collection, rel, forceFM)
			if err != nil {
				return result, err
			}
			col.Bound += n
		}
		result.Collections = append(result.Collections, col)
		result.TotalBound += col.Bound
	}
	return result, nil
}

func ensureDoctreeSchema(siteAbs string, c BindingCandidate, doctreeID string) (renamedFrom string, written bool, err error) {
	expectedPath := filepath.Join(siteAbs, "schema", c.Collection+".json")
	if _, err := os.Stat(expectedPath); err == nil {
		if err := patchDoctreeSchemaFile(expectedPath, c.Collection, c.DestRoot, doctreeID); err != nil {
			return "", false, err
		}
		return "", false, nil
	}

	oldName, oldPath, found, err := findDoctreeSchemaByRoot(siteAbs, c.DestRoot)
	if err != nil {
		return "", false, err
	}
	if found && oldName != c.Collection {
		if err := renameDoctreeSchemaFile(oldPath, expectedPath, c.Collection, c.DestRoot, doctreeID); err != nil {
			return "", false, err
		}
		return oldName, true, nil
	}

	schemaJSON := DefaultDoctreeCollectionSchema(c.Collection, c.DestRoot, doctreeID, DefaultDoctreeFieldMap())
	if err := os.WriteFile(expectedPath, []byte(schemaJSON), 0o644); err != nil {
		return "", false, err
	}
	return "", true, nil
}

func findDoctreeSchemaByRoot(siteAbs, destRoot string) (name, path string, found bool, err error) {
	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return "", "", false, err
	}
	for _, b := range bindings {
		if !schema.IsDoctreeCollection(b.Name) {
			continue
		}
		if b.Monms.Root == destRoot {
			return b.Name, filepath.Join(siteAbs, "schema", b.Name+".json"), true, nil
		}
	}
	return "", "", false, nil
}

func renameDoctreeSchemaFile(oldPath, newPath, name, root, doctreeID string) error {
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return err
	}
	patched, err := patchDoctreeSchemaBytes(data, name, root, doctreeID)
	if err != nil {
		return err
	}
	if err := os.WriteFile(newPath, patched, 0o644); err != nil {
		return err
	}
	return os.Remove(oldPath)
}

func patchDoctreeSchemaFile(path, name, root, doctreeID string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	patched, err := patchDoctreeSchemaBytes(data, name, root, doctreeID)
	if err != nil {
		return err
	}
	return os.WriteFile(path, patched, 0o644)
}

func patchDoctreeSchemaBytes(data []byte, name, root, doctreeID string) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	doc["name"] = name
	monms, _ := doc["monms"].(map[string]any)
	if monms == nil {
		monms = map[string]any{}
	}
	monms["root"] = root
	monms["doctree"] = doctreeID
	monms["source"] = "markdown"
	doc["monms"] = monms
	return json.MarshalIndent(doc, "", "  ")
}

func leafDirsFromCandidates(candidates []BindingCandidate) ([]string, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	stub := stubFromCandidates(candidates)
	var dirs []string
	for _, c := range candidates {
		if c.DestRoot == stub {
			dirs = append(dirs, ".")
			continue
		}
		rel := strings.TrimPrefix(c.DestRoot, stub+"/")
		dirs = append(dirs, rel)
	}
	return dirs, nil
}

func stubFromCandidates(candidates []BindingCandidate) string {
	if len(candidates) == 0 {
		return ""
	}
	return stubFromDestRoot(candidates[0].DestRoot)
}

// PruneCopiedTree removes siteAbs/stub and all contents.
func PruneCopiedTree(siteAbs, stub string) error {
	if err := ValidateDoctreeStub(stub); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(siteAbs, stub))
}

func pendingMigratePath(siteAbs string) string {
	return filepath.Join(siteAbs, pendingMigrateRel)
}

// WritePendingMigrate stores in-progress migration state.
func WritePendingMigrate(siteAbs string, p PendingMigrate) error {
	if err := os.MkdirAll(filepath.Join(siteAbs, ".monms"), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pendingMigratePath(siteAbs), data, 0o644)
}

// ReadPendingMigrate loads pending migration state.
func ReadPendingMigrate(siteAbs string) (PendingMigrate, bool, error) {
	data, err := os.ReadFile(pendingMigratePath(siteAbs))
	if err != nil {
		if os.IsNotExist(err) {
			return PendingMigrate{}, false, nil
		}
		return PendingMigrate{}, false, err
	}
	var p PendingMigrate
	if err := json.Unmarshal(data, &p); err != nil {
		return PendingMigrate{}, false, err
	}
	return p, true, nil
}

// ClearPendingMigrate removes pending migration state.
func ClearPendingMigrate(siteAbs string) error {
	err := os.Remove(pendingMigratePath(siteAbs))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
