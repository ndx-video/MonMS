package site

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateSite checks that a site has the minimum structure required
// before serve. Caller should print the error to stderr and exit 1 (D-06).
func ValidateSite(ws string) error {
	checks := []string{
		filepath.Join(ws, "templates/layouts/base.gohtml"),
		filepath.Join(ws, "assets"),
	}
	for _, p := range checks {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("site incomplete: missing %s\nRun: monms init", p)
		}
	}
	return nil
}
