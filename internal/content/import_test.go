package content

import (
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestImportFilesEmptyContentDir(t *testing.T) {
	ws := testutil.NewEditorialSite(t)
	app := bootstrapEditorialApp(t, ws)
	if err := ImportFiles(app, ws); err != nil {
		t.Fatalf("import empty: %v", err)
	}
}
