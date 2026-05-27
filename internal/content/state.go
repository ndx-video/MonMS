package content

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PublishState tracks last successful publish (PUB-08).
type PublishState struct {
	Checksum      string   `json:"checksum"`
	PublishedAt   string   `json:"publishedAt"`
	Collections   []string `json:"collections"`
}

// publishStatePath returns site/.monms/publish-state.json.
func publishStatePath(siteAbs string) string {
	return filepath.Join(siteAbs, ".monms", "publish-state.json")
}

// ReadPublishState loads publish state; missing file returns zero state (PUB-08).
func ReadPublishState(siteAbs string) (PublishState, error) {
	path := publishStatePath(siteAbs)
	if err := ensureUnderSite(siteAbs, path); err != nil {
		return PublishState{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PublishState{}, nil
		}
		return PublishState{}, fmt.Errorf("content state: read: %w", err)
	}
	if len(data) == 0 {
		return PublishState{}, nil
	}

	var state PublishState
	if err := json.Unmarshal(data, &state); err != nil {
		return PublishState{}, fmt.Errorf("content state: parse: %w", err)
	}
	return state, nil
}

// WritePublishState writes publish state under site/.monms/ only (PUB-08).
func WritePublishState(siteAbs string, state PublishState) error {
	path := publishStatePath(siteAbs)
	if err := ensureUnderSite(siteAbs, path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("content state: mkdir .monms: %w", err)
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("content state: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("content state: write: %w", err)
	}
	return nil
}
