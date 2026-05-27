package content

import (
	"path/filepath"

	"github.com/monms/monms/internal/schema"
)

// LoadEditorialCollectionNames returns editorial collection names from site/schema.
func LoadEditorialCollectionNames(siteAbs string) ([]string, error) {
	return schema.LoadEditorialCollectionNames(filepath.Join(siteAbs, "schema"))
}
