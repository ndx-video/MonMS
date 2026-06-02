package documents

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestAuditCollectionRename(t *testing.T) {
	site := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(site, "guide", "tutorials", "page.md"), "# Page\n")
	// Legacy basename-only schema name for same root
	schemaJSON := `{
  "name": "dt_tutorial",
  "type": "base",
  "editorial": true,
  "monms": { "source": "markdown", "root": "guide/tutorials", "doctree": "guide" },
  "fields": [{ "name": "id", "type": "text", "primaryKey": true, "required": true }]
}`
	testutil.WriteFile(t, filepath.Join(site, "schema", "dt_tutorial.json"), schemaJSON)

	issues, err := AuditDoctreeAlignment(site)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	var found bool
	for _, iss := range issues {
		if iss.Kind == AlignCollectionRename && iss.ExpectedCollection == "dt_guide_tutorials" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected collection_rename issue, got %+v", issues)
	}
}
