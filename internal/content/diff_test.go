package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monms/monms/internal/testutil"
)

func TestChecksumStable(t *testing.T) {
	files := []CollectionFile{{
		Collection: "hero_content",
		Records: []map[string]any{
			{"id": "homepage", "title": "A", "body": "B"},
		},
	}}
	c1, err := ChecksumExport(files)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	c2, err := ChecksumExport(files)
	if err != nil {
		t.Fatalf("checksum2: %v", err)
	}
	if c1 != c2 || !strings.HasPrefix(c1, "sha256:") {
		t.Fatalf("checksum %q, want stable sha256: prefix", c1)
	}
}

func TestDiffExportDetectsTitleChange(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	app := bootstrapEditorialApp(t, ws)

	if err := ExportAll(app, ws); err != nil {
		t.Fatalf("export: %v", err)
	}
	snap, err := ExportSnapshot(app, ws)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	sum, err := ChecksumExport(snap)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if err := WritePublishState(ws, PublishState{
		Checksum:    sum,
		PublishedAt: time.Now().UTC().Format(time.RFC3339),
		Collections: []string{"hero_content"},
	}); err != nil {
		t.Fatalf("write state: %v", err)
	}

	rec, err := app.FindRecordById("hero_content", "homepage")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	rec.Set("title", "Changed After Publish")
	if err := app.Save(rec); err != nil {
		t.Fatalf("save: %v", err)
	}

	diff, err := DiffExport(app, ws)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if !diff.HasChanges {
		t.Fatal("HasChanges = false, want true")
	}
	found := false
	for _, c := range diff.Changes {
		if strings.Contains(c, "title") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("changes %v, want title mentioned", diff.Changes)
	}
}

func TestDiffExportNoChangesWhenChecksumMatches(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	app := bootstrapEditorialApp(t, ws)

	snap, err := ExportSnapshot(app, ws)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	sum, err := ChecksumExport(snap)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if err := ExportAll(app, ws); err != nil {
		t.Fatalf("export: %v", err)
	}
	if err := WritePublishState(ws, PublishState{Checksum: sum}); err != nil {
		t.Fatalf("write state: %v", err)
	}

	diff, err := DiffExport(app, ws)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if diff.HasChanges {
		t.Fatalf("HasChanges = true, want false; changes=%v", diff.Changes)
	}
}

func TestDiffSnapshotsDetectsDeletedRecord(t *testing.T) {
	baseline := []CollectionFile{{
		Collection: "hero_content",
		Records: []map[string]any{
			{"id": "homepage", "title": "A"},
			{"id": "removed", "title": "Gone"},
		},
	}}
	current := []CollectionFile{{
		Collection: "hero_content",
		Records: []map[string]any{
			{"id": "homepage", "title": "A"},
		},
	}}

	changes := diffSnapshots(baseline, current)
	found := false
	for _, c := range changes {
		if c == "hero_content/removed: record deleted" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("changes %v, want deleted record", changes)
	}
}

func TestWritePublishStateUnderMonms(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	if err := WritePublishState(ws, PublishState{
		Checksum:    "sha256:abc",
		PublishedAt: "2026-05-23T12:00:00Z",
		Collections: []string{"hero_content"},
	}); err != nil {
		t.Fatalf("write: %v", err)
	}

	path := filepath.Join(ws, ".monms", "publish-state.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat publish-state: %v", err)
	}
	state, err := ReadPublishState(ws)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if state.Checksum != "sha256:abc" {
		t.Fatalf("checksum %q", state.Checksum)
	}
}

func TestReadPublishStateMissingReturnsZero(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	state, err := ReadPublishState(ws)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if state.Checksum != "" {
		t.Fatalf("checksum %q, want empty", state.Checksum)
	}
}
