package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SchemaMeta holds MonMS-specific metadata from raw schema JSON files.
// PocketBase ImportCollections strips unknown keys like editorial (D-54).
type SchemaMeta struct {
	Name      string `json:"name"`
	Editorial bool   `json:"editorial"`
}

// LoadEditorialCollectionNames scans dir for *.json schema files and returns
// collection names with editorial: true. Missing dir returns nil, nil.
func LoadEditorialCollectionNames(dir string) ([]string, error) {
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

	var names []string
	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		data = trimJSON(data)
		if len(data) == 0 || data[0] != '{' {
			continue
		}

		var meta SchemaMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		if meta.Name != "" && meta.Editorial {
			names = append(names, meta.Name)
		}
	}

	return names, nil
}
