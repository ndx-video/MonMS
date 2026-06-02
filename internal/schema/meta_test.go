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
	writeSchemaFile(t, dir, "articles.json", `{
  "name": "articles",
  "type": "base",
  "editorial": true,
  "monms": { "source": "markdown", "root": "documents/articles" }
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
	if len(names) != 2 {
		t.Fatalf("LoadEditorialCollectionNames() = %v, want 2 editorial collections", names)
	}

	pbNames, err := LoadPBNativeEditorialCollectionNames(dir)
	if err != nil {
		t.Fatalf("LoadPBNativeEditorialCollectionNames() error = %v", err)
	}
	if len(pbNames) != 1 || pbNames[0] != "hero_content" {
		t.Fatalf("LoadPBNativeEditorialCollectionNames() = %v, want [hero_content]", pbNames)
	}

	md, err := LoadMarkdownBindings(dir)
	if err != nil {
		t.Fatalf("LoadMarkdownBindings() error = %v", err)
	}
	if len(md) != 1 || md[0].Name != "articles" {
		t.Fatalf("LoadMarkdownBindings() = %+v, want articles", md)
	}
	if md[0].Monms.Root != "documents/articles" {
		t.Fatalf("articles root = %q", md[0].Monms.Root)
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

func TestCollectionMetaSourceDefaults(t *testing.T) {
	meta := CollectionMeta{Editorial: true}
	if meta.Source() != SourcePocketBase {
		t.Fatalf("default source = %q, want pocketbase", meta.Source())
	}
	if !meta.IsPBNative() {
		t.Fatal("expected PB-native by default")
	}
}

func writeSchemaFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write schema file %s: %v", path, err)
	}
}
