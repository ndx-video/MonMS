package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEditorialCollectionNames(t *testing.T) {
	dir := t.TempDir()

	writeSchemaFile(t, dir, "hero_content.json", `{
  "name": "hero_content",
  "type": "base",
  "editorial": true
}`)
	writeSchemaFile(t, dir, "press_releases.json", `{
  "name": "press_releases",
  "type": "base"
}`)
	writeSchemaFile(t, dir, "pages.json", `{
  "name": "pages",
  "type": "base",
  "editorial": false
}`)

	names, err := LoadEditorialCollectionNames(dir)
	if err != nil {
		t.Fatalf("LoadEditorialCollectionNames() error = %v", err)
	}
	if len(names) != 1 || names[0] != "hero_content" {
		t.Fatalf("LoadEditorialCollectionNames() = %v, want [hero_content]", names)
	}
}

func TestLoadEditorialCollectionNamesMissingDir(t *testing.T) {
	names, err := LoadEditorialCollectionNames(filepath.Join(t.TempDir(), "missing-schema"))
	if err != nil {
		t.Fatalf("LoadEditorialCollectionNames() error = %v, want nil", err)
	}
	if names != nil {
		t.Fatalf("LoadEditorialCollectionNames() = %v, want nil slice", names)
	}
}

func TestLoadEditorialCollectionNamesSkipsMalformed(t *testing.T) {
	dir := t.TempDir()

	writeSchemaFile(t, dir, "bad.json", `{not json`)
	writeSchemaFile(t, dir, "good.json", `{
  "name": "hero_content",
  "type": "base",
  "editorial": true
}`)

	names, err := LoadEditorialCollectionNames(dir)
	if err != nil {
		t.Fatalf("LoadEditorialCollectionNames() error = %v", err)
	}
	if len(names) != 1 || names[0] != "hero_content" {
		t.Fatalf("LoadEditorialCollectionNames() = %v, want [hero_content]", names)
	}
}

func writeSchemaFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write schema file %s: %v", path, err)
	}
}
