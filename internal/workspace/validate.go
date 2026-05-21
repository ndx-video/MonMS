package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateWorkspace checks that a workspace has the minimum structure required
// before serve. Caller should print the error to stderr and exit 1 (D-06).
func ValidateWorkspace(ws string) error {
	checks := []string{
		filepath.Join(ws, "templates/layouts/base.gohtml"),
		filepath.Join(ws, "assets"),
	}
	for _, p := range checks {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("workspace incomplete: missing %s\nRun: monms init", p)
		}
	}
	return nil
}
