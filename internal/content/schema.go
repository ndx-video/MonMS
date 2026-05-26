package content

import (
	"path/filepath"

	"github.com/monms/monms/internal/schema"
)

// LoadEditorialCollectionNames returns editorial collection names from workspace/schema.
func LoadEditorialCollectionNames(wsAbs string) ([]string, error) {
	return schema.LoadEditorialCollectionNames(filepath.Join(wsAbs, "schema"))
}
