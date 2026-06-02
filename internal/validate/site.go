package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateSiteTemplates runs template and HTML validation on all .gohtml files under site templates/.
func ValidateSiteTemplates(siteAbs string) error {
	root := filepath.Join(siteAbs, "templates")
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".gohtml") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("walk templates: %w", err)
	}
	if len(files) == 0 {
		return nil
	}
	return ValidateFiles(siteAbs, files)
}
