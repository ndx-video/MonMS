package schema

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monms/monms/internal/logging"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterBootstrapHook syncs site/schema/*.json into PocketBase after bootstrap.
func RegisterBootstrapHook(app core.App, siteAbs string) {
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		raw, err := loadSchemaJSONFiles(filepath.Join(siteAbs, "schema"))
		if err != nil {
			return err
		}
		if len(raw) > 0 {
			if err := e.App.ImportCollectionsByMarshaledJSON(raw, false); err != nil {
				return err
			}
			logging.Schema("schema sync: imported collections from site/schema")
		}

		if err := seedHeroHomepage(e.App); err != nil {
			slog.Warn("schema seed: hero homepage failed", "error", err)
		}
		return nil
	})
}

func loadSchemaJSONFiles(dir string) ([]byte, error) {
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

	var collections []json.RawMessage

	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			slog.Error("schema sync: read file failed", "path", path, "error", err)
			continue
		}

		data = trimJSON(data)
		if len(data) == 0 {
			continue
		}

		switch data[0] {
		case '[':
			var batch []json.RawMessage
			if err := json.Unmarshal(data, &batch); err != nil {
				slog.Error("schema sync: invalid JSON array", "path", path, "error", err)
				continue
			}
			collections = append(collections, batch...)
		case '{':
			var single json.RawMessage
			if err := json.Unmarshal(data, &single); err != nil {
				slog.Error("schema sync: invalid JSON object", "path", path, "error", err)
				continue
			}
			collections = append(collections, single)
		default:
			slog.Error("schema sync: invalid JSON document", "path", path)
		}
	}

	if len(collections) == 0 {
		return nil, nil
	}

	merged, err := json.Marshal(collections)
	if err != nil {
		return nil, fmt.Errorf("schema sync: marshal collections: %w", err)
	}

	return merged, nil
}

func trimJSON(data []byte) []byte {
	return []byte(strings.TrimSpace(string(data)))
}
