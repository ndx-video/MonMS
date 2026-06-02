package content

import (
	"path/filepath"

	"github.com/monms/monms/internal/schema"
)

// LoadEditorialCollectionNames returns all editorial collection names from site/schema.
func LoadEditorialCollectionNames(siteAbs string) ([]string, error) {
	return schema.LoadEditorialCollectionNames(filepath.Join(siteAbs, "schema"))
}

// LoadPBNativeEditorialCollectionNames returns editorial collections on the JSON publish rail.
func LoadPBNativeEditorialCollectionNames(siteAbs string) ([]string, error) {
	return schema.LoadPBNativeEditorialCollectionNames(filepath.Join(siteAbs, "schema"))
}

// LoadMarkdownCollectionNames returns Git markdown-backed editorial collection names.
func LoadMarkdownCollectionNames(siteAbs string) ([]string, error) {
	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return nil, err
	}
	names := make([]string, len(bindings))
	for i, b := range bindings {
		names[i] = b.Name
	}
	return names, nil
}
