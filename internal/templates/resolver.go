package templates

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotFound is returned when no template matches the slug.
var ErrNotFound = errors.New("template not found")

// ResolveSlug maps a URL slug to a workspace template file path using mirror+index rules.
func ResolveSlug(wsAbs, slug string) (pagePath string, err error) {
	slug = strings.TrimSuffix(slug, "/")

	templatesDir := filepath.Join(wsAbs, "templates")

	if slug == "" {
		pagePath = filepath.Join(templatesDir, "index.gohtml")
		if _, err := os.Stat(pagePath); err != nil {
			if os.IsNotExist(err) {
				return "", ErrNotFound
			}
			return "", err
		}
		return pagePath, nil
	}

	flatPath := filepath.Join(templatesDir, slug+".gohtml")
	if _, err := os.Stat(flatPath); err == nil {
		return flatPath, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	indexPath := filepath.Join(templatesDir, slug, "index.gohtml")
	if _, err := os.Stat(indexPath); err == nil {
		return indexPath, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	return "", ErrNotFound
}
