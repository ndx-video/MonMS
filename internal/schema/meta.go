package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// SourcePocketBase is the default editorial source (inline HTMX + JSON publish rail).
	SourcePocketBase = "pocketbase"
	// SourceMarkdown marks Git-tracked documents/ as canonical; PB is a derived index.
	SourceMarkdown = "markdown"
)

// MonmsBinding holds MonMS-specific collection binding from raw schema JSON.
// PocketBase ImportCollections strips the monms key (D-54).
type MonmsBinding struct {
	Source   string            `json:"source"`
	Root     string            `json:"root"`
	SlugFrom string            `json:"slugFrom"`
	IDFrom   string            `json:"idFrom"`
	Fields   map[string]string `json:"fields"`
	Body     string            `json:"body"`
}

// CollectionMeta is MonMS metadata for one schema collection file.
type CollectionMeta struct {
	Name      string
	Editorial bool
	Monms     MonmsBinding
}

// Source returns the effective content source for an editorial collection.
func (m CollectionMeta) Source() string {
	if m.Monms.Source == "" {
		return SourcePocketBase
	}
	return m.Monms.Source
}

// IsMarkdownSource reports whether the collection is Git markdown-backed.
func (m CollectionMeta) IsMarkdownSource() bool {
	return m.Editorial && m.Source() == SourceMarkdown
}

// IsPBNative reports whether the collection uses PocketBase as canonical store.
func (m CollectionMeta) IsPBNative() bool {
	return m.Editorial && !m.IsMarkdownSource()
}

// schemaFileMeta is the JSON shape read from schema/*.json files.
type schemaFileMeta struct {
	Name      string        `json:"name"`
	Editorial bool          `json:"editorial"`
	Monms     *MonmsBinding `json:"monms"`
}

// LoadCollectionMeta scans dir for *.json schema files and returns editorial metadata.
func LoadCollectionMeta(dir string) ([]CollectionMeta, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	var out []CollectionMeta
	for _, name := range files {
		meta, ok := readSchemaFileMeta(filepath.Join(dir, name))
		if !ok || meta.Name == "" || !meta.Editorial {
			continue
		}
		out = append(out, meta)
	}
	return out, nil
}

// LoadEditorialCollectionNames scans dir for *.json schema files and returns
// collection names with editorial: true. Missing dir returns nil, nil.
func LoadEditorialCollectionNames(dir string) ([]string, error) {
	meta, err := LoadCollectionMeta(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, m := range meta {
		names = append(names, m.Name)
	}
	return names, nil
}

// LoadPBNativeEditorialCollectionNames returns editorial collections on the JSON publish rail.
func LoadPBNativeEditorialCollectionNames(dir string) ([]string, error) {
	meta, err := LoadCollectionMeta(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, m := range meta {
		if m.IsPBNative() {
			names = append(names, m.Name)
		}
	}
	return names, nil
}

// LoadMarkdownBindings returns markdown-backed editorial collections with bindings.
func LoadMarkdownBindings(dir string) ([]CollectionMeta, error) {
	meta, err := LoadCollectionMeta(dir)
	if err != nil {
		return nil, err
	}
	var out []CollectionMeta
	for _, m := range meta {
		if m.IsMarkdownSource() {
			out = append(out, m)
		}
	}
	return out, nil
}

func readSchemaFileMeta(path string) (CollectionMeta, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return CollectionMeta{}, false
	}

	data = trimJSON(data)
	if len(data) == 0 || data[0] != '{' {
		return CollectionMeta{}, false
	}

	var raw schemaFileMeta
	if err := json.Unmarshal(data, &raw); err != nil {
		return CollectionMeta{}, false
	}

	meta := CollectionMeta{
		Name:      raw.Name,
		Editorial: raw.Editorial,
	}
	if raw.Monms != nil {
		meta.Monms = *raw.Monms
	}
	return meta, true
}
