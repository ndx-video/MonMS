package content

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ensureUnderWorkspace rejects paths outside wsRoot (T-04-03).
func ensureUnderWorkspace(wsRoot, dest string) error {
	wsRoot = filepath.Clean(wsRoot)
	dest = filepath.Clean(dest)
	rel, err := filepath.Rel(wsRoot, dest)
	if err != nil {
		return fmt.Errorf("resolve path under workspace: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("refusing path outside workspace: %s", dest)
	}
	return nil
}
