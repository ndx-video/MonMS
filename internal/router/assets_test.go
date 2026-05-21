package router

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeAssetPath(t *testing.T) {
	root := t.TempDir()
	assetsRoot := filepath.Join(root, "assets")
	if err := os.MkdirAll(assetsRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := safeAssetPath(assetsRoot, "../../../etc/passwd")
	if !errors.Is(err, errForbidden) {
		t.Fatalf("safeAssetPath traversal: got %v, want errForbidden", err)
	}

	got, err := safeAssetPath(assetsRoot, "main.css")
	if err != nil {
		t.Fatalf("safeAssetPath main.css: %v", err)
	}
	want := filepath.Join(assetsRoot, "main.css")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
