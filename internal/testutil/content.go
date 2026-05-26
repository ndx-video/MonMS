package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

const heroContentSchemaWithEditorial = `{
  "name": "hero_content",
  "type": "base",
  "editorial": true,
  "listRule": "",
  "viewRule": "",
  "updateRule": "@request.auth.id != ''",
  "createRule": "@request.auth.id != ''",
  "deleteRule": "@request.auth.id != ''",
  "fields": [
    {
      "name": "id",
      "type": "text",
      "required": true,
      "primaryKey": true,
      "system": true,
      "min": 1,
      "max": 50,
      "pattern": "^[a-z][a-z0-9_]*$"
    },
    {"name": "title", "type": "text"},
    {"name": "body", "type": "text"}
  ]
}`

// NewEditorialWorkspace creates a temp workspace with editorial hero_content schema
// and staging publish config for content export/import tests.
func NewEditorialWorkspace(t *testing.T) string {
	t.Helper()

	ws := NewWorkspace(t)
	WriteFile(t, filepath.Join(ws, "schema/hero_content.json"), heroContentSchemaWithEditorial)

	if err := os.MkdirAll(filepath.Join(ws, ".monms"), 0o755); err != nil {
		t.Fatalf("mkdir .monms: %v", err)
	}
	WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{"productionUrl":"http://127.0.0.1:0","publisherEmails":["publisher@test.local"]}`)

	return ws
}
