package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveMonmsServeSettings(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".monms"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ".monms", "config.json")
	if err := os.WriteFile(path, []byte(`{
  "_comment": "keep me",
  "_fieldDocs": {"bind": "docs"},
  "productionUrl": "https://example.com"
}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := SaveMonmsServeSettings(dir, ServeBindConfig{Host: "0.0.0.0", Port: "8090"}, []string{"localhost", "127.0.0.1"}); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["_comment"]; !ok {
		t.Fatal("_comment not preserved")
	}
	var bind ServeBindConfig
	if err := json.Unmarshal(doc["bind"], &bind); err != nil || bind.Host != "0.0.0.0" {
		t.Fatalf("bind = %#v, err = %v", bind, err)
	}
}
