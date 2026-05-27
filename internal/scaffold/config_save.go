package scaffold

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ServeBindConfig is the site listen address written to .monms/config.json.
type ServeBindConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// SaveMonmsServeSettings updates bind and allowedHosts in site/.monms/config.json,
// preserving other keys such as _comment and _fieldDocs.
func SaveMonmsServeSettings(siteAbs string, bind ServeBindConfig, allowedHosts []string) error {
	path := filepath.Join(siteAbs, ".monms", "config.json")
	if err := ensureUnderSite(siteAbs, path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("monms config: create dir: %w", err)
	}

	doc := map[string]json.RawMessage{}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("monms config: read: %w", err)
	}
	if err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("monms config: parse: %w", err)
		}
	}

	bindJSON, err := json.Marshal(bind)
	if err != nil {
		return fmt.Errorf("monms config: encode bind: %w", err)
	}
	hostsJSON, err := json.Marshal(allowedHosts)
	if err != nil {
		return fmt.Errorf("monms config: encode allowedHosts: %w", err)
	}
	doc["bind"] = bindJSON
	doc["allowedHosts"] = hostsJSON

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("monms config: encode: %w", err)
	}
	out = append(out, '\n')
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("monms config: write: %w", err)
	}
	return nil
}
