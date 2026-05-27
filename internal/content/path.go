package content

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ensureUnderSite rejects paths outside siteRoot (T-04-03).
func ensureUnderSite(siteRoot, dest string) error {
	siteRoot = filepath.Clean(siteRoot)
	dest = filepath.Clean(dest)
	rel, err := filepath.Rel(siteRoot, dest)
	if err != nil {
		return fmt.Errorf("resolve path under site: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("refusing path outside site: %s", dest)
	}
	return nil
}
