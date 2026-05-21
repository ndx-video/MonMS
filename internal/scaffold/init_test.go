package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestInitScaffold(t *testing.T) {
	ws := testutil.NewWorkspace(t)

	for _, rel := range []string{"schema", "templates/fragments", "assets"} {
		path := filepath.Join(ws, rel)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", rel, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", rel)
		}
	}

	if _, err := os.Stat(filepath.Join(ws, "assets/main.css")); err != nil {
		t.Fatalf("assets/main.css missing: %v", err)
	}
}

func TestInitGit(t *testing.T) {
	t.Skip("implemented in plan 01-05")
}

func TestBaseLayoutCDN(t *testing.T) {
	t.Skip("implemented in plan 01-05")
}
