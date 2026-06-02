package documents

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/testutil"
	"github.com/pocketbase/pocketbase/core"
)

func TestDiscoverUsesDtPrefix(t *testing.T) {
	site := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(site, "guide", "page.md"), "# Hi\n")
	candidates, err := DiscoverLeafCollections(site, "guide")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("candidates = %+v", candidates)
	}
	if candidates[0].Collection != "dt_guide" {
		t.Fatalf("collection = %q, want dt_guide", candidates[0].Collection)
	}
	if candidates[0].DoctreeID != "guide" {
		t.Fatalf("doctree_id = %q, want guide", candidates[0].DoctreeID)
	}
}

func TestRecordFromDocumentDoctreeFields(t *testing.T) {
	site := testutil.NewSite(t)
	root := filepath.Join(site, "guide")
	testutil.WriteFile(t, filepath.Join(root, "a.md"), "# A\nBody\n")

	binding := schema.CollectionMeta{
		Name: "dt_guide",
		Monms: schema.MonmsBinding{
			Source:  schema.SourceMarkdown,
			Root:    "guide",
			Doctree: "guide",
		},
	}
	doc, err := ParseFile(filepath.Join(root, "a.md"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	record, err := RecordFromDocument(binding, doc, root)
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if record["doctree_id"] != "guide" {
		t.Fatalf("doctree_id = %v", record["doctree_id"])
	}
	if record["leaf_path"] != "a" {
		t.Fatalf("leaf_path = %v", record["leaf_path"])
	}
}

func TestRecordFromDocumentDoctreeTitleFromFrontmatter(t *testing.T) {
	binding := schema.CollectionMeta{
		Name: "dt_guide",
		Monms: schema.MonmsBinding{
			Source:  schema.SourceMarkdown,
			Root:    "guide",
			Doctree: "guide",
			Fields:  DefaultDoctreeFieldMap(),
		},
	}
	doc := ParsedDocument{
		FilePath: filepath.Join("guide", "page.md"),
		Meta:     map[string]any{"title": "From FM"},
		Body:     "# Heading\n",
	}
	record, err := RecordFromDocument(binding, doc, "guide")
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if record["title"] != "From FM" {
		t.Fatalf("title = %v, want From FM", record["title"])
	}
}

func TestRecordFromDocumentDoctreeTitleFromH1(t *testing.T) {
	binding := schema.CollectionMeta{
		Name: "dt_guide",
		Monms: schema.MonmsBinding{
			Source:  schema.SourceMarkdown,
			Root:    "guide",
			Doctree: "guide",
		},
	}
	doc := ParsedDocument{
		FilePath: filepath.Join("guide", "intro.md"),
		Meta:     map[string]any{},
		Body:     "# My Doc\n\nBody.\n",
	}
	record, err := RecordFromDocument(binding, doc, "guide")
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if record["title"] != "My Doc" {
		t.Fatalf("title = %v, want My Doc", record["title"])
	}
}

func TestDtFileSyncCreatesRecordAndFile(t *testing.T) {
	site := testutil.NewSite(t)
	root := filepath.Join(site, "guide")
	testutil.WriteFile(t, filepath.Join(root, "new.md"), "# New\nBody\n")

	schemaJSON := DefaultDoctreeCollectionSchema("dt_guide", "guide", "guide", nil)
	testutil.WriteFile(t, filepath.Join(site, "schema", "dt_guide.json"), schemaJSON)

	app := bootstrapDtTest(t, site)

	bindings, err := schema.LoadMarkdownBindings(filepath.Join(site, "schema"))
	if err != nil {
		t.Fatalf("bindings: %v", err)
	}
	if len(bindings) != 1 {
		t.Fatalf("bindings = %d", len(bindings))
	}

	r, err := DtFileSyncCollection(app, site, bindings[0])
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if r.RecordsCreated+r.RecordsUpdated < 1 {
		t.Fatalf("records synced = %d created + %d updated, want at least 1",
			r.RecordsCreated, r.RecordsUpdated)
	}

	rec, err := app.FindRecordById("dt_guide", recordID("dt_guide", "new"))
	if err != nil {
		t.Fatalf("find record: %v", err)
	}
	if rec.GetString("doctree_id") != "guide" {
		t.Fatalf("doctree_id = %q", rec.GetString("doctree_id"))
	}
}

func TestDtFileSyncPBTitleWinsUpdatesFile(t *testing.T) {
	site := testutil.NewSite(t)
	root := filepath.Join(site, "guide")
	mdPath := filepath.Join(root, "page.md")
	testutil.WriteFile(t, mdPath, "# Old\n")

	schemaJSON := DefaultDoctreeCollectionSchema("dt_guide", "guide", "guide", nil)
	testutil.WriteFile(t, filepath.Join(site, "schema", "dt_guide.json"), schemaJSON)

	app := bootstrapDtTest(t, site)
	bindings, _ := schema.LoadMarkdownBindings(filepath.Join(site, "schema"))
	if _, err := DtFileSyncCollection(app, site, bindings[0]); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	id := recordID("dt_guide", "page")
	rec, err := app.FindRecordById("dt_guide", id)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	rec.Set("title", "From PB")
	rec.Set("ts_mod", time.Now().UTC().Format(time.RFC3339))
	if err := app.Save(rec); err != nil {
		t.Fatalf("save: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	_ = os.Chtimes(mdPath, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour))

	r, err := DtFileSyncCollection(app, site, bindings[0])
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if r.FilesUpdated < 1 {
		t.Fatalf("files updated = %d, want >= 1", r.FilesUpdated)
	}
	doc, err := ParseFile(mdPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc.Meta["title"] != "From PB" {
		t.Fatalf("title = %v", doc.Meta["title"])
	}
}

func bootstrapDtTest(t *testing.T, site string) core.App {
	t.Helper()
	app := testutil.BootstrapApp(t, site)
	if err := schema.ImportSiteCollections(app, site); err != nil {
		t.Fatalf("ImportSiteCollections: %v", err)
	}
	return app
}
