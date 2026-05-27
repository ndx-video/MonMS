package site

import (
	"fmt"
	"os"
	"path/filepath"
)

// Status describes site readiness before serve.
type Status int

const (
	StatusReady Status = iota
	StatusMissingRoot
	StatusIncomplete
)

// RequiredPaths returns minimum paths that must exist before serve.
func RequiredPaths(siteAbs string) []string {
	return []string{
		filepath.Join(siteAbs, "templates/layouts/base.gohtml"),
		filepath.Join(siteAbs, "assets"),
	}
}

// CheckSite classifies site readiness and lists missing required paths.
func CheckSite(siteAbs string) (Status, []string) {
	if info, err := os.Stat(siteAbs); err != nil {
		if os.IsNotExist(err) {
			return StatusMissingRoot, RequiredPaths(siteAbs)
		}
		return StatusIncomplete, []string{siteAbs}
	} else if !info.IsDir() {
		return StatusIncomplete, []string{siteAbs}
	}

	var missing []string
	for _, p := range RequiredPaths(siteAbs) {
		if _, err := os.Stat(p); err != nil {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		return StatusIncomplete, missing
	}
	return StatusReady, nil
}

// ValidateSite checks that a site has the minimum structure required
// before serve. Caller should print the error to stderr and exit 1 (D-06).
func ValidateSite(ws string) error {
	status, missing := CheckSite(ws)
	switch status {
	case StatusReady:
		return nil
	case StatusMissingRoot:
		return fmt.Errorf("site not found: %s", ws)
	default:
		if len(missing) == 1 {
			return fmt.Errorf("site incomplete: missing %s", missing[0])
		}
		return fmt.Errorf("site incomplete: missing required files under %s", ws)
	}
}
